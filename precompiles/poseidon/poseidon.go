package poseidon

import (
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000810")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	Hash2Method = "hash2"
	Hash3Method = "hash3"

	GasHash2 = 10000
	GasHash3 = 15000
)

type Precompile struct {
	cmn.Precompile
	abi abi.ABI
}

func NewPrecompile(bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPoseidonHasher ABI: %w", err)
	}
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:           storetypes.KVGasConfig(),
			TransientKVGasConfig:  storetypes.GasConfig{},
			ContractAddress:       address,
			BalanceHandlerFactory: cmn.NewBalanceHandlerFactory(bankKeeper),
		},
		abi: parsed,
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
	case Hash2Method:
		return GasHash2
	case Hash3Method:
		return GasHash3
	default:
		return 3000
	}
}

func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.execute(ctx, contract)
	})
}

func (p Precompile) IsTransaction(_ *abi.Method) bool {
	return false
}

func (p Precompile) execute(ctx sdk.Context, contract *vm.Contract) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.abi, contract, true, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	switch method.Name {
	case Hash2Method:
		return p.hash2(method, args)
	case Hash3Method:
		return p.hash3(method, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

// hash2 computes MiMC(left, right) over BN254 Fr using gnark-crypto.
func (p Precompile) hash2(method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("hash2 requires 2 arguments")
	}
	left, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("left: expected [32]byte, got %T", args[0])
	}
	right, ok := args[1].([32]byte)
	if !ok {
		return nil, fmt.Errorf("right: expected [32]byte, got %T", args[1])
	}

	h := mimc.NewMiMC()
	h.Write(left[:])
	h.Write(right[:])
	var digest [32]byte
	copy(digest[:], h.Sum(nil))

	return method.Outputs.Pack(digest)
}

// hash3 computes MiMC(a, b, c) over BN254 Fr using gnark-crypto.
func (p Precompile) hash3(method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("hash3 requires 3 arguments")
	}
	a, ok := args[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("a: expected [32]byte, got %T", args[0])
	}
	b, ok := args[1].([32]byte)
	if !ok {
		return nil, fmt.Errorf("b: expected [32]byte, got %T", args[1])
	}
	c, ok := args[2].([32]byte)
	if !ok {
		return nil, fmt.Errorf("c: expected [32]byte, got %T", args[2])
	}

	h := mimc.NewMiMC()
	h.Write(a[:])
	h.Write(b[:])
	h.Write(c[:])
	var digest [32]byte
	copy(digest[:], h.Sum(nil))

	return method.Outputs.Pack(digest)
}

const abiJSON = `[
	{
		"inputs": [
			{"name": "left", "type": "bytes32"},
			{"name": "right", "type": "bytes32"}
		],
		"name": "hash2",
		"outputs": [{"name": "", "type": "bytes32"}],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "a", "type": "bytes32"},
			{"name": "b", "type": "bytes32"},
			{"name": "c", "type": "bytes32"}
		],
		"name": "hash3",
		"outputs": [{"name": "", "type": "bytes32"}],
		"stateMutability": "pure",
		"type": "function"
	}
]`
