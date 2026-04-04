package keeper

import (
	"encoding/binary"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

// Scores are stored as milliscored int64 values (1.5 → 1500) for determinism.
const (
	L1RepKeyPrefix = "L1Rep/"
	L2RepKeyPrefix = "L2Rep/"

	L1MaxMilliscore int64 = 40_000  // 40.0
	L2MaxMilliscore int64 = 30_000  // 30.0
	TotalMaxRep     int64 = 100_000 // 100.0

	// L1 scoring millis per epoch
	scoreSignRateHigh    int64 = 1000    // +1.0 if sign rate > 95%
	scoreSignRateMedium  int64 = 500     // +0.5 if sign rate 80%-95%
	scoreHeartbeatActive int64 = 300     // +0.3 if at least 1 heartbeat this epoch
	scoreOnChainActive   int64 = 500     // +0.5 if >= 10 txs this epoch
	scoreContractUsage   int64 = 500     // +0.5 if deployed contract called by >= 5 addrs
	scoreAIChallengeTop  int64 = 2000    // +2.0 if AI score top 20%
	scoreAIChallengeGood int64 = 1000    // +1.0 if AI score top 50%
	scoreAIChallengePoor int64 = -1000   // -1.0 if bottom 20% or cheater
	scoreOfflineImm      int64 = -5000   // -5.0 immediate on going offline
	scoreDoubleSignImm   int64 = -40_000 // reset to 0

	decayL1Millis int64 = 100 // -0.1 per epoch
	decayL2Millis int64 = 50  // -0.05 per epoch

	minTxsForActive    uint64 = 10
	minCallersForUsage uint64 = 5
)

func l1RepKey(address string) []byte {
	return []byte(L1RepKeyPrefix + address)
}

func l2RepKey(address string) []byte {
	return []byte(L2RepKeyPrefix + address)
}

func (k Keeper) GetL1Score(ctx sdk.Context, address string) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(l1RepKey(address))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetL1Score(ctx sdk.Context, address string, score int64) {
	if score < 0 {
		score = 0
	}
	if score > k.l1CapMillis(ctx) {
		score = k.l1CapMillis(ctx)
	}
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(score))
	store.Set(l1RepKey(address), bz)
}

func (k Keeper) GetL2Score(ctx sdk.Context, address string) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(l2RepKey(address))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetL2Score(ctx sdk.Context, address string, score int64) {
	if score < 0 {
		score = 0
	}
	if score > k.l2CapMillis(ctx) {
		score = k.l2CapMillis(ctx)
	}
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(score))
	store.Set(l2RepKey(address), bz)
}

func (k Keeper) l1CapMillis(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	if params.L1Cap > 0 {
		return int64(params.L1Cap) * 1000
	}
	return L1MaxMilliscore
}

func (k Keeper) l2CapMillis(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	if params.L2Cap > 0 {
		return int64(params.L2Cap) * 1000
	}
	return L2MaxMilliscore
}

// GetTotalReputation returns the combined L1+L2 score in milliscored units.
func (k Keeper) GetTotalReputation(ctx sdk.Context, address string) int64 {
	total := k.GetL1Score(ctx, address) + k.GetL2Score(ctx, address)
	if total > TotalMaxRep {
		total = TotalMaxRep
	}
	if total < 0 {
		total = 0
	}
	return total
}

// ProcessL1Reputation is called at epoch end. It evaluates all agents and
// adjusts L1 scores based on on-chain behavior during the epoch.
func (k Keeper) ProcessL1Reputation(ctx sdk.Context, epoch uint64) {
	params := k.GetParams(ctx)

	aiScores := k.collectAIScores(ctx, epoch)
	aiPercentiles := computePercentiles(aiScores)

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
			return false
		}

		delta := int64(0)
		addr := agent.Address

		// Validator signing rate
		if k.isActiveValidatorAddress(ctx, addr) {
			signRate := k.getValidatorSignRate(ctx, addr, params.EpochLength)
			switch {
			case signRate > 95:
				delta += scoreSignRateHigh
			case signRate >= 80:
				delta += scoreSignRateMedium
			}
		}

		// Heartbeat activity
		activity := k.GetEpochActivity(ctx, epoch, addr)
		if activity > 0 {
			delta += scoreHeartbeatActive
		}

		// On-chain transaction activity
		if activity >= minTxsForActive {
			delta += scoreOnChainActive
		}

		// Contract usage (deployed contract called by >= 5 unique callers)
		calls := k.getCounter(ctx, types.KeyContractCall(epoch, addr))
		if calls >= minCallersForUsage {
			delta += scoreContractUsage
		}

		// AI challenge performance (percentile 100=best).
		// Cheaters and bottom 20% share the same penalty slot (OR, not AND)
		// to avoid double penalizing.
		isCheater := k.isEpochCheater(ctx, epoch, addr)
		if isCheater {
			delta += scoreAIChallengePoor
		} else if p, ok := aiPercentiles[addr]; ok {
			switch {
			case p <= 20:
				delta += scoreAIChallengePoor // bottom 20%
			case p > 80:
				delta += scoreAIChallengeTop // top 20%
			case p > 50:
				delta += scoreAIChallengeGood // top 50% (51-80)
			}
		}

		current := k.GetL1Score(ctx, addr)
		k.SetL1Score(ctx, addr, current+delta)

		return false
	})
}

// ApplyReputationDecay applies natural decay to both L1 and L2 scores.
// Reads decay rates from governance params; falls back to constants.
func (k Keeper) ApplyReputationDecay(ctx sdk.Context) {
	params := k.GetParams(ctx)

	l1Decay := decayL1Millis
	if params.L1DecayPerEpoch != "" {
		if v, err := parseDecToMillis(params.L1DecayPerEpoch); err == nil {
			l1Decay = v
		}
	}
	l2Decay := decayL2Millis
	if params.L2DecayPerEpoch != "" {
		if v, err := parseDecToMillis(params.L2DecayPerEpoch); err == nil {
			l2Decay = v
		}
	}

	l1Cap := k.l1CapMillis(ctx)
	l2Cap := k.l2CapMillis(ctx)

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		addr := agent.Address
		if k.shouldFreezeAgentReputationDuringDeregister(ctx, addr) {
			return false
		}

		l1 := k.GetL1Score(ctx, addr)
		if l1 > l1Cap {
			l1 = l1Cap
		}
		if l1 > 0 {
			l1 -= l1Decay
			if l1 < 0 {
				l1 = 0
			}
			k.SetL1Score(ctx, addr, l1)
		}

		l2 := k.GetL2Score(ctx, addr)
		if l2 > l2Cap {
			l2 = l2Cap
		}
		if l2 > 0 {
			l2 -= l2Decay
			if l2 < 0 {
				l2 = 0
			}
			k.SetL2Score(ctx, addr, l2)
		}

		return false
	})
}

func parseDecToMillis(s string) (int64, error) {
	d, err := math.LegacyNewDecFromStr(s)
	if err != nil {
		return 0, err
	}
	return d.MulInt64(1000).TruncateInt64(), nil
}

// SyncLegacyReputation updates the old agent.Reputation field from L1+L2 for backward compatibility.
func (k Keeper) SyncLegacyReputation(ctx sdk.Context) {
	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if k.shouldFreezeAgentReputationDuringDeregister(ctx, agent.Address) {
			return false
		}

		total := k.GetTotalReputation(ctx, agent.Address)
		legacy := uint64(total / 1000)
		if legacy > 100 {
			legacy = 100
		}
		if agent.Reputation != legacy {
			agent.Reputation = legacy
			k.SetAgent(ctx, agent)
		}
		return false
	})
}

// BootstrapLegacyReputation maps the legacy total reputation field into the
// dual-layer L1/L2 score stores when an agent is first created or imported.
func (k Keeper) BootstrapLegacyReputation(ctx sdk.Context, address string, reputation uint64) {
	remaining := int64(reputation) * 1000

	l1 := remaining
	if cap := k.l1CapMillis(ctx); l1 > cap {
		l1 = cap
	}
	remaining -= l1

	l2 := remaining
	if cap := k.l2CapMillis(ctx); l2 > cap {
		l2 = cap
	}

	k.SetL1Score(ctx, address, l1)
	k.SetL2Score(ctx, address, l2)
}

// ApplyOfflinePenalty applies immediate L1 penalty when agent goes offline.
func (k Keeper) ApplyOfflinePenalty(ctx sdk.Context, address string) {
	current := k.GetL1Score(ctx, address)
	k.SetL1Score(ctx, address, current+scoreOfflineImm)
}

// ApplyDoubleSignPenalty resets L1 to 0 on double-sign.
func (k Keeper) ApplyDoubleSignPenalty(ctx sdk.Context, address string) {
	k.SetL1Score(ctx, address, 0)
}

// getValidatorSignRate returns the signing percentage (0-100) for a validator in the current epoch.
// Uses the last commit info from the block header.
func (k Keeper) getValidatorSignRate(ctx sdk.Context, address string, epochLength uint64) int64 {
	// Approximate: use epoch activity as a proxy for signing participation.
	// A full implementation would track per-block LastCommitInfo.Votes in BeginBlocker.
	// For now, use heartbeat count × (epochLength / heartbeat_interval) as proxy.
	epoch := k.GetCurrentEpoch(ctx)
	activity := k.GetEpochActivity(ctx, epoch, address)
	if epochLength == 0 {
		return 0
	}
	// Each heartbeat represents one "alive" signal per HeartbeatInterval blocks.
	// Full epoch participation = epochLength / HeartbeatInterval heartbeats.
	rate := int64(activity) * 100 / int64(epochLength)
	if rate > 100 {
		rate = 100
	}
	return rate
}

// isEpochCheater checks if an agent was flagged as a cheater in the given epoch.
func (k Keeper) isEpochCheater(ctx sdk.Context, epoch uint64, address string) bool {
	responses := k.GetEpochResponses(ctx, epoch)
	for _, r := range responses {
		if r.ValidatorAddress == address && r.Evaluated && r.Score < 0 {
			return true
		}
	}
	return false
}

type aiScoreEntry struct {
	address string
	score   int64
}

func (k Keeper) collectAIScores(ctx sdk.Context, epoch uint64) []aiScoreEntry {
	responses := k.GetEpochResponses(ctx, epoch)
	var entries []aiScoreEntry
	for _, r := range responses {
		if r.Evaluated && r.Score > 0 {
			entries = append(entries, aiScoreEntry{address: r.ValidatorAddress, score: r.Score})
		}
	}
	return entries
}

// computePercentiles returns map[address] → percentile (0-100, 100=best).
func computePercentiles(entries []aiScoreEntry) map[string]int64 {
	if len(entries) == 0 {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].score < entries[j].score
	})
	result := make(map[string]int64, len(entries))
	n := int64(len(entries))
	for i, e := range entries {
		percentile := (int64(i) + 1) * 100 / n
		result[e.address] = percentile
	}
	return result
}
