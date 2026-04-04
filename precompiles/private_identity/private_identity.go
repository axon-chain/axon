package privateidentity

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	agentkeeper "github.com/axon-chain/axon/x/agent/keeper"
	privacykeeper "github.com/axon-chain/axon/x/privacy/keeper"
)

var (
	reputationProofVKID = sha256.Sum256([]byte("axon/circuit/reputation_proof/v1"))
	capabilityProofVKID = sha256.Sum256([]byte("axon/circuit/capability_proof/v1"))
	stakeProofVKID      = sha256.Sum256([]byte("axon/circuit/stake_proof/v1"))
)

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000812")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	RegisterIdentityCommitmentMethod = "registerIdentityCommitment"
	ProveReputationMethod            = "proveReputation"
	ProveCapabilityMethod            = "proveCapability"
	ProveStakeMethod                 = "proveStake"
	IsCommitmentRegisteredMethod     = "isCommitmentRegistered"

	GasRegisterIdentity       = 50000
	GasProveReputation        = 300000
	GasProveCapability        = 300000
	GasProveStake             = 300000
	GasIsCommitmentRegistered = 1000
)

type Precompile struct {
	cmn.Precompile
	abi         abi.ABI
	keeper      privacykeeper.Keeper
	agentKeeper agentkeeper.Keeper
}

func NewPrecompile(k privacykeeper.Keeper, ak agentkeeper.Keeper, bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPrivateIdentity ABI: %w", err)
	}
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:           storetypes.KVGasConfig(),
			TransientKVGasConfig:  storetypes.GasConfig{},
			ContractAddress:       address,
			BalanceHandlerFactory: cmn.NewBalanceHandlerFactory(bankKeeper),
		},
		abi:         parsed,
		keeper:      k,
		agentKeeper: ak,
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
	case RegisterIdentityCommitmentMethod:
		return GasRegisterIdentity
	case ProveReputationMethod:
		return GasProveReputation
	case ProveCapabilityMethod:
		return GasProveCapability
	case ProveStakeMethod:
		return GasProveStake
	case IsCommitmentRegisteredMethod:
		return GasIsCommitmentRegistered
	default:
		return 3000
	}
}

func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.execute(ctx, evm, contract, readonly)
	})
}

func (p Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case RegisterIdentityCommitmentMethod:
		return true
	default:
		return false
	}
}

func (p Precompile) execute(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.abi, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	switch method.Name {
	case RegisterIdentityCommitmentMethod:
		return p.registerIdentityCommitment(ctx, evm, contract, method, args)
	case ProveReputationMethod:
		return p.proveReputation(ctx, method, args)
	case ProveCapabilityMethod:
		return p.proveCapability(ctx, method, args)
	case ProveStakeMethod:
		return p.proveStake(ctx, method, args)
	case IsCommitmentRegisteredMethod:
		return p.isCommitmentRegistered(ctx, method, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

func (p Precompile) registerIdentityCommitment(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("registerIdentityCommitment requires 1 argument")
	}
	commitment, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("identityCommitment: expected [32]byte, got %T", args[0])
	}
	if commitment == [32]byte{} {
		return nil, fmt.Errorf("identity commitment cannot be zero")
	}

	caller := resolvePrivateIdentitySender(ctx, evm, contract, p.agentKeeper)
	cosmosAddr := sdk.AccAddress(caller.Bytes())

	agentAddr := cosmosAddr.String()
	if _, found := p.agentKeeper.GetAgent(ctx, agentAddr); !found {
		return nil, fmt.Errorf("caller is not a registered agent")
	}

	if p.keeper.HasAgentIdentity(ctx, agentAddr) {
		return nil, fmt.Errorf("agent already has a registered identity commitment")
	}

	if p.keeper.IsIdentityRegistered(ctx, commitment[:]) {
		return nil, fmt.Errorf("identity commitment already registered")
	}

	p.keeper.RegisterIdentity(ctx, commitment[:])
	if p.agentKeeper.IsV111UpgradeActivated(ctx) {
		p.keeper.SetAgentIdentity(ctx, agentAddr, commitment[:])
	} else {
		p.keeper.SetAgentIdentity(ctx, agentAddr, nil)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"private_identity_registered",
		sdk.NewAttribute("agent", cosmosAddr.String()),
	))

	return method.Outputs.Pack()
}

func resolvePrivateIdentitySender(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, agentKeeper agentkeeper.Keeper) common.Address {
	if shouldUseLegacyOriginSender(ctx, agentKeeper) {
		if evm != nil && evm.Origin != (common.Address{}) {
			return evm.Origin
		}
		if contract != nil {
			return contract.Caller()
		}
		return common.Address{}
	}

	if contract != nil {
		caller := contract.Caller()
		if caller != (common.Address{}) {
			return caller
		}
	}
	if evm != nil && evm.Origin != (common.Address{}) {
		return evm.Origin
	}
	return common.Address{}
}

func shouldUseLegacyOriginSender(ctx sdk.Context, agentKeeper agentkeeper.Keeper) bool {
	return !agentKeeper.IsV110UpgradeActivated(ctx)
}

func (p Precompile) proveReputation(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("proveReputation requires 3 arguments")
	}
	proofBytes, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("proof: expected []byte, got %T", args[0])
	}
	minReputation, ok := args[1].(uint64)
	if !ok {
		return nil, fmt.Errorf("minReputation: expected uint64, got %T", args[1])
	}
	commitment, ok := args[2].([32]byte)
	if !ok {
		return nil, fmt.Errorf("identityCommitment: expected [32]byte, got %T", args[2])
	}

	if !p.keeper.IsIdentityRegistered(ctx, commitment[:]) {
		return method.Outputs.Pack(false)
	}

	vkBytes := p.keeper.GetVK(ctx, reputationProofVKID[:])
	if vkBytes == nil {
		return method.Outputs.Pack(false)
	}

	publicInputs := []*big.Int{
		new(big.Int).SetBytes(commitment[:]),
		new(big.Int).SetUint64(minReputation),
	}

	verified, err := privacykeeper.VerifyGroth16BN254(vkBytes, proofBytes, publicInputs)
	if err != nil {
		ctx.Logger().With("module", "private_identity").Info("reputation proof failed", "error", err)
		return method.Outputs.Pack(false)
	}

	return method.Outputs.Pack(verified)
}

func (p Precompile) proveCapability(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("proveCapability requires 3 arguments")
	}
	proofBytes, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("proof: expected []byte, got %T", args[0])
	}
	capabilityHash, ok := args[1].([32]byte)
	if !ok {
		return nil, fmt.Errorf("capabilityHash: expected [32]byte, got %T", args[1])
	}
	commitment, ok := args[2].([32]byte)
	if !ok {
		return nil, fmt.Errorf("identityCommitment: expected [32]byte, got %T", args[2])
	}

	if !p.keeper.IsIdentityRegistered(ctx, commitment[:]) {
		return method.Outputs.Pack(false)
	}

	vkBytes := p.keeper.GetVK(ctx, capabilityProofVKID[:])
	if vkBytes == nil {
		return method.Outputs.Pack(false)
	}

	publicInputs := []*big.Int{
		new(big.Int).SetBytes(commitment[:]),
		new(big.Int).SetBytes(capabilityHash[:]),
	}

	verified, err := privacykeeper.VerifyGroth16BN254(vkBytes, proofBytes, publicInputs)
	if err != nil {
		ctx.Logger().With("module", "private_identity").Info("capability proof failed", "error", err)
		return method.Outputs.Pack(false)
	}

	return method.Outputs.Pack(verified)
}

func (p Precompile) proveStake(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("proveStake requires 3 arguments")
	}
	proofBytes, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("proof: expected []byte, got %T", args[0])
	}
	minStake, ok := args[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("minStake: expected *big.Int, got %T", args[1])
	}
	commitment, ok := args[2].([32]byte)
	if !ok {
		return nil, fmt.Errorf("identityCommitment: expected [32]byte, got %T", args[2])
	}

	if !p.keeper.IsIdentityRegistered(ctx, commitment[:]) {
		return method.Outputs.Pack(false)
	}

	vkBytes := p.keeper.GetVK(ctx, stakeProofVKID[:])
	if vkBytes == nil {
		return method.Outputs.Pack(false)
	}

	publicInputs := []*big.Int{
		new(big.Int).SetBytes(commitment[:]),
		minStake,
	}

	verified, err := privacykeeper.VerifyGroth16BN254(vkBytes, proofBytes, publicInputs)
	if err != nil {
		ctx.Logger().With("module", "private_identity").Info("stake proof failed", "error", err)
		return method.Outputs.Pack(false)
	}

	return method.Outputs.Pack(verified)
}

func (p Precompile) isCommitmentRegistered(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isCommitmentRegistered requires 1 argument")
	}
	commitment, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("commitment: expected [32]byte, got %T", args[0])
	}

	registered := p.keeper.IsIdentityRegistered(ctx, commitment[:])
	return method.Outputs.Pack(registered)
}

const abiJSON = `[
	{
		"inputs": [{"name": "identityCommitment", "type": "bytes32"}],
		"name": "registerIdentityCommitment",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "minReputation", "type": "uint64"},
			{"name": "identityCommitment", "type": "bytes32"}
		],
		"name": "proveReputation",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "capabilityHash", "type": "bytes32"},
			{"name": "identityCommitment", "type": "bytes32"}
		],
		"name": "proveCapability",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "minStake", "type": "uint256"},
			{"name": "identityCommitment", "type": "bytes32"}
		],
		"name": "proveStake",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "commitment", "type": "bytes32"}],
		"name": "isCommitmentRegistered",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`
