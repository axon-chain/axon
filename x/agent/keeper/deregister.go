package keeper

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) SetDeregisterRequest(ctx sdk.Context, address string, blockHeight int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyDeregisterQueue(address), types.Uint64ToBytes(uint64(blockHeight)))
}

func (k Keeper) HasDeregisterRequest(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.KeyDeregisterQueue(address))
}

func (k Keeper) GetDeregisterRequest(ctx sdk.Context, address string) (int64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyDeregisterQueue(address))
	if bz == nil || len(bz) < 8 {
		return 0, false
	}
	return int64(types.BytesToUint64(bz)), true
}

func (k Keeper) DeleteDeregisterRequest(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyDeregisterQueue(address))
}

// ProcessDeregisterQueue processes all deregister requests whose cooldown has expired.
func (k Keeper) ProcessDeregisterQueue(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, []byte(types.DeregisterQueueKeyPrefix))
	defer iterator.Close()

	currentBlock := ctx.BlockHeight()
	params := k.GetParams(ctx)

	var toProcess []string

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		if len(bz) < 8 {
			continue
		}
		requestBlock := int64(types.BytesToUint64(bz))
		if currentBlock-requestBlock >= types.DeregisterCooldownBlocks {
			address := string(iterator.Key()[len(types.DeregisterQueueKeyPrefix):])
			toProcess = append(toProcess, address)
		}
	}

	for _, address := range toProcess {
		k.executeDeregister(ctx, address, params)
	}
}

func (k Keeper) executeDeregister(ctx sdk.Context, address string, params types.Params) {
	agent, found := k.GetAgent(ctx, address)
	if !found {
		k.DeleteDeregisterRequest(ctx, address)
		return
	}

	recipientAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		k.Logger(ctx).Error("invalid address in deregister queue", "address", address, "error", err)
		k.DeleteDeregisterRequest(ctx, address)
		return
	}

	// Use snapshot from registration instead of current params
	var burnedAmount sdk.Coin
	if agent.BurnedAtRegister.Denom != "" && agent.BurnedAtRegister.IsPositive() {
		burnedAmount = agent.BurnedAtRegister
	} else {
		burnInt := sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(int64(params.RegisterBurnAmount)), oneAxon))
		burnedAmount = sdk.NewCoin("aaxon", burnInt)
	}
	var moduleHeld sdk.Coin
	if agent.StakeAmount.IsLT(burnedAmount) {
		moduleHeld = sdk.NewInt64Coin("aaxon", 0)
	} else {
		moduleHeld = agent.StakeAmount.Sub(burnedAmount)
	}

	if agent.Reputation == 0 && moduleHeld.IsPositive() {
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(moduleHeld)); err != nil {
			k.Logger(ctx).Error("failed to burn remaining stake — will retry next block", "address", address, "error", err)
			return
		}
	} else if moduleHeld.IsPositive() {
		refundCoins := sdk.NewCoins(moduleHeld)
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, refundCoins); err != nil {
			k.Logger(ctx).Error("failed to refund stake — will retry next block", "address", address, "error", err)
			return
		}
	}

	k.DeleteAgent(ctx, address)
	k.DeleteDeregisterRequest(ctx, address)
	k.DeleteAIBonus(ctx, address)
	k.cleanupAgentEpochData(ctx, address)
	if k.IsV111UpgradeActivated(ctx) && k.privacyKeeper != nil {
		k.privacyKeeper.DeleteAgentIdentity(ctx, address)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_deregistered",
		sdk.NewAttribute("address", address),
		sdk.NewAttribute("module_held", fmt.Sprintf("%s", moduleHeld)),
	))

	k.Logger(ctx).Info("agent deregistered after cooldown", "address", address)
}

// cleanupAgentEpochData removes epoch-scoped data for a deregistered agent.
// Keys use the format "Prefix/<epoch_bytes>/<address>", so we match on the
// "/" + address suffix to avoid accidentally deleting data for other agents
// whose address might be a raw substring of the key.
func (k Keeper) cleanupAgentEpochData(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)

	prefixes := []string{
		types.EpochActivityKeyPrefix,
		types.DeployCountKeyPrefix,
		types.ContractCallKeyPrefix,
		types.AIResponseKeyPrefix,
	}

	suffix := "/" + address

	for _, prefix := range prefixes {
		iterator := storetypes.KVStorePrefixIterator(store, []byte(prefix))
		var toDelete [][]byte
		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			keyStr := string(key)
			if len(keyStr) >= len(suffix) && keyStr[len(keyStr)-len(suffix):] == suffix {
				toDelete = append(toDelete, key)
			}
		}
		iterator.Close()
		for _, key := range toDelete {
			store.Delete(key)
		}
	}
}
