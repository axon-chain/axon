package registry

import (
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

	"github.com/axon-chain/axon/x/agent/keeper"
	"github.com/axon-chain/axon/x/agent/types"
)

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000801")
	_       = vm.PrecompiledContract(&Precompile{})
)

const (
	IsAgentMethod           = "isAgent"
	GetAgentMethod          = "getAgent"
	RegisterMethod          = "register"
	AddStakeMethod          = "addStake"
	ReduceStakeMethod       = "reduceStake"
	ClaimReducedStakeMethod = "claimReducedStake"
	GetStakeInfoMethod      = "getStakeInfo"
	UpdateAgentMethod       = "updateAgent"
	HeartbeatMethod         = "heartbeat"
	DeregisterMethod        = "deregister"

	GasIsAgent           = 200
	GasGetAgent          = 1000
	GasRegister          = 50000
	GasAddStake          = 30000
	GasReduceStake       = 30000
	GasClaimReducedStake = 30000
	GasGetStakeInfo      = 500
	GasUpdate            = 10000
	GasHeartbeat         = 5000
	GasDeregister        = 20000

	mainnetChainID                   = "axon_8210-1"
	legacyMutationSenderCompatHeight = int64(18392)
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
	case AddStakeMethod:
		return GasAddStake
	case ReduceStakeMethod:
		return GasReduceStake
	case ClaimReducedStakeMethod:
		return GasClaimReducedStake
	case GetStakeInfoMethod:
		return GasGetStakeInfo
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
	case RegisterMethod, AddStakeMethod, ReduceStakeMethod, ClaimReducedStakeMethod, UpdateAgentMethod, HeartbeatMethod, DeregisterMethod:
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
	case AddStakeMethod:
		return p.addStake(ctx, evm, contract, method)
	case ReduceStakeMethod:
		return p.reduceStake(ctx, evm, contract, method, args)
	case ClaimReducedStakeMethod:
		return p.claimReducedStake(ctx, evm, contract, method)
	case GetStakeInfoMethod:
		return p.getStakeInfo(ctx, evm, contract, method, args)
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

	caller := p.resolveRegisterMutationSender(ctx, evm, contract)
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

func (p Precompile) addStake(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method) ([]byte, error) {
	msgValue := contract.Value()
	if msgValue == nil || msgValue.IsZero() {
		return nil, fmt.Errorf("must send AXON as msg.value for addStake")
	}

	caller := p.resolveMutationSender(ctx, evm, contract)
	stakeAmount := sdk.NewCoin("aaxon", sdkmath.NewIntFromBigInt(msgValue.ToBig()))
	precompileAddr := sdk.AccAddress(address.Bytes())

	resp, err := p.keeper.AddStakeToAgent(ctx, caller.String(), stakeAmount, precompileAddr)
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

	caller := p.resolveMutationSender(ctx, evm, contract)

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
	caller := p.resolveMutationSender(ctx, evm, contract)

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
	caller := p.resolveMutationSender(ctx, evm, contract)

	msgServer := keeper.NewMsgServerImpl(p.keeper)
	_, err := msgServer.Deregister(ctx, &types.MsgDeregister{
		Sender: caller.String(),
	})
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

func (p Precompile) reduceStake(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("reduceStake requires 1 argument: amount")
	}
	amountBig, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("amount: expected *big.Int, got %T", args[0])
	}

	caller := p.resolveMutationSender(ctx, evm, contract)
	amount := sdk.NewCoin("aaxon", sdkmath.NewIntFromBigInt(amountBig))

	if err := p.keeper.ReduceStakeFromAgent(ctx, caller.String(), amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack()
}

func (p Precompile) claimReducedStake(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract, method *abi.Method) ([]byte, error) {
	caller := p.resolveMutationSender(ctx, evm, contract)
	if err := p.keeper.ClaimReducedStake(ctx, caller.String()); err != nil {
		return nil, err
	}

	return method.Outputs.Pack()
}

func (p Precompile) getStakeInfo(ctx sdk.Context, _ *vm.EVM, _ *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getStakeInfo requires 1 argument: agent address")
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address argument")
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	totalStake, pendingReduce, reduceUnlockHeight, found := p.keeper.GetStakeInfo(ctx, cosmosAddr.String())
	if !found {
		return method.Outputs.Pack(big.NewInt(0), big.NewInt(0), uint64(0))
	}

	return method.Outputs.Pack(totalStake.BigInt(), pendingReduce.BigInt(), uint64(reduceUnlockHeight))
}

func useLegacyMutationSenderCompat(ctx sdk.Context) bool {
	return ctx.ChainID() == mainnetChainID && ctx.BlockHeight() <= legacyMutationSenderCompatHeight
}

func (p Precompile) resolveRegisterMutationSender(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract) sdk.AccAddress {
	if useLegacyMutationSenderCompat(ctx) {
		return resolveRegisterSender(evm, contract)
	}
	return resolveMutationSender(contract, evm)
}

func (p Precompile) resolveMutationSender(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract) sdk.AccAddress {
	if useLegacyMutationSenderCompat(ctx) {
		return p.resolveAgentSender(ctx, evm, contract)
	}
	return resolveMutationSender(contract, evm)
}

// resolveMutationSender binds registry mutations to the immediate EVM caller.
// This keeps the identity owner and the msg.value payer aligned and matches the
// semantics used by Cosmos EVM's built-in stateful precompiles.
func resolveMutationSender(contract *vm.Contract, evm *vm.EVM) sdk.AccAddress {
	if contract != nil {
		caller := contract.Caller()
		if caller != (common.Address{}) {
			return sdk.AccAddress(caller.Bytes())
		}
	}
	if evm != nil && evm.Origin != (common.Address{}) {
		return sdk.AccAddress(evm.Origin.Bytes())
	}
	return sdk.AccAddress{}
}

// resolveRegisterSender binds new registration to tx origin.
// This avoids intermediary contracts becoming the newly-registered account.
func resolveRegisterSender(evm *vm.EVM, contract *vm.Contract) sdk.AccAddress {
	if evm != nil && evm.Origin != (common.Address{}) {
		return sdk.AccAddress(evm.Origin.Bytes())
	}
	if contract != nil {
		return sdk.AccAddress(contract.Caller().Bytes())
	}
	return sdk.AccAddress{}
}

// resolveAgentSender keeps compatibility for historical caller-based registrations:
// prefer tx origin if registered; otherwise fall back to caller if registered.
func (p Precompile) resolveAgentSender(ctx sdk.Context, evm *vm.EVM, contract *vm.Contract) sdk.AccAddress {
	caller := sdk.AccAddress{}
	if contract != nil {
		caller = sdk.AccAddress(contract.Caller().Bytes())
	}
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
		"inputs": [],
		"name": "addStake",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "reduceStake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "claimReducedStake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "account", "type": "address"}],
		"name": "getStakeInfo",
		"outputs": [
			{"name": "totalStake", "type": "uint256"},
			{"name": "pendingReduce", "type": "uint256"},
			{"name": "reduceUnlockHeight", "type": "uint64"}
		],
		"stateMutability": "view",
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
