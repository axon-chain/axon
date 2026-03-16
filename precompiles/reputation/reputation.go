package reputation

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/keeper"
)

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000802")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	GetReputationMethod   = "getReputation"
	GetReputationsMethod  = "getReputations"
	MeetsReputationMethod = "meetsReputation"

	GasGetReputation   = 200
	GasGetReputations  = 200
	GasPerReputation   = 200
	GasMeetsReputation = 200

	MaxReputationBatch = 100
)

type Precompile struct {
	cmn.Precompile
	abi    abi.ABI
	keeper keeper.Keeper
}

func NewPrecompile(k keeper.Keeper, bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IAgentReputation ABI: %w", err)
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
	case GetReputationMethod:
		return GasGetReputation
	case GetReputationsMethod:
		// Dynamic gas: base + per-element cost
		// ABI encoding: 4 (selector) + 32 (offset) + 32 (length) + 32*N (elements)
		if len(input) >= 68 {
			arrLen := new(big.Int).SetBytes(input[36:68]).Uint64()
			if arrLen > MaxReputationBatch {
				arrLen = MaxReputationBatch
			}
			return GasGetReputations + arrLen*GasPerReputation
		}
		return GasGetReputations
	case MeetsReputationMethod:
		return GasMeetsReputation
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
	if len(contract.Input) < 4 {
		return nil, vm.ErrExecutionReverted
	}
	method, err := p.abi.MethodById(contract.Input[:4])
	if err != nil {
		return nil, err
	}
	args, err := method.Inputs.Unpack(contract.Input[4:])
	if err != nil {
		return nil, err
	}

	switch method.Name {
	case GetReputationMethod:
		return p.getReputation(ctx, method, args)
	case GetReputationsMethod:
		return p.getReputations(ctx, method, args)
	case MeetsReputationMethod:
		return p.meetsReputation(ctx, method, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

func (p Precompile) getReputation(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getReputation requires 1 argument")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address argument")
	}
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	totalMillis := p.keeper.GetTotalReputation(ctx, cosmosAddr.String())
	rep := uint64(totalMillis / 1000)
	return method.Outputs.Pack(rep)
}

func (p Precompile) getReputations(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getReputations requires 1 argument")
	}
	addrs, ok := args[0].([]common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address array argument")
	}
	if len(addrs) > MaxReputationBatch {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(addrs), MaxReputationBatch)
	}

	reps := make([]uint64, len(addrs))
	for i, addr := range addrs {
		cosmosAddr := sdk.AccAddress(addr.Bytes())
		reps[i] = uint64(p.keeper.GetTotalReputation(ctx, cosmosAddr.String()) / 1000)
	}
	return method.Outputs.Pack(reps)
}

func (p Precompile) meetsReputation(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("meetsReputation requires 2 arguments")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address argument")
	}
	minRep, ok := args[1].(uint64)
	if !ok {
		return nil, fmt.Errorf("invalid minReputation argument")
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	rep := uint64(p.keeper.GetTotalReputation(ctx, cosmosAddr.String()) / 1000)
	return method.Outputs.Pack(rep >= minRep)
}

const abiJSON = `[
	{
		"inputs": [{"name": "agent", "type": "address"}],
		"name": "getReputation",
		"outputs": [{"name": "", "type": "uint64"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "agents", "type": "address[]"}],
		"name": "getReputations",
		"outputs": [{"name": "", "type": "uint64[]"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "agent", "type": "address"},
			{"name": "minReputation", "type": "uint64"}
		],
		"name": "meetsReputation",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`
