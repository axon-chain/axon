package registry

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/keeper"
	"github.com/axon-chain/axon/x/agent/types"
)

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000801")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	IsAgentMethod     = "isAgent"
	GetAgentMethod    = "getAgent"
	RegisterMethod    = "register"
	UpdateAgentMethod = "updateAgent"
	HeartbeatMethod   = "heartbeat"
	DeregisterMethod  = "deregister"

	GasIsAgent    = 200
	GasGetAgent   = 1000
	GasRegister   = 50000
	GasUpdate     = 10000
	GasHeartbeat  = 5000
	GasDeregister = 20000
)

type Precompile struct {
	cmn.Precompile
	abi    abi.ABI
	keeper keeper.Keeper
}

func NewPrecompile(k keeper.Keeper, bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IAgentRegistry ABI: %w", err)
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
	case IsAgentMethod:
		return GasIsAgent
	case GetAgentMethod:
		return GasGetAgent
	case RegisterMethod:
		return GasRegister
	case UpdateAgentMethod:
		return GasUpdate
	case HeartbeatMethod:
		return GasHeartbeat
	case DeregisterMethod:
		return GasDeregister
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
	case RegisterMethod, UpdateAgentMethod, HeartbeatMethod, DeregisterMethod:
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
	case IsAgentMethod:
		return p.isAgent(ctx, method, args)
	case GetAgentMethod:
		return p.getAgent(ctx, method, args)
	case RegisterMethod:
		return p.register(ctx, evm, contract, method, args)
	case UpdateAgentMethod:
		return p.updateAgent(ctx, evm, contract, method, args)
	case HeartbeatMethod:
		return p.heartbeat(ctx, evm, contract, method)
	case DeregisterMethod:
		return p.deregister(ctx, evm, contract, method)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

func (p Precompile) isAgent(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isAgent requires 1 argument")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address argument")
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	result := p.keeper.IsAgent(ctx, cosmosAddr.String())
	return method.Outputs.Pack(result)
}

func (p Precompile) getAgent(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getAgent requires 1 argument")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address argument")
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	agent, found := p.keeper.GetAgent(ctx, cosmosAddr.String())
	if !found {
		return method.Outputs.Pack("", []string{}, "", uint64(0), false)
	}

	isOnline := agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE
	return method.Outputs.Pack(
		agent.AgentId,
		agent.Capabilities,
		agent.Model,
		agent.Reputation,
		isOnline,
	)
}

func (p Precompile) register(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("register requires 2 arguments: capabilities, model")
	}
	capabilities, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("capabilities: expected string, got %T", args[0])
	}
	model, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("model: expected string, got %T", args[1])
	}

	msgValue := contract.Value()
	if msgValue == nil || msgValue.IsZero() {
		return nil, fmt.Errorf("must send AXON as msg.value for staking")
	}

	caller := resolveRegisterSender(evm, contract)
	stakeAmount := sdk.NewCoin("aaxon", sdkmath.NewIntFromBigInt(msgValue.ToBig()))

	// Funds already transferred from sender to precompile address by EVM.
	// Use RegisterFromPrecompile to deduct from precompile address (not sender).
	precompileAddr := sdk.AccAddress(address.Bytes())
	resp, err := p.keeper.RegisterFromPrecompile(ctx, &types.MsgRegister{
		Sender:       caller.String(),
		Capabilities: capabilities,
		Model:        model,
		Stake:        stakeAmount,
	}, precompileAddr)
	if err != nil {
		return nil, err
	}

	_ = resp
	return method.Outputs.Pack()
}

func (p Precompile) updateAgent(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("updateAgent requires 2 arguments")
	}
	capabilities, _ := args[0].(string)
	model, _ := args[1].(string)

	caller := p.resolveAgentSender(ctx, evm, contract)

	msgServer := keeper.NewMsgServerImpl(p.keeper)
	_, err := msgServer.UpdateAgent(ctx, &types.MsgUpdateAgent{
		Sender:       caller.String(),
		Capabilities: capabilities,
		Model:        model,
	})
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

func (p Precompile) heartbeat(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method) ([]byte, error) {
	caller := p.resolveAgentSender(ctx, evm, contract)

	msgServer := keeper.NewMsgServerImpl(p.keeper)
	_, err := msgServer.Heartbeat(ctx, &types.MsgHeartbeat{
		Sender: caller.String(),
	})
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

func (p Precompile) deregister(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method) ([]byte, error) {
	caller := p.resolveAgentSender(ctx, evm, contract)

	msgServer := keeper.NewMsgServerImpl(p.keeper)
	_, err := msgServer.Deregister(ctx, &types.MsgDeregister{
		Sender: caller.String(),
	})
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

// resolveRegisterSender binds new registration to tx origin.
// This avoids intermediary contracts becoming the newly-registered account.
func resolveRegisterSender(evm *vm.EVM, contract *vm.Contract) sdk.AccAddress {
	if evm != nil && evm.Origin != (common.Address{}) {
		return sdk.AccAddress(evm.Origin.Bytes())
	}
	return sdk.AccAddress(contract.Caller().Bytes())
}

// resolveAgentSender keeps compatibility for historical caller-based registrations:
// prefer tx origin if registered; otherwise fall back to caller if registered.
func (p Precompile) resolveAgentSender(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract) sdk.AccAddress {
	caller := sdk.AccAddress(contract.Caller().Bytes())
	if evm == nil || evm.Origin == (common.Address{}) {
		return caller
	}

	origin := sdk.AccAddress(evm.Origin.Bytes())
	if p.keeper.IsAgent(ctx, origin.String()) {
		return origin
	}
	if p.keeper.IsAgent(ctx, caller.String()) {
		return caller
	}
	return origin
}

const abiJSON = `[
	{
		"inputs": [{"name": "account", "type": "address"}],
		"name": "isAgent",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "account", "type": "address"}],
		"name": "getAgent",
		"outputs": [
			{"name": "agentId", "type": "string"},
			{"name": "capabilities", "type": "string[]"},
			{"name": "model", "type": "string"},
			{"name": "reputation", "type": "uint64"},
			{"name": "isOnline", "type": "bool"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "capabilities", "type": "string"},
			{"name": "model", "type": "string"}
		],
		"name": "register",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "capabilities", "type": "string"},
			{"name": "model", "type": "string"}
		],
		"name": "updateAgent",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "heartbeat",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "deregister",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
