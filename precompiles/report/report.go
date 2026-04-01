package report

import (
	"fmt"
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
	address = common.HexToAddress("0x0000000000000000000000000000000000000807")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	SubmitReportMethod          = "submitReport"
	GetContractReputationMethod = "getContractReputation"
	GetEpochReportCountMethod   = "getEpochReportCount"
	HasReportedMethod           = "hasReported"

	GasSubmitReport = 50000
	GasQuery        = 1000
)

type Precompile struct {
	cmn.Precompile
	abi    abi.ABI
	keeper keeper.Keeper
}

func NewPrecompile(k keeper.Keeper, bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IReputationReport ABI: %w", err)
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
	case SubmitReportMethod:
		return GasSubmitReport
	default:
		return GasQuery
	}
}

func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.execute(ctx, evm, contract, readonly)
	})
}

func (p Precompile) execute(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	if len(contract.Input) < 4 {
		return nil, fmt.Errorf("input too short")
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
	case SubmitReportMethod:
		if readonly {
			return nil, fmt.Errorf("submitReport is not allowed in static call")
		}
		return p.submitReport(ctx, evm, contract, method, args)
	case GetContractReputationMethod:
		return p.getContractReputation(ctx, method, args)
	case GetEpochReportCountMethod:
		return p.getEpochReportCount(ctx, method, args)
	case HasReportedMethod:
		return p.hasReported(ctx, method, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

func (p Precompile) submitReport(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("submitReport requires 4 arguments")
	}

	targetAddr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid target address")
	}
	score, ok := args[1].(int8)
	if !ok {
		return nil, fmt.Errorf("invalid score")
	}
	var evidence [32]byte
	switch v := args[2].(type) {
	case [32]byte:
		evidence = v
	case common.Hash:
		evidence = [32]byte(v)
	default:
		return nil, fmt.Errorf("invalid evidence type")
	}

	reason, _ := args[3].(string)

	caller := sdk.AccAddress(resolveReportSender(ctx, evm, contract, p.keeper).Bytes())
	target := sdk.AccAddress(targetAddr.Bytes())

	evidenceHex := ""
	if evidence != [32]byte{} {
		evidenceHex = common.Bytes2Hex(evidence[:])
	}

	if err := p.keeper.SubmitL2Report(ctx, caller.String(), target.String(), score, evidenceHex, reason); err != nil {
		return nil, err
	}

	return method.Outputs.Pack()
}

func resolveReportSender(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, agentKeeper keeper.Keeper) common.Address {
	if !agentKeeper.IsV110UpgradeActivated(ctx) {
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

func (p Precompile) getContractReputation(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getContractReputation requires 1 argument")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address")
	}
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	addrStr := cosmosAddr.String()
	l2Score := p.keeper.GetL2Score(ctx, addrStr)
	epoch := p.keeper.GetCurrentEpoch(ctx)
	posCount, negCount, uniqueCount := p.keeper.GetL2ReportStats(ctx, epoch, addrStr)

	return method.Outputs.Pack(l2Score, posCount, negCount, uniqueCount)
}

func (p Precompile) getEpochReportCount(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getEpochReportCount requires 1 argument")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address")
	}
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	epoch := p.keeper.GetCurrentEpoch(ctx)
	count := p.keeper.GetL2TargetReportCount(ctx, epoch, cosmosAddr.String())

	return method.Outputs.Pack(uint64(count))
}

func (p Precompile) hasReported(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("hasReported requires 2 arguments")
	}
	reporterAddr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid reporter address")
	}
	targetAddr, ok := args[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid target address")
	}

	reporter := sdk.AccAddress(reporterAddr.Bytes())
	target := sdk.AccAddress(targetAddr.Bytes())
	epoch := p.keeper.GetCurrentEpoch(ctx)

	has := p.keeper.HasL2Report(ctx, epoch, reporter.String(), target.String())
	return method.Outputs.Pack(has)
}

func (p Precompile) IsTransaction(method *abi.Method) bool {
	return method.Name == SubmitReportMethod
}

const abiJSON = `[
	{
		"inputs": [
			{"name": "targetAgent", "type": "address"},
			{"name": "score", "type": "int8"},
			{"name": "evidence", "type": "bytes32"},
			{"name": "reason", "type": "string"}
		],
		"name": "submitReport",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "agent", "type": "address"}],
		"name": "getContractReputation",
		"outputs": [
			{"name": "score", "type": "int64"},
			{"name": "positiveCount", "type": "uint64"},
			{"name": "negativeCount", "type": "uint64"},
			{"name": "uniqueReporters", "type": "uint64"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "agent", "type": "address"}],
		"name": "getEpochReportCount",
		"outputs": [{"name": "", "type": "uint64"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "reporter", "type": "address"},
			{"name": "target", "type": "address"}
		],
		"name": "hasReported",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	}
]`
