package privatetransfer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	privacykeeper "github.com/axon-chain/axon/x/privacy/keeper"
)

var (
	unshieldVKID        = sha256.Sum256([]byte("axon/circuit/unshield/v1"))
	privateTransferVKID = sha256.Sum256([]byte("axon/circuit/private_transfer/v1"))
)

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000811")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	MethodShield          = "shield"
	MethodUnshield        = "unshield"
	MethodPrivateTransfer = "privateTransfer"
	MethodIsKnownRoot     = "isKnownRoot"
	MethodIsSpent         = "isSpent"
	MethodGetTreeSize     = "getTreeSize"

	GasShield          = 100000
	GasUnshield        = 350000
	GasPrivateTransfer = 500000
	GasIsKnownRoot     = 1000
	GasIsSpent         = 1000
	GasGetTreeSize     = 500

	ModuleName = "privacy"
)

// PrivacyBankKeeper extends cmn.BankKeeper with the module-account transfer
// primitives required by the shielded pool.
type PrivacyBankKeeper interface {
	cmn.BankKeeper
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type Precompile struct {
	cmn.Precompile
	abi        abi.ABI
	keeper     privacykeeper.Keeper
	bankKeeper PrivacyBankKeeper
}

func NewPrecompile(k privacykeeper.Keeper, bankKeeper PrivacyBankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPrivateTransfer ABI: %w", err)
	}
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:           storetypes.KVGasConfig(),
			TransientKVGasConfig:  storetypes.GasConfig{},
			ContractAddress:       address,
			BalanceHandlerFactory: cmn.NewBalanceHandlerFactory(bankKeeper),
		},
		abi:        parsed,
		keeper:     k,
		bankKeeper: bankKeeper,
	}, nil
}

func (Precompile) Address() common.Address { return address }

func (p Precompile) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 3000
	}
	method, err := p.abi.MethodById(input[:4])
	if err != nil {
		return 3000
	}
	switch method.Name {
	case MethodShield:
		return GasShield
	case MethodUnshield:
		return GasUnshield
	case MethodPrivateTransfer:
		return GasPrivateTransfer
	case MethodIsKnownRoot:
		return GasIsKnownRoot
	case MethodIsSpent:
		return GasIsSpent
	case MethodGetTreeSize:
		return GasGetTreeSize
	default:
		return 3000
	}
}

func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.dispatch(ctx, evm, contract, readonly)
	})
}

func (p Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case MethodShield, MethodUnshield, MethodPrivateTransfer:
		return true
	default:
		return false
	}
}

func (p Precompile) dispatch(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.abi, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	switch method.Name {
	case MethodShield:
		return p.shield(ctx, evm, contract, method, args)
	case MethodUnshield:
		return p.unshield(ctx, method, args)
	case MethodPrivateTransfer:
		return p.privateTransfer(ctx, method, args)
	case MethodIsKnownRoot:
		return p.isKnownRoot(ctx, method, args)
	case MethodIsSpent:
		return p.isSpent(ctx, method, args)
	case MethodGetTreeSize:
		return p.getTreeSize(ctx, method)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

// ============================================================
// Write methods
// ============================================================

func (p Precompile) shield(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("shield requires 1 argument")
	}

	commitment, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid commitment argument")
	}
	if commitment == [32]byte{} {
		return nil, fmt.Errorf("commitment cannot be zero")
	}

	msgValue := contract.Value()
	if msgValue == nil || msgValue.IsZero() {
		return nil, fmt.Errorf("must send value to shield")
	}

	depositAmount := sdkmath.NewIntFromBigInt(msgValue.ToBig())

	privacyParams := p.keeper.GetParams(ctx)
	if maxStr := privacyParams.MaxShieldAmount; maxStr != "" {
		maxShield, ok := new(big.Int).SetString(maxStr, 10)
		if ok && depositAmount.BigInt().Cmp(maxShield) > 0 {
			return nil, fmt.Errorf("shield amount exceeds MaxShieldAmount (%s)", maxStr)
		}
	}
	if privacyParams.PoolCapRatio > 0 {
		currentPool := p.keeper.GetShieldedBalance(ctx)
		newPool := currentPool.Add(depositAmount)
		totalSupply := p.bankKeeper.GetSupply(ctx, "aaxon")
		if totalSupply.Amount.IsPositive() {
			capAmount := totalSupply.Amount.MulRaw(int64(privacyParams.PoolCapRatio)).QuoRaw(100)
			if capAmount.IsPositive() && newPool.GT(capAmount) {
				return nil, fmt.Errorf("shield would exceed pool cap (%d%% of total supply)", privacyParams.PoolCapRatio)
			}
		}
	}

	coins := sdk.NewCoins(sdk.NewCoin("aaxon", depositAmount))
	precompileAddr := sdk.AccAddress(address.Bytes())

	if err := p.bankKeeper.SendCoinsFromAccountToModule(ctx, precompileAddr, ModuleName, coins); err != nil {
		return nil, fmt.Errorf("failed to transfer to shielded pool: %w", err)
	}

	if err := p.keeper.AddToShieldedPool(ctx, depositAmount); err != nil {
		return nil, fmt.Errorf("failed to update shielded pool balance: %w", err)
	}

	if _, err := p.keeper.InsertCommitment(ctx, commitment[:]); err != nil {
		return nil, fmt.Errorf("failed to insert commitment: %w", err)
	}

	caller := evm.TxContext.Origin
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"shielded_deposit",
		sdk.NewAttribute("from", caller.Hex()),
		sdk.NewAttribute("commitment", common.Bytes2Hex(commitment[:])),
		sdk.NewAttribute("amount", msgValue.ToBig().String()),
	))

	return method.Outputs.Pack()
}

func (p Precompile) unshield(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 5 {
		return nil, fmt.Errorf("unshield requires 5 arguments")
	}

	proof, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid proof argument")
	}
	merkleRoot, ok := args[1].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid merkleRoot argument")
	}
	nullifier, ok := args[2].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid nullifier argument")
	}
	recipient, ok := args[3].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid recipient argument")
	}
	amountBig, ok := args[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount argument")
	}

	if recipient == (common.Address{}) {
		return nil, fmt.Errorf("recipient cannot be zero address")
	}
	if amountBig.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if !p.keeper.IsKnownRoot(ctx, merkleRoot[:]) {
		return nil, fmt.Errorf("unknown merkle root")
	}
	if p.keeper.IsNullifierSpent(ctx, nullifier[:]) {
		return nil, fmt.Errorf("nullifier already spent")
	}

	vkBytes := p.keeper.GetVK(ctx, unshieldVKID[:])
	if vkBytes == nil {
		return nil, fmt.Errorf("unshield circuit verifying key not registered")
	}
	publicInputs := []*big.Int{
		new(big.Int).SetBytes(merkleRoot[:]),
		new(big.Int).SetBytes(nullifier[:]),
		new(big.Int).SetBytes(recipient.Bytes()),
		amountBig,
	}
	verified, err := privacykeeper.VerifyGroth16BN254(vkBytes, proof, publicInputs)
	if err != nil {
		return nil, fmt.Errorf("unshield proof verification error: %w", err)
	}
	if !verified {
		return nil, fmt.Errorf("unshield proof verification failed")
	}

	p.keeper.MarkNullifierSpent(ctx, nullifier[:])

	withdrawAmount := sdkmath.NewIntFromBigInt(amountBig)
	if err := p.keeper.SubFromShieldedPool(ctx, withdrawAmount); err != nil {
		return nil, fmt.Errorf("failed to subtract from shielded pool: %w", err)
	}

	coins := sdk.NewCoins(sdk.NewCoin("aaxon", withdrawAmount))
	recipientAddr := sdk.AccAddress(recipient.Bytes())
	if err := p.bankKeeper.SendCoinsFromModuleToAccount(ctx, ModuleName, recipientAddr, coins); err != nil {
		return nil, fmt.Errorf("failed to send coins to recipient: %w", err)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"shielded_withdrawal",
		sdk.NewAttribute("nullifier", common.Bytes2Hex(nullifier[:])),
		sdk.NewAttribute("recipient", recipient.Hex()),
		sdk.NewAttribute("amount", amountBig.String()),
	))

	return method.Outputs.Pack()
}

func (p Precompile) privateTransfer(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("privateTransfer requires 4 arguments")
	}

	proof, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid proof argument")
	}
	merkleRoot, ok := args[1].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid merkleRoot argument")
	}
	inputNullifiers, ok := args[2].([2][32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid inputNullifiers argument")
	}
	outputCommitments, ok := args[3].([2][32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid outputCommitments argument")
	}

	if !p.keeper.IsKnownRoot(ctx, merkleRoot[:]) {
		return nil, fmt.Errorf("unknown merkle root")
	}
	if inputNullifiers[0] == inputNullifiers[1] {
		return nil, fmt.Errorf("input nullifiers must be different")
	}
	if p.keeper.IsNullifierSpent(ctx, inputNullifiers[0][:]) {
		return nil, fmt.Errorf("first nullifier already spent")
	}
	if p.keeper.IsNullifierSpent(ctx, inputNullifiers[1][:]) {
		return nil, fmt.Errorf("second nullifier already spent")
	}
	if outputCommitments[0] == outputCommitments[1] {
		return nil, fmt.Errorf("output commitments must be different")
	}

	vkBytes := p.keeper.GetVK(ctx, privateTransferVKID[:])
	if vkBytes == nil {
		return nil, fmt.Errorf("private transfer circuit verifying key not registered")
	}
	publicInputs := []*big.Int{
		new(big.Int).SetBytes(merkleRoot[:]),
		new(big.Int).SetBytes(inputNullifiers[0][:]),
		new(big.Int).SetBytes(inputNullifiers[1][:]),
		new(big.Int).SetBytes(outputCommitments[0][:]),
		new(big.Int).SetBytes(outputCommitments[1][:]),
	}
	verified, err := privacykeeper.VerifyGroth16BN254(vkBytes, proof, publicInputs)
	if err != nil {
		return nil, fmt.Errorf("private transfer proof verification error: %w", err)
	}
	if !verified {
		return nil, fmt.Errorf("private transfer proof verification failed")
	}

	p.keeper.MarkNullifierSpent(ctx, inputNullifiers[0][:])
	p.keeper.MarkNullifierSpent(ctx, inputNullifiers[1][:])

	if _, err := p.keeper.InsertCommitment(ctx, outputCommitments[0][:]); err != nil {
		return nil, fmt.Errorf("failed to insert first output commitment: %w", err)
	}
	if _, err := p.keeper.InsertCommitment(ctx, outputCommitments[1][:]); err != nil {
		return nil, fmt.Errorf("failed to insert second output commitment: %w", err)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"private_transfer",
		sdk.NewAttribute("nullifier_0", common.Bytes2Hex(inputNullifiers[0][:])),
		sdk.NewAttribute("nullifier_1", common.Bytes2Hex(inputNullifiers[1][:])),
		sdk.NewAttribute("commitment_0", common.Bytes2Hex(outputCommitments[0][:])),
		sdk.NewAttribute("commitment_1", common.Bytes2Hex(outputCommitments[1][:])),
	))

	return method.Outputs.Pack()
}

// ============================================================
// View methods
// ============================================================

func (p Precompile) isKnownRoot(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isKnownRoot requires 1 argument")
	}
	root, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid root argument")
	}
	return method.Outputs.Pack(p.keeper.IsKnownRoot(ctx, root[:]))
}

func (p Precompile) isSpent(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isSpent requires 1 argument")
	}
	nullifier, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid nullifier argument")
	}
	return method.Outputs.Pack(p.keeper.IsNullifierSpent(ctx, nullifier[:]))
}

func (p Precompile) getTreeSize(ctx sdk.Context, method *abi.Method) ([]byte, error) {
	size := p.keeper.GetTreeSize(ctx)
	return method.Outputs.Pack(new(big.Int).SetUint64(size))
}

// ============================================================
// ABI definition
// ============================================================

const abiJSON = `[
	{
		"inputs": [{"name": "commitment", "type": "bytes32"}],
		"name": "shield",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "merkleRoot", "type": "bytes32"},
			{"name": "nullifier", "type": "bytes32"},
			{"name": "recipient", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "unshield",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "merkleRoot", "type": "bytes32"},
			{"name": "inputNullifiers", "type": "bytes32[2]"},
			{"name": "outputCommitments", "type": "bytes32[2]"}
		],
		"name": "privateTransfer",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "root", "type": "bytes32"}],
		"name": "isKnownRoot",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "nullifier", "type": "bytes32"}],
		"name": "isSpent",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getTreeSize",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	}
]`
