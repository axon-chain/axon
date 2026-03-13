package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)
	blockHeight := ctx.BlockHeight()

	k.DistributeBlockRewards(ctx)
	k.MintContributionRewards(ctx)

	// Epoch transition — process all skipped epochs to avoid losing rewards
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

	// Only generate challenges for the real current epoch.
	// During catchup, intermediate epochs have already passed — no one can respond.
	if !isCatchup {
		k.GenerateChallenge(ctx, epoch)
	}

	if previousEpoch > 0 {
		k.DistributeEpochRewards(ctx, previousEpoch)
		k.DistributeContributionRewards(ctx, previousEpoch)
	}
}

func (k Keeper) onEpochEnd(ctx sdk.Context) {
	epoch := k.GetCurrentEpoch(ctx)
	k.Logger(ctx).Info("epoch ending", "epoch", epoch)

	k.EvaluateEpochChallenges(ctx, epoch)
	k.ProcessEpochReputation(ctx, epoch)
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

		k.Logger(ctx).Info("agent went offline",
			"address", o.agent.Address,
			"last_heartbeat", o.agent.LastHeartbeat,
			"current_block", blockHeight,
		)
	}
}

// Epoch state tracking to survive param changes

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
