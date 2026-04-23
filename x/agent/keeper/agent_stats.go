package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

const (
	AgentStatsKeyPrefix     = "AgentStats/"
	ReputationHistoryPrefix = "ReputationHistory/"
	ChallengeStatsPrefix    = "ChallengeStats/"
)

func (k Keeper) InitAgentStats(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	stats := types.AgentStatistics{
		Address:               address,
		TotalChallenges:       0,
		SuccessfulChallenges: 0,
		TotalResponses:       0,
		LastChallengeEpoch:   0,
		ConsecutiveSuccesses: 0,
		ConsecutiveFailures:  0,
		SuccessRate:          0,
	}
	bz, _ := json.Marshal(&stats)
	store.Set([]byte(AgentStatsKeyPrefix+address), bz)
}

func (k Keeper) GetAgentStats(ctx sdk.Context, address string) (types.AgentStatistics, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(AgentStatsKeyPrefix + address))
	if bz == nil {
		return types.AgentStatistics{}, false
	}
	var stats types.AgentStatistics
	if err := json.Unmarshal(bz, &stats); err != nil {
		return types.AgentStatistics{}, false
	}
	return stats, true
}

func (k Keeper) SetAgentStats(ctx sdk.Context, stats types.AgentStatistics) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(&stats)
	store.Set([]byte(AgentStatsKeyPrefix+stats.Address), bz)
}

func (k Keeper) UpdateChallengeStats(ctx sdk.Context, address string, epoch uint64, score int64) {
	stats, found := k.GetAgentStats(ctx, address)
	if !found {
		k.InitAgentStats(ctx, address)
		stats, _ = k.GetAgentStats(ctx, address)
	}

	stats.TotalChallenges++
	stats.TotalResponses++
	stats.LastChallengeEpoch = epoch

	if score >= 50 {
		stats.SuccessfulChallenges++
		stats.ConsecutiveSuccesses++
		stats.ConsecutiveFailures = 0
	} else if score >= 0 {
		stats.ConsecutiveFailures++
		stats.ConsecutiveSuccesses = 0
	}

	if stats.TotalChallenges > 0 {
		stats.SuccessRate = float64(stats.SuccessfulChallenges) / float64(stats.TotalChallenges)
	}

	k.SetAgentStats(ctx, stats)
}

func (k Keeper) RecordReputationHistory(ctx sdk.Context, address string, oldRep, newRep uint64, reason string) {
	store := ctx.KVStore(k.storeKey)

	epoch := k.GetCurrentEpoch(ctx)
	entry := types.ReputationHistoryEntry{
		Epoch:       epoch,
		BlockHeight: ctx.BlockHeight(),
		OldRep:     oldRep,
		NewRep:     newRep,
		Delta:      int64(newRep) - int64(oldRep),
		Reason:     reason,
	}

	key := []byte(ReputationHistoryPrefix + address + "/")
	key = append(key, types.Uint64ToBytes(epoch)...)

	bz, _ := json.Marshal(&entry)
	store.Set(key, bz)
}

func (k Keeper) GetReputationHistory(ctx sdk.Context, address string, limit uint64) []types.ReputationHistoryEntry {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(ReputationHistoryPrefix + address + "/")

	var entries []types.ReputationHistoryEntry
	count := uint64(0)

	iterator := storetypes.KVStoreReversePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid() && count < limit; iterator.Next() {
		var entry types.ReputationHistoryEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err == nil {
			entries = append(entries, entry)
			count++
		}
	}

	return entries
}

func (k Keeper) UpdateReputationWithHistory(ctx sdk.Context, address string, delta int64, reason string) {
	agent, found := k.GetAgent(ctx, address)
	if !found {
		return
	}

	oldRep := agent.Reputation

	k.UpdateReputation(ctx, address, delta)

	agent, _ = k.GetAgent(ctx, address)

	k.RecordReputationHistory(ctx, address, oldRep, agent.Reputation, reason)
}

func (k Keeper) GetAllAgentStats(ctx sdk.Context) []types.AgentStatistics {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(AgentStatsKeyPrefix)

	var allStats []types.AgentStatistics
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stats types.AgentStatistics
		if err := json.Unmarshal(iterator.Value(), &stats); err == nil {
			allStats = append(allStats, stats)
		}
	}

	return allStats
}

func (k Keeper) GetTopAgentsByReputation(ctx sdk.Context, limit int) []types.AgentStatistics {
	allAgents := k.GetAllAgents(ctx)

	type agentWithRep struct {
		agent  types.Agent
		stats  types.AgentStatistics
		exists bool
	}

	agentsWithStats := make([]agentWithRep, 0, len(allAgents))
	for _, agent := range allAgents {
		stats, found := k.GetAgentStats(ctx, agent.Address)
		agentsWithStats = append(agentsWithStats, agentWithRep{
			agent:  agent,
			stats:  stats,
			exists: found,
		})
	}

	for i := 0; i < len(agentsWithStats)-1; i++ {
		for j := i + 1; j < len(agentsWithStats); j++ {
			if agentsWithStats[j].agent.Reputation > agentsWithStats[i].agent.Reputation {
				agentsWithStats[i], agentsWithStats[j] = agentsWithStats[j], agentsWithStats[i]
			}
		}
	}

	result := make([]types.AgentStatistics, 0, limit)
	for i := 0; i < len(agentsWithStats) && i < limit; i++ {
		stats := agentsWithStats[i].stats
		stats.Address = agentsWithStats[i].agent.Address
		stats.Reputation = agentsWithStats[i].agent.Reputation
		stats.Status = agentsWithStats[i].agent.Status.String()
		result = append(result, stats)
	}

	return result
}

func (k Keeper) GetTopAgentsBySuccessRate(ctx sdk.Context, minResponses uint64, limit int) []types.AgentStatistics {
	allStats := k.GetAllAgentStats(ctx)

	var filtered []types.AgentStatistics
	for _, stats := range allStats {
		if stats.TotalResponses >= minResponses {
			filtered = append(filtered, stats)
		}
	}

	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].SuccessRate > filtered[i].SuccessRate {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

func (k Keeper) GetAgentsByCapability(ctx sdk.Context, capability string) []types.Agent {
	var matchingAgents []types.Agent

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		for _, cap := range agent.Capabilities {
			if cap == capability {
				matchingAgents = append(matchingAgents, agent)
				return false
			}
		}
		return false
	})

	return matchingAgents
}

func (k Keeper) GetAgentsByCapabilities(ctx sdk.Context, capabilities []string, matchAll bool) []types.Agent {
	var matchingAgents []types.Agent

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		matched := 0
		for _, requiredCap := range capabilities {
			for _, agentCap := range agent.Capabilities {
				if agentCap == requiredCap {
					matched++
					break
				}
			}
		}

		if matchAll && matched == len(capabilities) {
			matchingAgents = append(matchingAgents, agent)
		} else if !matchAll && matched > 0 {
			matchingAgents = append(matchingAgents, agent)
		}
		return false
	})

	return matchingAgents
}

func (k Keeper) GetAgentChallengeHistory(ctx sdk.Context, address string, limit uint64) []types.AIResponse {
	store := ctx.KVStore(k.storeKey)

	var responses []types.AIResponse
	count := uint64(0)

	iterator := storetypes.KVStoreReversePrefixIterator(store, []byte(types.AIResponseKeyPrefix))
	defer iterator.Close()

	for ; iterator.Valid() && count < limit; iterator.Next() {
		var resp types.AIResponse
		k.cdc.MustUnmarshal(iterator.Value(), &resp)
		if resp.ValidatorAddress == address {
			responses = append(responses, resp)
			count++
		}
	}

	return responses
}

func (k Keeper) RecordChallengeMetrics(ctx sdk.Context, epoch uint64, responders []types.AIResponse) {
	store := ctx.KVStore(k.storeKey)

	avgScore := int64(0)
	scoreCount := int64(0)
	minScore := int64(100)
	maxScore := int64(0)
	passCount := int64(0)

	for _, resp := range responders {
		if resp.Score >= 0 {
			avgScore += resp.Score
			scoreCount++
			if resp.Score < minScore {
				minScore = resp.Score
			}
			if resp.Score > maxScore {
				maxScore = resp.Score
			}
			if resp.Score >= 50 {
				passCount++
			}
		}
	}

	metrics := types.ChallengeMetrics{
		Epoch:           epoch,
		TotalResponders: uint64(len(responders)),
		PassCount:       uint64(passCount),
		PassRate:        0,
		AverageScore:    0,
		MinScore:        minScore,
		MaxScore:        maxScore,
	}

	if scoreCount > 0 {
		metrics.AverageScore = avgScore / scoreCount
		metrics.PassRate = float64(passCount) / float64(len(responders))
	}

	bz, _ := json.Marshal(&metrics)
	key := []byte(ChallengeStatsPrefix)
	key = append(key, types.Uint64ToBytes(epoch)...)
	store.Set(key, bz)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"challenge_metrics_recorded",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("responders", fmt.Sprintf("%d", len(responders))),
		sdk.NewAttribute("pass_rate", fmt.Sprintf("%.2f", metrics.PassRate)),
		sdk.NewAttribute("avg_score", fmt.Sprintf("%d", metrics.AverageScore)),
	))
}

func (k Keeper) GetChallengeMetrics(ctx sdk.Context, epoch uint64) (types.ChallengeMetrics, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(append([]byte(ChallengeStatsPrefix), types.Uint64ToBytes(epoch)...))
	if bz == nil {
		return types.ChallengeMetrics{}, false
	}
	var metrics types.ChallengeMetrics
	if err := json.Unmarshal(bz, &metrics); err != nil {
		return types.ChallengeMetrics{}, false
	}
	return metrics, true
}