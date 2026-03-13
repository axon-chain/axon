package wallet

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	cmn "github.com/cosmos/evm/precompiles/common"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/keeper"
)

var (
	contractAddress = common.HexToAddress("0x0000000000000000000000000000000000000803")
	_               = vm.PrecompiledContract(&Precompile{})
)

// Method names
const (
	MethodCreateWallet  = "createWallet"
	MethodExecute       = "execute"
	MethodSetTrust      = "setTrust"
	MethodRemoveTrust   = "removeTrust"
	MethodFreeze        = "freeze"
	MethodRecover       = "recover"
	MethodGetWalletInfo = "getWalletInfo"
	MethodGetTrust      = "getTrust"
)

// Gas costs
const (
	GasCreateWallet  = 50000
	GasExecute       = 30000
	GasSetTrust      = 20000
	GasRemoveTrust   = 10000
	GasFreeze        = 10000
	GasRecover       = 30000
	GasGetWalletInfo = 1000
	GasGetTrust      = 1000
)

const BlocksPerDay int64 = 17280 // at ~5 s/block

type Precompile struct {
	cmn.Precompile
	abi    abi.ABI
	keeper keeper.Keeper
}

func NewPrecompile(k keeper.Keeper, bankKeeper cmn.BankKeeper) (*Precompile, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IAgentWallet ABI: %w", err)
	}
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:           storetypes.KVGasConfig(),
			TransientKVGasConfig:  storetypes.GasConfig{},
			ContractAddress:       contractAddress,
			BalanceHandlerFactory: cmn.NewBalanceHandlerFactory(bankKeeper),
		},
		abi:    parsed,
		keeper: k,
	}, nil
}

func (Precompile) Address() common.Address { return contractAddress }

func (p Precompile) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 3000
	}
	method, err := p.abi.MethodById(input[:4])
	if err != nil {
		return 3000
	}
	switch method.Name {
	case MethodCreateWallet:
		return GasCreateWallet
	case MethodExecute:
		return GasExecute
	case MethodSetTrust:
		return GasSetTrust
	case MethodRemoveTrust:
		return GasRemoveTrust
	case MethodFreeze:
		return GasFreeze
	case MethodRecover:
		return GasRecover
	case MethodGetWalletInfo:
		return GasGetWalletInfo
	case MethodGetTrust:
		return GasGetTrust
	default:
		return 3000
	}
}

func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.dispatch(ctx, contract, readonly, evm)
	})
}

func (p Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case MethodCreateWallet, MethodExecute, MethodSetTrust, MethodRemoveTrust, MethodFreeze, MethodRecover:
		return true
	default:
		return false
	}
}

func (p Precompile) dispatch(ctx sdk.Context, contract *vm.Contract, readOnly bool, evm *vm.EVM) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.abi, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	switch method.Name {
	case MethodCreateWallet:
		return p.createWallet(ctx, contract, method, args)
	case MethodExecute:
		return p.executeWallet(ctx, contract, method, args, evm)
	case MethodSetTrust:
		return p.setTrust(ctx, contract, method, args)
	case MethodRemoveTrust:
		return p.removeTrust(ctx, contract, method, args)
	case MethodFreeze:
		return p.freezeWallet(ctx, contract, method, args)
	case MethodRecover:
		return p.recoverWallet(ctx, contract, method, args)
	case MethodGetWalletInfo:
		return p.getWalletInfo(ctx, method, args)
	case MethodGetTrust:
		return p.getTrustInfo(ctx, method, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", method.Name)
	}
}

// ============================================================
// Write methods
// ============================================================

// createWallet: caller becomes Owner; specifies operator, guardian, and default limits.
func (p Precompile) createWallet(ctx sdk.Context, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 5 {
		return nil, fmt.Errorf("createWallet requires 5 arguments")
	}

	operator, _ := args[0].(common.Address)
	guardian, _ := args[1].(common.Address)
	txLimit, _ := args[2].(*big.Int)
	dailyLimit, _ := args[3].(*big.Int)
	cooldownBlocks, _ := args[4].(*big.Int)

	zeroAddr := common.Address{}
	if operator == zeroAddr {
		return nil, fmt.Errorf("operator cannot be zero address")
	}
	if guardian == zeroAddr {
		return nil, fmt.Errorf("guardian cannot be zero address")
	}
	if operator == guardian {
		return nil, fmt.Errorf("operator and guardian must be different addresses")
	}

	owner := contract.Caller()
	walletAddr := generateWalletAddress(owner, ctx.BlockHeight())

	if _, exists := p.loadWallet(ctx, walletAddr); exists {
		return nil, fmt.Errorf("wallet already exists for this address in this block")
	}

	wallet := WalletInfo{
		Owner:          owner,
		Operator:       operator,
		Guardian:       guardian,
		TxLimit:        txLimit,
		DailyLimit:     dailyLimit,
		CooldownBlocks: cooldownBlocks,
		DailySpent:     big.NewInt(0),
		LastResetBlock: ctx.BlockHeight(),
		IsFrozen:       false,
	}

	p.storeWallet(ctx, walletAddr, wallet)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_wallet_created",
		sdk.NewAttribute("wallet", walletAddr.Hex()),
		sdk.NewAttribute("owner", owner.Hex()),
		sdk.NewAttribute("operator", operator.Hex()),
		sdk.NewAttribute("guardian", guardian.Hex()),
	))

	return method.Outputs.Pack(walletAddr)
}

// executeWallet: trust-aware transaction execution by Operator.
//
// Resolution order:
//  1. Wallet frozen? → reject
//  2. Caller == Operator? → reject if not
//  3. Lookup TrustedChannel for target
//     - TrustBlocked → reject
//     - TrustFull    → execute immediately, no caps
//     - TrustLimited → per-channel txLimit + dailyLimit
//     - Unknown      → wallet default txLimit + dailyLimit
func (p Precompile) executeWallet(ctx sdk.Context, contract *vm.Contract, method *abi.Method, args []interface{}, evm *vm.EVM) ([]byte, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("execute requires 4 arguments")
	}

	walletAddr, _ := args[0].(common.Address)
	target, _ := args[1].(common.Address)
	value, _ := args[2].(*big.Int)
	data, _ := args[3].([]byte)

	if len(data) > 0 {
		return nil, fmt.Errorf("contract calls via smart wallet will be available in a future upgrade; currently only value transfers are supported")
	}

	wallet, found := p.loadWallet(ctx, walletAddr)
	if !found {
		return nil, fmt.Errorf("wallet not found")
	}
	if wallet.IsFrozen {
		return nil, fmt.Errorf("wallet is frozen")
	}
	if contract.Caller() != wallet.Operator {
		return nil, fmt.Errorf("only operator can execute")
	}

	operatorBech32 := sdk.AccAddress(wallet.Operator.Bytes()).String()
	if _, agentExists := p.keeper.GetAgent(ctx, operatorBech32); !agentExists {
		return nil, fmt.Errorf("operator agent is no longer registered")
	}

	channel, hasTrust := p.loadTrustChannel(ctx, walletAddr, target)

	// TrustBlocked never expires — check it before applying expiry logic
	if hasTrust && channel.Level == TrustBlocked {
		return nil, fmt.Errorf("target is blacklisted")
	}

	if hasTrust && channel.ExpiresAt > 0 && ctx.BlockHeight() > channel.ExpiresAt {
		hasTrust = false
	}

	if hasTrust {
		switch channel.Level {

		case TrustFull:
			if err := p.doTransfer(evm, walletAddr, target, value); err != nil {
				return nil, err
			}
			return method.Outputs.Pack()

		case TrustLimited:
			if err := p.checkChannelLimits(ctx, walletAddr, target, value, channel); err != nil {
				return nil, err
			}
			if err := p.doTransfer(evm, walletAddr, target, value); err != nil {
				return nil, err
			}
			return method.Outputs.Pack()
		}
	}

	// Unknown target — apply wallet-level default limits
	if err := p.checkDefaultLimits(ctx, walletAddr, value, &wallet); err != nil {
		return nil, err
	}
	p.storeWallet(ctx, walletAddr, wallet)
	if err := p.doTransfer(evm, walletAddr, target, value); err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

// setTrust: Owner authorizes a trust channel for a target contract.
func (p Precompile) setTrust(ctx sdk.Context, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 6 {
		return nil, fmt.Errorf("setTrust requires 6 arguments")
	}

	walletAddr, _ := args[0].(common.Address)
	target, _ := args[1].(common.Address)
	level, _ := args[2].(*big.Int)
	txLimit, _ := args[3].(*big.Int)
	dailyLimit, _ := args[4].(*big.Int)
	expiresAt, _ := args[5].(*big.Int)

	wallet, found := p.loadWallet(ctx, walletAddr)
	if !found {
		return nil, fmt.Errorf("wallet not found")
	}
	if contract.Caller() != wallet.Owner {
		return nil, fmt.Errorf("only owner can set trust level")
	}

	if !level.IsUint64() || level.Uint64() > uint64(TrustFull) {
		return nil, fmt.Errorf("invalid trust level: must be 0-3")
	}
	trustLevel := TrustLevel(level.Uint64())

	maxInt64 := new(big.Int).SetInt64(math.MaxInt64)
	if expiresAt == nil || expiresAt.Sign() < 0 || expiresAt.Cmp(maxInt64) > 0 {
		return nil, fmt.Errorf("expiresAt out of range: must be 0..%d", math.MaxInt64)
	}

	tc := TrustedChannel{
		Level:        trustLevel,
		TxLimit:      txLimit,
		DailyLimit:   dailyLimit,
		AuthorizedAt: ctx.BlockHeight(),
		ExpiresAt:    expiresAt.Int64(),
	}
	p.storeTrustChannel(ctx, walletAddr, target, tc)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_trust_set",
		sdk.NewAttribute("wallet", walletAddr.Hex()),
		sdk.NewAttribute("target", target.Hex()),
		sdk.NewAttribute("level", fmt.Sprintf("%d", trustLevel)),
	))

	return method.Outputs.Pack()
}

// removeTrust: Owner revokes a trust channel.
func (p Precompile) removeTrust(ctx sdk.Context, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("removeTrust requires 2 arguments")
	}

	walletAddr, _ := args[0].(common.Address)
	target, _ := args[1].(common.Address)

	wallet, found := p.loadWallet(ctx, walletAddr)
	if !found {
		return nil, fmt.Errorf("wallet not found")
	}
	if contract.Caller() != wallet.Owner {
		return nil, fmt.Errorf("only owner can remove trust")
	}

	p.deleteTrustChannel(ctx, walletAddr, target)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_trust_removed",
		sdk.NewAttribute("wallet", walletAddr.Hex()),
		sdk.NewAttribute("target", target.Hex()),
	))

	return method.Outputs.Pack()
}

// freezeWallet: Guardian OR Owner can freeze a wallet.
func (p Precompile) freezeWallet(ctx sdk.Context, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("freeze requires 1 argument")
	}
	walletAddr, _ := args[0].(common.Address)

	wallet, found := p.loadWallet(ctx, walletAddr)
	if !found {
		return nil, fmt.Errorf("wallet not found")
	}

	caller := contract.Caller()
	if caller != wallet.Guardian && caller != wallet.Owner {
		return nil, fmt.Errorf("only guardian or owner can freeze")
	}

	wallet.IsFrozen = true
	p.storeWallet(ctx, walletAddr, wallet)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_wallet_frozen",
		sdk.NewAttribute("wallet", walletAddr.Hex()),
		sdk.NewAttribute("by", caller.Hex()),
	))

	return method.Outputs.Pack()
}

// recoverWallet: Guardian replaces the Operator and auto-unfreezes.
func (p Precompile) recoverWallet(ctx sdk.Context, contract *vm.Contract, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("recover requires 2 arguments")
	}
	walletAddr, _ := args[0].(common.Address)
	newOperator, _ := args[1].(common.Address)

	wallet, found := p.loadWallet(ctx, walletAddr)
	if !found {
		return nil, fmt.Errorf("wallet not found")
	}
	if contract.Caller() != wallet.Guardian {
		return nil, fmt.Errorf("only guardian can recover")
	}
	if newOperator == (common.Address{}) {
		return nil, fmt.Errorf("new operator cannot be zero address")
	}

	wallet.Operator = newOperator
	wallet.IsFrozen = false
	p.storeWallet(ctx, walletAddr, wallet)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_wallet_recovered",
		sdk.NewAttribute("wallet", walletAddr.Hex()),
		sdk.NewAttribute("new_operator", newOperator.Hex()),
	))

	return method.Outputs.Pack()
}

// ============================================================
// Read methods
// ============================================================

func (p Precompile) getWalletInfo(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getWalletInfo requires 1 argument")
	}
	walletAddr, _ := args[0].(common.Address)

	wallet, found := p.loadWallet(ctx, walletAddr)
	if !found {
		zero := big.NewInt(0)
		return method.Outputs.Pack(
			zero, zero, zero,
			false, common.Address{}, common.Address{}, common.Address{},
		)
	}

	dailySpent := wallet.DailySpent
	if ctx.BlockHeight()-wallet.LastResetBlock >= BlocksPerDay {
		dailySpent = big.NewInt(0)
	}

	return method.Outputs.Pack(
		wallet.TxLimit,
		wallet.DailyLimit,
		dailySpent,
		wallet.IsFrozen,
		wallet.Owner,
		wallet.Operator,
		wallet.Guardian,
	)
}

func (p Precompile) getTrustInfo(ctx sdk.Context, method *abi.Method, args []interface{}) ([]byte, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("getTrust requires 2 arguments")
	}
	walletAddr, _ := args[0].(common.Address)
	target, _ := args[1].(common.Address)

	tc, found := p.loadTrustChannel(ctx, walletAddr, target)
	if !found {
		zero := big.NewInt(0)
		return method.Outputs.Pack(
			big.NewInt(int64(TrustUnknown)),
			zero, zero, zero, zero,
		)
	}

	return method.Outputs.Pack(
		big.NewInt(int64(tc.Level)),
		tc.TxLimit,
		tc.DailyLimit,
		big.NewInt(tc.AuthorizedAt),
		big.NewInt(tc.ExpiresAt),
	)
}

// ============================================================
// Limit checking helpers
// ============================================================

// checkChannelLimits enforces per-channel txLimit and dailyLimit for TrustLimited targets.
func (p Precompile) checkChannelLimits(ctx sdk.Context, walletAddr, target common.Address, value *big.Int, ch TrustedChannel) error {
	if ch.TxLimit != nil && ch.TxLimit.Sign() > 0 && value.Cmp(ch.TxLimit) > 0 {
		return fmt.Errorf("exceeds channel tx limit")
	}
	if ch.DailyLimit != nil && ch.DailyLimit.Sign() > 0 {
		spent := p.loadDailySpentTo(ctx, walletAddr, target)
		if ctx.BlockHeight()-spent.ResetBlock >= BlocksPerDay {
			spent = DailySpentEntry{Amount: big.NewInt(0), ResetBlock: ctx.BlockHeight()}
		}
		newSpent := new(big.Int).Add(spent.Amount, value)
		if newSpent.Cmp(ch.DailyLimit) > 0 {
			return fmt.Errorf("exceeds channel daily limit")
		}
		spent.Amount = newSpent
		p.storeDailySpentTo(ctx, walletAddr, target, spent)
	}
	return nil
}

// checkDefaultLimits enforces wallet-level txLimit and dailyLimit for unknown targets.
func (p Precompile) checkDefaultLimits(ctx sdk.Context, walletAddr common.Address, value *big.Int, wallet *WalletInfo) error {
	if wallet.TxLimit != nil && wallet.TxLimit.Sign() > 0 && value.Cmp(wallet.TxLimit) > 0 {
		return fmt.Errorf("exceeds default tx limit")
	}

	if ctx.BlockHeight()-wallet.LastResetBlock >= BlocksPerDay {
		wallet.DailySpent = big.NewInt(0)
		wallet.LastResetBlock = ctx.BlockHeight()
	}

	newSpent := new(big.Int).Add(wallet.DailySpent, value)
	if wallet.DailyLimit != nil && wallet.DailyLimit.Sign() > 0 && newSpent.Cmp(wallet.DailyLimit) > 0 {
		return fmt.Errorf("exceeds default daily limit")
	}
	wallet.DailySpent = newSpent
	return nil
}

func (p Precompile) doTransfer(evm *vm.EVM, from, to common.Address, value *big.Int) error {
	if value.Sign() > 0 {
		val, overflow := uint256.FromBig(value)
		if overflow {
			return fmt.Errorf("transfer value exceeds uint256 max")
		}
		balance := evm.StateDB.GetBalance(from)
		if balance.Cmp(val) < 0 {
			return fmt.Errorf("insufficient wallet balance: have %s, need %s", balance, val)
		}
		evm.Context.Transfer(evm.StateDB, from, to, val)
	}
	return nil
}

// ============================================================
// KV store helpers
// ============================================================

func (p Precompile) storeWallet(ctx sdk.Context, addr common.Address, w WalletInfo) {
	store := ctx.KVStore(p.keeper.StoreKey())
	store.Set(walletKey(addr), encodeWallet(w))
}

func (p Precompile) loadWallet(ctx sdk.Context, addr common.Address) (WalletInfo, bool) {
	store := ctx.KVStore(p.keeper.StoreKey())
	bz := store.Get(walletKey(addr))
	if bz == nil {
		return WalletInfo{}, false
	}
	return decodeWallet(bz), true
}

func (p Precompile) storeTrustChannel(ctx sdk.Context, wallet, target common.Address, tc TrustedChannel) {
	store := ctx.KVStore(p.keeper.StoreKey())
	store.Set(trustKey(wallet, target), encodeTrustChannel(tc))
}

func (p Precompile) loadTrustChannel(ctx sdk.Context, wallet, target common.Address) (TrustedChannel, bool) {
	store := ctx.KVStore(p.keeper.StoreKey())
	bz := store.Get(trustKey(wallet, target))
	if bz == nil {
		return TrustedChannel{}, false
	}
	return decodeTrustChannel(bz), true
}

func (p Precompile) deleteTrustChannel(ctx sdk.Context, wallet, target common.Address) {
	store := ctx.KVStore(p.keeper.StoreKey())
	store.Delete(trustKey(wallet, target))
}

func (p Precompile) loadDailySpentTo(ctx sdk.Context, wallet, target common.Address) DailySpentEntry {
	store := ctx.KVStore(p.keeper.StoreKey())
	bz := store.Get(spentToKey(wallet, target))
	if bz == nil {
		return DailySpentEntry{Amount: big.NewInt(0), ResetBlock: ctx.BlockHeight()}
	}
	return decodeDailySpent(bz)
}

func (p Precompile) storeDailySpentTo(ctx sdk.Context, wallet, target common.Address, e DailySpentEntry) {
	store := ctx.KVStore(p.keeper.StoreKey())
	store.Set(spentToKey(wallet, target), encodeDailySpent(e))
}

// ============================================================
// ABI definition
// ============================================================

const abiJSON = `[
	{
		"inputs": [
			{"name": "operator", "type": "address"},
			{"name": "guardian", "type": "address"},
			{"name": "txLimit", "type": "uint256"},
			{"name": "dailyLimit", "type": "uint256"},
			{"name": "cooldownBlocks", "type": "uint256"}
		],
		"name": "createWallet",
		"outputs": [{"name": "wallet", "type": "address"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "wallet", "type": "address"},
			{"name": "target", "type": "address"},
			{"name": "value", "type": "uint256"},
			{"name": "data", "type": "bytes"}
		],
		"name": "execute",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "wallet", "type": "address"},
			{"name": "target", "type": "address"},
			{"name": "level", "type": "uint256"},
			{"name": "txLimit", "type": "uint256"},
			{"name": "dailyLimit", "type": "uint256"},
			{"name": "expiresAt", "type": "uint256"}
		],
		"name": "setTrust",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "wallet", "type": "address"},
			{"name": "target", "type": "address"}
		],
		"name": "removeTrust",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "wallet", "type": "address"}],
		"name": "freeze",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "wallet", "type": "address"},
			{"name": "newOperator", "type": "address"}
		],
		"name": "recover",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "wallet", "type": "address"}],
		"name": "getWalletInfo",
		"outputs": [
			{"name": "txLimit", "type": "uint256"},
			{"name": "dailyLimit", "type": "uint256"},
			{"name": "dailySpent", "type": "uint256"},
			{"name": "isFrozen", "type": "bool"},
			{"name": "owner", "type": "address"},
			{"name": "operator", "type": "address"},
			{"name": "guardian", "type": "address"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "wallet", "type": "address"},
			{"name": "target", "type": "address"}
		],
		"name": "getTrust",
		"outputs": [
			{"name": "level", "type": "uint256"},
			{"name": "txLimit", "type": "uint256"},
			{"name": "dailyLimit", "type": "uint256"},
			{"name": "authorizedAt", "type": "uint256"},
			{"name": "expiresAt", "type": "uint256"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`
