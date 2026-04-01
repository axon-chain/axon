package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) cleanupOldEpochData(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.KeyChallenge(epoch))

	prefixes := [][]byte{
		types.KeyAIResponsePrefix(epoch),
		append([]byte(types.EpochActivityKeyPrefix), types.Uint64ToBytes(epoch)...),
		append([]byte(types.DeployCountKeyPrefix), types.Uint64ToBytes(epoch)...),
		append([]byte(types.ContractCallKeyPrefix), types.Uint64ToBytes(epoch)...),
	}

	for _, prefix := range prefixes {
		iterator := storetypes.KVStorePrefixIterator(store, prefix)
		var keysToDelete [][]byte
		for ; iterator.Valid(); iterator.Next() {
			keysToDelete = append(keysToDelete, iterator.Key())
		}
		iterator.Close()

		for _, key := range keysToDelete {
			store.Delete(key)
		}
	}
}

func (k Keeper) cleanupEvidenceTxHashes(ctx sdk.Context, height int64) {
	if height < 0 {
		return
	}

	store := ctx.KVStore(k.storeKey)
	prefix := append([]byte(types.EvidenceTxHeightKeyPrefix), types.Uint64ToBytes(uint64(height))...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	var keysToDelete [][]byte
	var evidenceKeys [][]byte
	for ; iterator.Valid(); iterator.Next() {
		key := append([]byte(nil), iterator.Key()...)
		keysToDelete = append(keysToDelete, key)

		hashStart := len(prefix) + 1
		if len(key) > hashStart {
			evidenceKeys = append(evidenceKeys, []byte(types.EvidenceTxKeyPrefix+string(key[hashStart:])))
		}
	}
	iterator.Close()

	for _, key := range evidenceKeys {
		store.Delete(key)
	}
	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// maxDailyRegCleanupPerBlock caps the number of DailyReg entries deleted per
// block to avoid a gas/latency spike on the first run after upgrade, when the
// backlog may contain ~2600 agents × 15 days ≈ 39 000 entries.  The cleanup
// will not advance LastDailyRegCleanupDay until the backlog is fully drained,
// so subsequent blocks continue where the previous one left off.
const maxDailyRegCleanupPerBlock = 5000

// cleanupOldDailyRegData removes DailyReg entries older than cutoffDay.
// Returns true if all expired entries have been removed (i.e. cleanup is
// complete); false if the per-block cap was hit and more work remains.
func (k Keeper) cleanupOldDailyRegData(ctx sdk.Context, currentDay int64) bool {
	if currentDay <= 1 {
		return true
	}

	cutoffDay := currentDay - 1
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, []byte(types.DailyRegKeyPrefix))
	hitCap := false
	var keysToDelete [][]byte
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		if len(key) < len(types.DailyRegKeyPrefix)+1+8 {
			continue
		}

		day := int64(types.BytesToUint64(key[len(key)-8:]))
		if day < cutoffDay {
			keysToDelete = append(keysToDelete, append([]byte(nil), key...))
			if len(keysToDelete) >= maxDailyRegCleanupPerBlock {
				hitCap = true
				break
			}
		}
	}
	iterator.Close()

	for _, key := range keysToDelete {
		store.Delete(key)
	}
	return !hitCap
}
