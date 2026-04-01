package keeper

import (
	"encoding/binary"
	"math/big"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

// SetAIBonus stores the AIBonus percentage for a validator using signed encoding.
func (k Keeper) SetAIBonus(ctx sdk.Context, address string, bonus int64) {
	if bonus < types.MinAIBonus {
		bonus = types.MinAIBonus
	}
	if bonus > types.MaxAIBonus {
		bonus = types.MaxAIBonus
	}
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(bonus+128))
	store.Set(types.KeyAIBonus(address), bz)
}

// GetAIBonus returns the AIBonus percentage for a validator.
func (k Keeper) GetAIBonus(ctx sdk.Context, address string) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyAIBonus(address))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	raw := binary.BigEndian.Uint64(bz)
	// New offset encoding: stored = bonus + 128, valid range [123, 158]
	// Legacy raw int64-as-uint64: positive values 0-30, negative values >= 2^63
	const maxNewEncoded = uint64(types.MaxAIBonus + 128) // 158
	if raw <= maxNewEncoded {
		result := int64(raw) - 128
		if result < types.MinAIBonus {
			return types.MinAIBonus
		}
		if result > types.MaxAIBonus {
			return types.MaxAIBonus
		}
		return result
	}
	// Legacy or corrupted data — clamp to valid range
	legacy := int64(raw)
	if legacy < types.MinAIBonus {
		return types.MinAIBonus
	}
	if legacy > types.MaxAIBonus {
		return types.MaxAIBonus
	}
	return legacy
}

// DeleteAIBonus removes the AIBonus entry for an address.
func (k Keeper) DeleteAIBonus(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyAIBonus(address))
}

// IncrementEpochActivity increments the transaction count for an agent in the current epoch.
func (k Keeper) IncrementEpochActivity(ctx sdk.Context, address string) {
	epoch := k.GetCurrentEpoch(ctx)
	store := ctx.KVStore(k.storeKey)
	key := types.KeyEpochActivity(epoch, address)

	count := uint64(0)
	bz := store.Get(key)
	if bz != nil && len(bz) >= 8 {
		count = binary.BigEndian.Uint64(bz)
	}
	count++

	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(key, bz)
}

// GetEpochActivity returns the transaction count for an agent in a given epoch.
func (k Keeper) GetEpochActivity(ctx sdk.Context, epoch uint64, address string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyEpochActivity(epoch, address))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// ProcessEpochReputation updates reputation for all agents at epoch boundaries.
// Collects updates first, then applies them to avoid mutating state during iteration.
func (k Keeper) ProcessEpochReputation(ctx sdk.Context, epoch uint64) {
	params := k.GetParams(ctx)

	type reputationUpdate struct {
		address string
		delta   int64
	}

	var updates []reputationUpdate

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
			return false
		}
		if agent.Reputation == 0 {
			return false
		}

		delta := int64(0)

		if agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE {
			delta += types.ReputationGainEpochOnline
		}

		if agent.Status == types.AgentStatus_AGENT_STATUS_OFFLINE {
			delta += types.ReputationLossNoHeartbeatEpoch
		}

		activity := k.GetEpochActivity(ctx, epoch, agent.Address)
		if activity >= 10 {
			delta += types.ReputationGainActiveEpoch
		}

		if delta != 0 {
			updates = append(updates, reputationUpdate{address: agent.Address, delta: delta})
		}

		return false
	})

	// Apply updates outside of iteration
	for _, u := range updates {
		k.UpdateReputation(ctx, u.address, u.delta)

		updatedAgent, found := k.GetAgent(ctx, u.address)
		if found && updatedAgent.Reputation == 0 {
			k.handleZeroReputation(ctx, updatedAgent, params)
		}
	}
}

// handleZeroReputation burns remaining stake and suspends the agent.
// Uses the snapshot BurnedAtRegister instead of current params to avoid parameter-change drift.
func (k Keeper) handleZeroReputation(ctx sdk.Context, agent types.Agent, params types.Params) {
	var burnedAtRegister sdk.Coin
	if agent.BurnedAtRegister.Denom != "" && agent.BurnedAtRegister.IsPositive() {
		burnedAtRegister = agent.BurnedAtRegister
	} else {
		// Fallback for legacy agents without snapshot
		burnInt := new(big.Int).Mul(big.NewInt(int64(params.RegisterBurnAmount)), oneAxon)
		burnedAtRegister = sdk.NewCoin("aaxon", sdkmath.NewIntFromBigInt(burnInt))
	}

	var remaining sdk.Coin
	if agent.StakeAmount.IsLT(burnedAtRegister) {
		remaining = sdk.NewInt64Coin("aaxon", 0)
	} else {
		remaining = agent.StakeAmount.Sub(burnedAtRegister)
	}

	burned := false
	if remaining.IsPositive() {
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(remaining)); err != nil {
			k.Logger(ctx).Error("failed to burn stake for zero reputation — suspending without burn", "address", agent.Address, "error", err)
		} else {
			burned = true
		}
	}

	agent.Status = types.AgentStatus_AGENT_STATUS_SUSPENDED
	if burned || !remaining.IsPositive() {
		agent.StakeAmount = sdk.NewInt64Coin("aaxon", 0)
	}
	k.SetAgent(ctx, agent)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_stake_burned",
		sdk.NewAttribute("address", agent.Address),
		sdk.NewAttribute("burned", remaining.String()),
		sdk.NewAttribute("reason", "reputation_zero"),
	))
}

// oneAxon = 10^18
var oneAxon = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

// Daily registration rate limiting

const dailyBlockWindow int64 = 17280 // ~24h at 5s/block

func dailyRegisterKey(address string, day int64) []byte {
	return []byte(types.DailyRegKeyPrefix + address + "/" + string(types.Uint64ToBytes(uint64(day))))
}

func (k Keeper) GetDailyRegisterCount(ctx sdk.Context, address string) uint64 {
	day := ctx.BlockHeight() / dailyBlockWindow
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(dailyRegisterKey(address, day))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) ExportAIBonuses(ctx sdk.Context) map[string]int64 {
	result := make(map[string]int64)
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.AIBonusKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(prefix):])
		bonus := k.GetAIBonus(ctx, address)
		result[address] = bonus
	}
	return result
}

func (k Keeper) ImportAIBonuses(ctx sdk.Context, bonuses map[string]int64) {
	for address, bonus := range bonuses {
		k.SetAIBonus(ctx, address, bonus)
	}
}

func (k Keeper) IncrementDailyRegisterCount(ctx sdk.Context, address string) {
	day := ctx.BlockHeight() / dailyBlockWindow
	store := ctx.KVStore(k.storeKey)
	key := dailyRegisterKey(address, day)
	count := uint64(0)
	bz := store.Get(key)
	if bz != nil && len(bz) >= 8 {
		count = binary.BigEndian.Uint64(bz)
	}
	count++
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(key, bz)
}
