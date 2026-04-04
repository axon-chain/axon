package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/privacy/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	bankKeeper types.BankKeeper
}

const agentIdentityCommitmentLength = 32

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bk types.BankKeeper) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, bankKeeper: bk}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// --- Commitment Tree ---

func (k Keeper) GetTreeSize(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte{types.CommitmentSizeKey})
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) setTreeSize(ctx sdk.Context, size uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, size)
	store.Set([]byte{types.CommitmentSizeKey}, bz)
}

func (k Keeper) InsertCommitment(ctx sdk.Context, commitment []byte) (uint64, error) {
	size := k.GetTreeSize(ctx)
	maxSize := uint64(1) << types.TreeDepth
	if size >= maxSize {
		return 0, types.ErrTreeFull
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.CommitmentKey(size), commitment)
	size++
	k.setTreeSize(ctx, size)

	newRoot := k.computeMerkleRoot(ctx, size)
	k.setCurrentRoot(ctx, newRoot)
	k.addHistoricalRoot(ctx, newRoot)

	return size - 1, nil
}

func (k Keeper) GetCurrentRoot(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get([]byte{types.CommitmentRootKey})
}

func (k Keeper) setCurrentRoot(ctx sdk.Context, root []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte{types.CommitmentRootKey}, root)
}

func (k Keeper) IsKnownRoot(ctx sdk.Context, root []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.HistoricalRootKey(root))
}

// addHistoricalRoot stores a root and evicts the oldest if over capacity (FIFO).
// Uses a monotonic counter; O(1) eviction per insert.
func (k Keeper) addHistoricalRoot(ctx sdk.Context, root []byte) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.HistoricalRootKey(root), []byte{1})

	head := k.getHistoricalRootCounter(ctx)
	store.Set(types.HistoricalRootIndexKey(head), root)
	k.setHistoricalRootCounter(ctx, head+1)

	params := k.GetParams(ctx)
	maxRoots := params.MaxKnownRoots
	if maxRoots == 0 {
		maxRoots = 100
	}

	if head+1 > maxRoots {
		evictIdx := head + 1 - maxRoots - 1
		idxKey := types.HistoricalRootIndexKey(evictIdx)
		oldRoot := store.Get(idxKey)
		if oldRoot != nil {
			store.Delete(types.HistoricalRootKey(oldRoot))
			store.Delete(idxKey)
		}
	}
}

func (k Keeper) getHistoricalRootCounter(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte{types.HistoricalRootCounterKey})
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) setHistoricalRootCounter(ctx sdk.Context, counter uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, counter)
	store.Set([]byte{types.HistoricalRootCounterKey}, bz)
}

// computeMerkleRoot computes the root of a fixed-depth incremental Merkle tree.
// Uses a frontier-based approach: only O(TreeDepth) nodes are hashed per insert,
// not the full tree. The frontier stores the latest left-child at each level.
func (k Keeper) computeMerkleRoot(ctx sdk.Context, size uint64) []byte {
	if size == 0 {
		return zeroHashes[types.TreeDepth]
	}

	store := ctx.KVStore(k.storeKey)

	frontier := k.loadFrontier(ctx)

	lastIdx := size - 1
	leaf := store.Get(types.CommitmentKey(lastIdx))
	if leaf == nil {
		leaf = zeroHashes[0]
	}

	current := leaf
	idx := lastIdx
	for depth := 0; depth < types.TreeDepth; depth++ {
		if idx%2 == 1 {
			current = PoseidonHash2(frontier[depth], current)
		} else {
			frontier[depth] = current
			current = PoseidonHash2(current, zeroHashes[depth])
		}
		idx /= 2
	}

	k.saveFrontier(ctx, frontier)
	return current
}

// computeFullRoot recomputes the root from scratch (used for verification/genesis only).
func (k Keeper) computeFullRoot(ctx sdk.Context, size uint64) []byte {
	if size == 0 {
		return zeroHashes[types.TreeDepth]
	}

	store := ctx.KVStore(k.storeKey)
	currentLevel := make([][]byte, size)
	for i := uint64(0); i < size; i++ {
		leaf := store.Get(types.CommitmentKey(i))
		if leaf == nil {
			leaf = zeroHashes[0]
		}
		currentLevel[i] = leaf
	}

	for depth := 0; depth < types.TreeDepth; depth++ {
		n := uint64(len(currentLevel))
		nextSize := (n + 1) / 2
		nextLevel := make([][]byte, nextSize)
		for i := uint64(0); i < nextSize; i++ {
			left := currentLevel[i*2]
			var right []byte
			if i*2+1 < n {
				right = currentLevel[i*2+1]
			} else {
				right = zeroHashes[depth]
			}
			nextLevel[i] = PoseidonHash2(left, right)
		}
		currentLevel = nextLevel
	}

	if len(currentLevel) > 0 {
		return currentLevel[0]
	}
	return zeroHashes[types.TreeDepth]
}

const frontierKeyPrefix = "MerkleFrontier/"

func (k Keeper) loadFrontier(ctx sdk.Context) [types.TreeDepth][]byte {
	store := ctx.KVStore(k.storeKey)
	var frontier [types.TreeDepth][]byte
	for i := 0; i < types.TreeDepth; i++ {
		key := []byte(fmt.Sprintf("%s%d", frontierKeyPrefix, i))
		bz := store.Get(key)
		if bz != nil {
			frontier[i] = bz
		} else {
			frontier[i] = zeroHashes[i]
		}
	}
	return frontier
}

func (k Keeper) saveFrontier(ctx sdk.Context, frontier [types.TreeDepth][]byte) {
	store := ctx.KVStore(k.storeKey)
	for i := 0; i < types.TreeDepth; i++ {
		key := []byte(fmt.Sprintf("%s%d", frontierKeyPrefix, i))
		store.Set(key, frontier[i])
	}
}

// zeroHashes caches zero-value hashes for each tree depth level.
var zeroHashes [types.TreeDepth + 1][]byte

func init() {
	zeroHashes[0] = make([]byte, 32) // zero leaf
	for i := 1; i <= types.TreeDepth; i++ {
		zeroHashes[i] = PoseidonHash2(zeroHashes[i-1], zeroHashes[i-1])
	}
}

// --- Nullifier Set ---

func (k Keeper) IsNullifierSpent(ctx sdk.Context, nullifier []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.NullifierKey(nullifier))
}

func (k Keeper) MarkNullifierSpent(ctx sdk.Context, nullifier []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NullifierKey(nullifier), []byte{1})
}

// --- Shielded Pool ---

func (k Keeper) GetShieldedBalance(ctx sdk.Context) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte{types.ShieldedBalanceKey})
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	var bal sdkmath.Int
	if err := bal.Unmarshal(bz); err != nil {
		return sdkmath.ZeroInt()
	}
	return bal
}

func (k Keeper) setShieldedBalance(ctx sdk.Context, bal sdkmath.Int) error {
	bz, err := bal.Marshal()
	if err != nil {
		return fmt.Errorf("marshal shielded balance: %w", err)
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte{types.ShieldedBalanceKey}, bz)
	return nil
}

func (k Keeper) AddToShieldedPool(ctx sdk.Context, amount sdkmath.Int) error {
	bal := k.GetShieldedBalance(ctx)
	return k.setShieldedBalance(ctx, bal.Add(amount))
}

func (k Keeper) SubFromShieldedPool(ctx sdk.Context, amount sdkmath.Int) error {
	bal := k.GetShieldedBalance(ctx)
	if bal.LT(amount) {
		return types.ErrInsufficientPool
	}
	return k.setShieldedBalance(ctx, bal.Sub(amount))
}

// --- Identity Commitments ---

func (k Keeper) IsIdentityRegistered(ctx sdk.Context, commitment []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.IdentityKey(commitment))
}

func (k Keeper) RegisterIdentity(ctx sdk.Context, commitment []byte) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	store.Set(types.IdentityKey(commitment), bz)
}

func (k Keeper) HasAgentIdentity(ctx sdk.Context, agentAddr string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.AgentIdentityKey(agentAddr))
}

func (k Keeper) GetAgentIdentityCommitment(ctx sdk.Context, agentAddr string) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AgentIdentityKey(agentAddr))
	if len(bz) != agentIdentityCommitmentLength {
		return nil, false
	}
	commitment := make([]byte, len(bz))
	copy(commitment, bz)
	return commitment, true
}

func (k Keeper) SetAgentIdentity(ctx sdk.Context, agentAddr string, commitment []byte) {
	store := ctx.KVStore(k.storeKey)
	if len(commitment) == agentIdentityCommitmentLength {
		stored := make([]byte, len(commitment))
		copy(stored, commitment)
		store.Set(types.AgentIdentityKey(agentAddr), stored)
		return
	}
	store.Set(types.AgentIdentityKey(agentAddr), []byte{1})
}

// DeleteAgentIdentity removes the agent -> identity index and, when available,
// also deletes the registered commitment. Legacy one-byte marker values are
// still deleted from the agent index but do not carry enough information to
// remove the original commitment entry.
func (k Keeper) DeleteAgentIdentity(ctx sdk.Context, agentAddr string) {
	store := ctx.KVStore(k.storeKey)
	key := types.AgentIdentityKey(agentAddr)
	if commitment, ok := k.GetAgentIdentityCommitment(ctx, agentAddr); ok {
		store.Delete(types.IdentityKey(commitment))
	}
	store.Delete(key)
}

// --- Verifying Keys ---

func (k Keeper) IsVKRegistered(ctx sdk.Context, keyId []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.VerifyingKeyKey(keyId))
}

func (k Keeper) GetVK(ctx sdk.Context, keyId []byte) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(types.VerifyingKeyKey(keyId))
}

func (k Keeper) RegisterVK(ctx sdk.Context, keyId, vk []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.VerifyingKeyKey(keyId), vk)
}

// --- Params ---

const paramsKey = "PrivacyParams"

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(paramsKey))
	if bz == nil {
		return types.DefaultParams()
	}
	var p types.Params
	if err := json.Unmarshal(bz, &p); err != nil {
		return types.DefaultParams()
	}
	return p
}

func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	bz, err := json.Marshal(&p)
	if err != nil {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(paramsKey), bz)
}

// PoseidonHash2 computes MiMC(left, right) over BN254 Fr using gnark-crypto.
// MiMC is the standard SNARK-friendly hash in the gnark ecosystem, fully
// compatible with in-circuit MiMC gadgets for Merkle proof verification.
func PoseidonHash2(left, right []byte) []byte {
	h := mimc.NewMiMC()
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}

// PoseidonHash3 computes MiMC(a, b, c) over BN254 Fr.
func PoseidonHash3(a, b, c []byte) []byte {
	h := mimc.NewMiMC()
	h.Write(a)
	h.Write(b)
	h.Write(c)
	return h.Sum(nil)
}
