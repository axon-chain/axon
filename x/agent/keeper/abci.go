package keeper

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)
	blockHeight := ctx.BlockHeight()
	currentDay := blockHeight / dailyBlockWindow
	upgradeActive := k.IsV110UpgradeActivated(ctx)

	k.processDoubleSignEvidence(ctx)

	// F5 fix: only do O(1) proposer reward per block; accumulate the rest.
	k.AccumulateBlockReward(ctx, params)
	k.MintContributionRewards(ctx)

	if upgradeActive && currentDay > k.GetLastDailyRegCleanupDay(ctx) {
		if k.cleanupOldDailyRegData(ctx, currentDay) {
			k.SetLastDailyRegCleanupDay(ctx, currentDay)
		}
	}
	if upgradeActive && blockHeight > evidenceTxRetentionBlocks {
		cleanupTarget := blockHeight - evidenceTxRetentionBlocks
		// Evidence recording starts at V110UpgradeHeight; skip cleanup for heights
		// before that to avoid pointless empty iterator scans every block.
		if cleanupTarget >= V110UpgradeHeight {
			k.cleanupEvidenceTxHashes(ctx, cleanupTarget)
		}
	}

	if params.EpochLength > 0 && blockHeight > 0 {
		currentEpoch := uint64(blockHeight) / params.EpochLength
		lastProcessedEpoch := k.GetLastProcessedEpoch(ctx)

		if currentEpoch > lastProcessedEpoch {
			maxCatchup := uint64(10)
			start := lastProcessedEpoch + 1
			if currentEpoch-lastProcessedEpoch > maxCatchup {
				start = currentEpoch - maxCatchup + 1
			}
			for e := start; e <= currentEpoch; e++ {
				isCatchup := e < currentEpoch
				k.onEpochStart(ctx, params, e, e-1, isCatchup)
			}
			k.SetLastProcessedEpoch(ctx, currentEpoch)
		}
	}

	k.checkHeartbeatTimeouts(ctx, params)
	k.ProcessDeregisterQueue(ctx)
}

func (k Keeper) EndBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)
	blockHeight := ctx.BlockHeight()

	if params.EpochLength > 0 && blockHeight > 0 && uint64(blockHeight)%params.EpochLength == params.EpochLength-1 {
		k.onEpochEnd(ctx)
	}
}

func (k Keeper) onEpochStart(ctx sdk.Context, params types.Params, epoch, previousEpoch uint64, isCatchup bool) {
	k.Logger(ctx).Info("new epoch started", "epoch", epoch, "catchup", isCatchup)

	if !isCatchup {
		k.GenerateChallenge(ctx, epoch)
	}

	if previousEpoch > 0 {
		if isCatchup {
			// F6 fix: during catchup, clear accumulated rewards to prevent
			// them from being distributed to current validators for historical epochs.
			k.setAccumulatedReward(ctx, accumulatedValidatorRewardKey, sdkmath.ZeroInt())
			k.setAccumulatedReward(ctx, accumulatedReputationRewardKey, sdkmath.ZeroInt())
			k.Logger(ctx).Info("cleared accumulated rewards for catchup epoch",
				"epoch", previousEpoch)
		} else {
			// F5 fix: full distribution only at epoch boundary, not every block.
			k.DistributeAccumulatedBlockRewards(ctx)
			k.DistributeContributionRewards(ctx, previousEpoch)
		}
	}
}

func (k Keeper) onEpochEnd(ctx sdk.Context) {
	epoch := k.GetCurrentEpoch(ctx)
	k.Logger(ctx).Info("epoch ending", "epoch", epoch)

	// AI challenge evaluation (existing)
	k.EvaluateEpochChallenges(ctx, epoch)

	// M2: L1 reputation scoring based on on-chain behavior
	k.ProcessL1Reputation(ctx, epoch)

	// M5+M6+M7+M8: L2 reputation settlement (reports → anti-cheat → budget → apply)
	k.SettleL2Reputation(ctx, epoch)

	if k.IsV110UpgradeActivated(ctx) && epoch > 2 {
		k.cleanupOldEpochData(ctx, epoch-2)
	}

	// M3: Natural decay for both L1 and L2 (AFTER all gains)
	k.ApplyReputationDecay(ctx)

	// Sync legacy agent.Reputation field from L1+L2
	k.SyncLegacyReputation(ctx)

	// NOTE: ProcessEpochReputation (v1 logic) is intentionally removed.
	// v2 dual-layer reputation (L1+L2) fully replaces the old single-layer system.

	// M1: Recompute mining powers for all agents and store for validator updates
	powers := k.ComputeAllMiningPowers(ctx)
	if len(powers) > 0 {
		k.StoreMiningPowers(ctx, powers)
		k.Logger(ctx).Info("mining powers recomputed", "count", len(powers))
	}
}

// checkHeartbeatTimeouts collects offline agents first, then applies changes.
func (k Keeper) checkHeartbeatTimeouts(ctx sdk.Context, params types.Params) {
	blockHeight := ctx.BlockHeight()

	type offlineAgent struct {
		agent types.Agent
	}

	var toOffline []offlineAgent

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE &&
			blockHeight-agent.LastHeartbeat > params.HeartbeatTimeout {
			toOffline = append(toOffline, offlineAgent{agent: agent})
		}
		return false
	})

	for _, o := range toOffline {
		o.agent.Status = types.AgentStatus_AGENT_STATUS_OFFLINE
		k.SetAgent(ctx, o.agent)
		k.UpdateReputation(ctx, o.agent.Address, types.ReputationLossOffline)
		k.ApplyOfflinePenalty(ctx, o.agent.Address)

		k.Logger(ctx).Info("agent went offline",
			"address", o.agent.Address,
			"last_heartbeat", o.agent.LastHeartbeat,
			"current_block", blockHeight,
		)
	}
}

func (k Keeper) GetLastProcessedEpoch(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("LastProcessedEpoch"))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return types.BytesToUint64(bz)
}

func (k Keeper) SetLastProcessedEpoch(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("LastProcessedEpoch"), types.Uint64ToBytes(epoch))
}

// --- F5: Accumulated block reward pool (proposer-only per block, rest at epoch) ---

const (
	accumulatedValidatorRewardKey  = "AccumulatedValidatorReward"
	accumulatedReputationRewardKey = "AccumulatedReputationReward"
)

// AccumulateBlockReward mints the per-block reward and distributes only the
// proposer share immediately. Validator and AI shares are accumulated in KV
// and distributed once per epoch via DistributeAccumulatedBlockRewards.
func (k Keeper) AccumulateBlockReward(ctx sdk.Context, params types.Params) {
	blockHeight := ctx.BlockHeight()
	if blockHeight <= 1 {
		return
	}

	reward := calculateBlockReward(blockHeight)
	if reward.IsZero() {
		return
	}

	maxSupplyBig, ok := new(big.Int).SetString(MaxBlockRewardSupplyStr, 10)
	if !ok {
		return
	}
	totalMinted := k.GetTotalBlockRewardsMinted(ctx)
	remaining := sdkmath.NewIntFromBigInt(new(big.Int).Sub(maxSupplyBig, totalMinted.BigInt()))
	if !remaining.IsPositive() {
		return
	}
	if reward.GT(remaining) {
		reward = remaining
	}

	rewardCoin := sdk.NewCoin("aaxon", reward)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(rewardCoin)); err != nil {
		k.Logger(ctx).Error("failed to mint block rewards", "error", err)
		return
	}
	k.addTotalBlockRewardsMinted(ctx, reward)

	proposerBps := int64(params.ProposerRewardBps)
	validatorBps := int64(params.ValidatorPoolBps)
	if proposerBps <= 0 || validatorBps <= 0 || proposerBps+validatorBps > 10000 {
		proposerBps = int64(ProposerSharePercent) * 100
		validatorBps = int64(ValidatorPoolSharePercent) * 100
	}

	proposerReward := reward.MulRaw(proposerBps).QuoRaw(10000)
	validatorReward := reward.MulRaw(validatorBps).QuoRaw(10000)
	reputationReward := reward.Sub(proposerReward).Sub(validatorReward)

	// O(1) per block: only proposer
	if undistributed := k.distributeProposerReward(ctx, proposerReward); undistributed.IsPositive() {
		validatorReward = validatorReward.Add(undistributed)
	}

	// Accumulate the rest for epoch-boundary distribution
	k.addAccumulatedReward(ctx, accumulatedValidatorRewardKey, validatorReward)
	k.addAccumulatedReward(ctx, accumulatedReputationRewardKey, reputationReward)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"block_rewards",
		sdk.NewAttribute("height", fmt.Sprintf("%d", blockHeight)),
		sdk.NewAttribute("total", rewardCoin.String()),
		sdk.NewAttribute("proposer", proposerReward.String()),
		sdk.NewAttribute("validators_accumulated", validatorReward.String()),
		sdk.NewAttribute("reputation_accumulated", reputationReward.String()),
	))
}

// DistributeAccumulatedBlockRewards flushes the accumulated validator and AI
// reward pools. Called once per epoch, not every block.
func (k Keeper) DistributeAccumulatedBlockRewards(ctx sdk.Context) {
	valReward := k.getAccumulatedReward(ctx, accumulatedValidatorRewardKey)
	if valReward.IsPositive() {
		k.setAccumulatedReward(ctx, accumulatedValidatorRewardKey, sdkmath.ZeroInt())
		if undistributed := k.distributeValidatorRewards(ctx, valReward); undistributed.IsPositive() {
			k.addAccumulatedReward(ctx, accumulatedValidatorRewardKey, undistributed)
		}
	}

	repReward := k.getAccumulatedReward(ctx, accumulatedReputationRewardKey)
	if repReward.IsPositive() {
		k.setAccumulatedReward(ctx, accumulatedReputationRewardKey, sdkmath.ZeroInt())
		if undistributed := k.distributeReputationRewards(ctx, repReward); undistributed.IsPositive() {
			k.addAccumulatedReward(ctx, accumulatedReputationRewardKey, undistributed)
		}
	}
}

func (k Keeper) getAccumulatedReward(ctx sdk.Context, key string) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(key))
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	var amount sdkmath.Int
	if err := amount.Unmarshal(bz); err != nil {
		return sdkmath.ZeroInt()
	}
	return amount
}

func (k Keeper) setAccumulatedReward(ctx sdk.Context, key string, amount sdkmath.Int) {
	bz, err := amount.Marshal()
	if err != nil {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), bz)
}

func (k Keeper) addAccumulatedReward(ctx sdk.Context, key string, amount sdkmath.Int) {
	current := k.getAccumulatedReward(ctx, key)
	k.setAccumulatedReward(ctx, key, current.Add(amount))
}

const jailedFlagKeyPrefix = "AgentJailedFlag/"

// processDoubleSignEvidence detects newly jailed validators among registered
// agents and applies the L1 double-sign penalty exactly once per jailing event.
func (k Keeper) processDoubleSignEvidence(ctx sdk.Context) {
	if k.stakingKeeper == nil {
		return
	}

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
			return false
		}

		accAddr, err := sdk.AccAddressFromBech32(agent.Address)
		if err != nil {
			return false
		}
		validator, err := k.stakingKeeper.GetValidator(ctx, sdk.ValAddress(accAddr))
		if err != nil {
			return false
		}

		store := ctx.KVStore(k.storeKey)
		flagKey := []byte(jailedFlagKeyPrefix + agent.Address)
		wasJailed := store.Has(flagKey)

		if validator.IsJailed() && !wasJailed {
			k.ApplyDoubleSignPenalty(ctx, agent.Address)
			k.UpdateReputation(ctx, agent.Address, types.ReputationLossSlashing)
			store.Set(flagKey, []byte{1})
			k.Logger(ctx).Info("double-sign penalty applied",
				"address", agent.Address,
				"block", ctx.BlockHeight(),
			)
		} else if !validator.IsJailed() && wasJailed {
			store.Delete(flagKey)
		}

		return false
	})
}
