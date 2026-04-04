package keeper

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"

	"cosmossdk.io/log/v2"
	storetypes "cosmossdk.io/store/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	privacyKeeper types.PrivacyKeeper
}

const (
	mainnetChainID            = "axon_8210-1"
	V110UpgradeHeight         = int64(259051)
	V111UpgradeHeight         = int64(295500)
	evidenceTxRetentionBlocks = dailyBlockWindow
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
	}
}

func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

func (k *Keeper) SetPrivacyKeeper(privacyKeeper types.PrivacyKeeper) {
	k.privacyKeeper = privacyKeeper
}

func (k Keeper) IsV110UpgradeActivated(ctx sdk.Context) bool {
	if ctx.ChainID() != mainnetChainID {
		return true
	}
	return ctx.BlockHeight() >= V110UpgradeHeight
}

func (k Keeper) IsV111UpgradeActivated(ctx sdk.Context) bool {
	if ctx.ChainID() != mainnetChainID {
		return true
	}
	return ctx.BlockHeight() >= V111UpgradeHeight
}

func (k Keeper) RecordEvidenceTxHash(ctx sdk.Context, txHash common.Hash) {
	if txHash == (common.Hash{}) {
		return
	}
	store := ctx.KVStore(k.storeKey)
	normalized := strings.ToLower(txHash.Hex()[2:])
	store.Set([]byte(types.EvidenceTxKeyPrefix+normalized), types.Uint64ToBytes(uint64(ctx.BlockHeight())))
	heightKey := append([]byte(types.EvidenceTxHeightKeyPrefix), types.Uint64ToBytes(uint64(ctx.BlockHeight()))...)
	heightKey = append(heightKey, []byte("/"+normalized)...)
	store.Set(heightKey, []byte{1})
}

func (k Keeper) HasEvidenceTxHash(ctx sdk.Context, txHash string) bool {
	normalized, ok := normalizeEvidenceHash(txHash)
	if !ok {
		return false
	}
	store := ctx.KVStore(k.storeKey)
	return store.Has([]byte(types.EvidenceTxKeyPrefix + normalized))
}

func (k Keeper) GetLastDailyRegCleanupDay(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.LastDailyRegCleanupDayKey))
	if bz == nil || len(bz) < 8 {
		return -1
	}
	return int64(types.BytesToUint64(bz))
}

func (k Keeper) SetLastDailyRegCleanupDay(ctx sdk.Context, day int64) {
	if day < 0 {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.LastDailyRegCleanupDayKey), types.Uint64ToBytes(uint64(day)))
}

func (k Keeper) shouldFreezeAgentReputationDuringDeregister(ctx sdk.Context, address string) bool {
	return k.IsV111UpgradeActivated(ctx) && k.HasDeregisterRequest(ctx, address)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.ParamsKey))
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set([]byte(types.ParamsKey), bz)
	return nil
}

func (k Keeper) GetAgent(ctx sdk.Context, address string) (types.Agent, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyAgent(address))
	if bz == nil {
		return types.Agent{}, false
	}
	var agent types.Agent
	k.cdc.MustUnmarshal(bz, &agent)
	return agent, true
}

func (k Keeper) SetAgent(ctx sdk.Context, agent types.Agent) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&agent)
	store.Set(types.KeyAgent(agent.Address), bz)
}

func (k Keeper) DeleteAgent(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyAgent(address))
}

func (k Keeper) IterateAgents(ctx sdk.Context, cb func(agent types.Agent) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, []byte(types.AgentKeyPrefix))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var agent types.Agent
		k.cdc.MustUnmarshal(iterator.Value(), &agent)
		if cb(agent) {
			break
		}
	}
}

func (k Keeper) GetAllAgents(ctx sdk.Context) []types.Agent {
	var agents []types.Agent
	k.IterateAgents(ctx, func(agent types.Agent) bool {
		agents = append(agents, agent)
		return false
	})
	return agents
}

func (k Keeper) IsAgent(ctx sdk.Context, address string) bool {
	_, found := k.GetAgent(ctx, address)
	return found
}

func (k Keeper) isActiveValidatorAddress(ctx sdk.Context, address string) bool {
	if k.stakingKeeper == nil {
		return false
	}

	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false
	}

	validator, err := k.stakingKeeper.GetValidator(ctx, sdk.ValAddress(accAddr))
	if err != nil {
		return false
	}

	return validator.IsBonded() && !validator.IsJailed()
}

// Contract deployer tracking for contribution rewards

func (k Keeper) SetContractDeployer(ctx sdk.Context, contractAddr, deployerAddr string) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("ContractDeployer/"+contractAddr), []byte(deployerAddr))
}

func (k Keeper) GetContractDeployer(ctx sdk.Context, contractAddr string) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("ContractDeployer/" + contractAddr))
	if bz == nil {
		return ""
	}
	return string(bz)
}

func (k Keeper) ExportContractDeployers(ctx sdk.Context) map[string]string {
	result := make(map[string]string)
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("ContractDeployer/")
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		contractAddr := string(iterator.Key()[len(prefix):])
		deployerAddr := string(iterator.Value())
		result[contractAddr] = deployerAddr
	}
	return result
}

func (k Keeper) ImportContractDeployers(ctx sdk.Context, deployers map[string]string) {
	for contractAddr, deployerAddr := range deployers {
		k.SetContractDeployer(ctx, contractAddr, deployerAddr)
	}
}

// RegisterFromPrecompile is like MsgServer.Register but deducts stake from
// fundsSource (the precompile address that already received msg.value via EVM)
// instead of from the agent's own address, avoiding double deduction.
func (k Keeper) RegisterFromPrecompile(ctx sdk.Context, msg *types.MsgRegister, fundsSource sdk.AccAddress) (*types.MsgRegisterResponse, error) {
	params := k.GetParams(ctx)

	if k.IsAgent(ctx, msg.Sender) {
		return nil, types.ErrAgentAlreadyRegistered
	}

	if msg.Stake.Denom != "aaxon" {
		return nil, fmt.Errorf("invalid stake denom: expected aaxon, got %s", msg.Stake.Denom)
	}
	minStakeInt := sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(int64(params.MinRegisterStake)), oneAxon))
	minStake := sdk.NewCoin("aaxon", minStakeInt)
	if msg.Stake.IsLT(minStake) {
		return nil, types.ErrInsufficientStake
	}

	if k.GetDailyRegisterCount(ctx, msg.Sender) >= types.MaxDailyRegistrations {
		return nil, types.ErrDailyRegisterLimitExceeded
	}

	stakeCoins := sdk.NewCoins(msg.Stake)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fundsSource, types.ModuleName, stakeCoins); err != nil {
		return nil, err
	}

	burnInt := sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(int64(params.RegisterBurnAmount)), oneAxon))
	burnAmount := sdk.NewCoin("aaxon", burnInt)
	burnCoins := sdk.NewCoins(burnAmount)
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
		return nil, err
	}

	if len(msg.Capabilities) > 1024 {
		return nil, fmt.Errorf("capabilities too long: max 1024 bytes")
	}
	if len(msg.Model) > 256 {
		return nil, fmt.Errorf("model name too long: max 256 bytes")
	}
	capabilities := strings.Split(msg.Capabilities, ",")
	for i := range capabilities {
		capabilities[i] = strings.TrimSpace(capabilities[i])
	}

	agent := types.Agent{
		Address:          msg.Sender,
		AgentId:          generateAgentID(msg.Sender, ctx.BlockHeight()),
		Capabilities:     capabilities,
		Model:            msg.Model,
		Reputation:       params.InitialReputation,
		Status:           types.AgentStatus_AGENT_STATUS_ONLINE,
		StakeAmount:      msg.Stake,
		BurnedAtRegister: burnAmount,
		RegisteredAt:     ctx.BlockHeight(),
		LastHeartbeat:    ctx.BlockHeight(),
	}

	k.SetAgent(ctx, agent)
	k.BootstrapLegacyReputation(ctx, agent.Address, agent.Reputation)
	k.IncrementDailyRegisterCount(ctx, msg.Sender)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_registered",
		sdk.NewAttribute("address", msg.Sender),
		sdk.NewAttribute("agent_id", agent.AgentId),
		sdk.NewAttribute("stake", msg.Stake.String()),
		sdk.NewAttribute("burned", burnAmount.String()),
		sdk.NewAttribute("reputation", fmt.Sprintf("%d", agent.Reputation)),
	))

	return &types.MsgRegisterResponse{AgentId: agent.AgentId}, nil
}

func (k Keeper) AddStakeToAgent(ctx sdk.Context, sender string, stake sdk.Coin, fundsSource sdk.AccAddress) (*types.MsgAddStakeResponse, error) {
	agent, found := k.GetAgent(ctx, sender)
	if !found {
		return nil, types.ErrAgentNotFound
	}
	if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
		return nil, types.ErrAgentSuspended
	}
	if k.HasDeregisterRequest(ctx, sender) {
		return nil, types.ErrDeregisterCooldown
	}

	if stake.Denom != "aaxon" {
		return nil, types.ErrInvalidStakeDenom
	}
	if !stake.IsPositive() {
		return nil, types.ErrStakeAmountMustBePositive
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fundsSource, types.ModuleName, sdk.NewCoins(stake)); err != nil {
		return nil, err
	}

	agent.StakeAmount = agent.StakeAmount.Add(stake)
	k.SetAgent(ctx, agent)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_stake_added",
		sdk.NewAttribute("address", sender),
		sdk.NewAttribute("amount", stake.String()),
		sdk.NewAttribute("total_stake", agent.StakeAmount.String()),
	))

	return &types.MsgAddStakeResponse{TotalStake: agent.StakeAmount}, nil
}

// ReduceStakeFromAgent initiates a stake reduction with an unbonding period.
// The reduced amount is locked until reduceUnlockHeight, then claimable.
func (k Keeper) ReduceStakeFromAgent(ctx sdk.Context, sender string, amount sdk.Coin) error {
	agent, found := k.GetAgent(ctx, sender)
	if !found {
		return types.ErrAgentNotFound
	}
	if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
		return types.ErrAgentSuspended
	}
	if k.HasDeregisterRequest(ctx, sender) {
		return types.ErrDeregisterCooldown
	}
	if amount.Denom != "aaxon" {
		return types.ErrInvalidStakeDenom
	}
	if !amount.IsPositive() {
		return types.ErrStakeAmountMustBePositive
	}

	params := k.GetParams(ctx)
	minStakeInt := sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(int64(params.MinRegisterStake)), oneAxon))
	minStake := sdk.NewCoin("aaxon", minStakeInt)
	remaining := agent.StakeAmount.Sub(amount)
	if remaining.IsLT(minStake) {
		return types.ErrBelowMinimumStake
	}

	if k.hasPendingReduce(ctx, sender) {
		return types.ErrPendingReduceExists
	}

	unlockHeight := ctx.BlockHeight() + types.DeregisterCooldownBlocks

	agent.StakeAmount = remaining
	k.SetAgent(ctx, agent)

	k.setPendingReduce(ctx, sender, amount.Amount, unlockHeight)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_stake_reduce_initiated",
		sdk.NewAttribute("address", sender),
		sdk.NewAttribute("amount", amount.String()),
		sdk.NewAttribute("unlock_height", fmt.Sprintf("%d", unlockHeight)),
		sdk.NewAttribute("remaining_stake", remaining.String()),
	))

	return nil
}

// ClaimReducedStake releases funds after the unbonding period.
func (k Keeper) ClaimReducedStake(ctx sdk.Context, sender string) error {
	amount, unlockHeight, found := k.getPendingReduce(ctx, sender)
	if !found {
		return types.ErrNoReducePending
	}
	if ctx.BlockHeight() < unlockHeight {
		return types.ErrReduceNotUnlocked
	}

	recipientAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin("aaxon", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coins); err != nil {
		return err
	}

	k.deletePendingReduce(ctx, sender)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_stake_reduce_claimed",
		sdk.NewAttribute("address", sender),
		sdk.NewAttribute("amount", amount.String()),
	))

	return nil
}

// GetStakeInfo returns stake details for an agent.
func (k Keeper) GetStakeInfo(ctx sdk.Context, address string) (totalStake sdkmath.Int, pendingReduce sdkmath.Int, reduceUnlockHeight int64, found bool) {
	agent, agentFound := k.GetAgent(ctx, address)
	if !agentFound {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), 0, false
	}
	totalStake = agent.StakeAmount.Amount
	pendingReduce = sdkmath.ZeroInt()
	reduceUnlockHeight = 0
	amt, uh, hasPending := k.getPendingReduce(ctx, address)
	if hasPending {
		pendingReduce = amt
		reduceUnlockHeight = uh
	}
	return totalStake, pendingReduce, reduceUnlockHeight, true
}

func (k Keeper) hasPendingReduce(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.KeyPendingReduceStake(address))
}

func (k Keeper) setPendingReduce(ctx sdk.Context, address string, amount sdkmath.Int, unlockHeight int64) {
	store := ctx.KVStore(k.storeKey)
	amtBz, _ := amount.Marshal()
	heightBz := types.Uint64ToBytes(uint64(unlockHeight))
	value := append(amtBz, heightBz...)
	store.Set(types.KeyPendingReduceStake(address), value)
}

func (k Keeper) getPendingReduce(ctx sdk.Context, address string) (sdkmath.Int, int64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPendingReduceStake(address))
	if bz == nil || len(bz) < 9 {
		return sdkmath.ZeroInt(), 0, false
	}
	amtBz := bz[:len(bz)-8]
	heightBz := bz[len(bz)-8:]
	var amount sdkmath.Int
	if err := amount.Unmarshal(amtBz); err != nil {
		return sdkmath.ZeroInt(), 0, false
	}
	unlockHeight := int64(types.BytesToUint64(heightBz))
	return amount, unlockHeight, true
}

func (k Keeper) deletePendingReduce(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyPendingReduceStake(address))
}

func (k Keeper) GetReputation(ctx sdk.Context, address string) uint64 {
	agent, found := k.GetAgent(ctx, address)
	if !found {
		return 0
	}
	return agent.Reputation
}

func (k Keeper) UpdateReputation(ctx sdk.Context, address string, delta int64) {
	agent, found := k.GetAgent(ctx, address)
	if !found {
		return
	}

	params := k.GetParams(ctx)
	newRep := int64(agent.Reputation) + delta
	if newRep < 0 {
		newRep = 0
	}
	if newRep > int64(params.MaxReputation) {
		newRep = int64(params.MaxReputation)
	}
	agent.Reputation = uint64(newRep)
	k.SetAgent(ctx, agent)
}

func (k Keeper) GetCurrentEpoch(ctx sdk.Context) uint64 {
	params := k.GetParams(ctx)
	if params.EpochLength == 0 {
		return 0
	}
	return uint64(ctx.BlockHeight()) / params.EpochLength
}

const walletKVPrefix = "AgentWallet/"

func (k Keeper) ExportWalletData(ctx sdk.Context) map[string]string {
	result := make(map[string]string)
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(walletKVPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keyHex := hex.EncodeToString(iterator.Key())
		valHex := hex.EncodeToString(iterator.Value())
		result[keyHex] = valHex
	}
	return result
}

func (k Keeper) ImportWalletData(ctx sdk.Context, data map[string]string) {
	store := ctx.KVStore(k.storeKey)
	for keyHex, valHex := range data {
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			continue
		}
		val, err := hex.DecodeString(valHex)
		if err != nil {
			continue
		}
		store.Set(key, val)
	}
}
