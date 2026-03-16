package zkverifier

import (
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
	address = common.HexToAddress("0x0000000000000000000000000000000000000813")
	_       = vm.PrecompiledContract(&Precompile{})
)

// Pre-registered verifying key IDs for built-in privacy circuits.
var (
	UNSHIELD_KEY         = sha256Key("axon/circuit/unshield/v1")
	PRIVATE_TRANSFER_KEY = sha256Key("axon/circuit/private_transfer/v1")
	REPUTATION_PROOF_KEY = sha256Key("axon/circuit/reputation_proof/v1")
	CAPABILITY_PROOF_KEY = sha256Key("axon/circuit/capability_proof/v1")
	STAKE_PROOF_KEY      = sha256Key("axon/circuit/stake_proof/v1")
)

func sha256Key(label string) [32]byte {
	return sha256.Sum256([]byte(label))
}

const (
	VerifyGroth16Method        = "verifyGroth16"
	RegisterVerifyingKeyMethod = "registerVerifyingKey"
	IsKeyRegisteredMethod      = "isKeyRegistered"

	GasVerifyGroth16Base    = 200000
	GasPerPublicInput       = 10000
	GasRegisterVerifyingKey = 100000
	GasIsKeyRegistered      = 1000

	RegistrationCostAaxon = "100000000000000000000" // 100 * 10^18 aaxon = 100 AXON
)

type Precompile struct {
	cmn.Precompile
	abi    abi.ABI
	keeper privacykeeper.Keeper
}

func NewPrecompile(k privacykeeper.Keeper, bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IZKVerifier ABI: %w", err)
	}
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:           storetypes.KVGasConfig(),
			TransientKVGasConfig:  storetypes.GasConfig{},
			ContractAddress:       address,
			BalanceHandlerFactory: cmn.NewBalanceHandlerFactory(bankKeeper),
		},
		abi:    parsed,
		keeper: k,
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
	case VerifyGroth16Method:
		gas := uint64(GasVerifyGroth16Base)
		// Dynamic gas: parse public inputs array length from ABI encoding.
		// Layout: 4 (selector) + 32 (vkId) + 32 (proof offset) + 32 (inputs offset) ...
		// The public inputs array length is at the offset pointed to by the third param.
		if len(input) >= 100 {
			offsetRaw := new(big.Int).SetBytes(input[68:100])
			if offsetRaw.IsInt64() && offsetRaw.Int64() >= 0 {
				arrLenStart := 4 + int(offsetRaw.Int64())
				arrLenEnd := arrLenStart + 32
				if arrLenStart >= 0 && arrLenEnd > arrLenStart && len(input) >= arrLenEnd {
					arrLen := new(big.Int).SetBytes(input[arrLenStart:arrLenEnd]).Uint64()
					if arrLen > 256 {
						arrLen = 256
					}
					gas += arrLen * GasPerPublicInput
				}
			}
		}
		return gas
	case RegisterVerifyingKeyMethod:
		return GasRegisterVerifyingKey
	case IsKeyRegisteredMethod:
		return GasIsKeyRegistered
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
	case RegisterVerifyingKeyMethod:
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
	case VerifyGroth16Method:
		return p.verifyGroth16(ctx, method, args)
	case RegisterVerifyingKeyMethod:
		return p.registerVerifyingKey(ctx, evm, contract, method, args)
	case IsKeyRegisteredMethod:
		return p.isKeyRegistered(ctx, method, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

func (p Precompile) verifyGroth16(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("verifyGroth16 requires 3 arguments")
	}
	vkId, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("verifyingKeyId: expected [32]byte, got %T", args[0])
	}
	proofBytes, ok := args[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("proof: expected []byte, got %T", args[1])
	}
	publicInputs, ok := args[2].([]*big.Int)
	if !ok {
		return nil, fmt.Errorf("publicInputs: expected []*big.Int, got %T", args[2])
	}

	isBuiltin := vkId == UNSHIELD_KEY ||
		vkId == PRIVATE_TRANSFER_KEY ||
		vkId == REPUTATION_PROOF_KEY ||
		vkId == CAPABILITY_PROOF_KEY ||
		vkId == STAKE_PROOF_KEY

	var vkBytes []byte
	if isBuiltin {
		vkBytes = p.keeper.GetVK(ctx, vkId[:])
		if vkBytes == nil {
			return method.Outputs.Pack(false)
		}
	} else {
		if !p.keeper.IsVKRegistered(ctx, vkId[:]) {
			return method.Outputs.Pack(false)
		}
		vkBytes = p.keeper.GetVK(ctx, vkId[:])
		if vkBytes == nil {
			return method.Outputs.Pack(false)
		}
	}

	verified, err := privacykeeper.VerifyGroth16BN254(vkBytes, proofBytes, publicInputs)
	if err != nil {
		ctx.Logger().With("module", "zk_verifier").Info("groth16 verification failed", "error", err)
		return method.Outputs.Pack(false)
	}

	return method.Outputs.Pack(verified)
}

func (p Precompile) registerVerifyingKey(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("registerVerifyingKey requires 1 argument")
	}
	vkBytes, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("vk: expected []byte, got %T", args[0])
	}
	if len(vkBytes) == 0 {
		return nil, fmt.Errorf("verifying key must not be empty")
	}

	minCost, _ := sdkmath.NewIntFromString(RegistrationCostAaxon)
	msgValue := contract.Value()
	if msgValue == nil || msgValue.IsZero() {
		return nil, fmt.Errorf("must send >= 100 AXON as msg.value to register a verifying key")
	}
	sent := sdkmath.NewIntFromBigInt(msgValue.ToBig())
	if sent.LT(minCost) {
		return nil, fmt.Errorf("insufficient payment: need %s aaxon, got %s", minCost, sent)
	}

	keyId := sha256.Sum256(vkBytes)

	if p.keeper.IsVKRegistered(ctx, keyId[:]) {
		return nil, fmt.Errorf("verifying key already registered with id %x", keyId)
	}

	p.keeper.RegisterVK(ctx, keyId[:], vkBytes)

	caller := evm.TxContext.Origin
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"zk_verifying_key_registered",
		sdk.NewAttribute("key_id", fmt.Sprintf("%x", keyId)),
		sdk.NewAttribute("registrant", sdk.AccAddress(caller.Bytes()).String()),
	))

	return method.Outputs.Pack(keyId)
}

func (p Precompile) isKeyRegistered(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isKeyRegistered requires 1 argument")
	}
	keyId, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("keyId: expected [32]byte, got %T", args[0])
	}

	isBuiltin := keyId == UNSHIELD_KEY ||
		keyId == PRIVATE_TRANSFER_KEY ||
		keyId == REPUTATION_PROOF_KEY ||
		keyId == CAPABILITY_PROOF_KEY ||
		keyId == STAKE_PROOF_KEY

	registered := isBuiltin || p.keeper.IsVKRegistered(ctx, keyId[:])
	return method.Outputs.Pack(registered)
}

const abiJSON = `[
	{
		"inputs": [
			{"name": "verifyingKeyId", "type": "bytes32"},
			{"name": "proof", "type": "bytes"},
			{"name": "publicInputs", "type": "uint256[]"}
		],
		"name": "verifyGroth16",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "vk", "type": "bytes"}],
		"name": "registerVerifyingKey",
		"outputs": [{"name": "keyId", "type": "bytes32"}],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [{"name": "keyId", "type": "bytes32"}],
		"name": "isKeyRegistered",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`
