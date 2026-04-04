package keeper

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/log/v2"
	store "cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/axon-chain/axon/x/privacy/types"
)

func newPrivacyTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore := store.NewCommitMultiStore(dbm.NewMemDB(), log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("load latest version: %v", err)
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	k := NewKeeper(cdc, storeKey, nil)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1}, false, log.NewNopLogger())
	return k, ctx
}

func TestDeleteAgentIdentityRemovesCommitmentForNewFormat(t *testing.T) {
	k, ctx := newPrivacyTestKeeper(t)

	commitment := make([]byte, 32)
	for i := range commitment {
		commitment[i] = byte(i + 1)
	}

	k.RegisterIdentity(ctx, commitment)
	k.SetAgentIdentity(ctx, "agent1", commitment)
	k.DeleteAgentIdentity(ctx, "agent1")

	if k.HasAgentIdentity(ctx, "agent1") {
		t.Fatal("agent identity index should be deleted")
	}
	if k.IsIdentityRegistered(ctx, commitment) {
		t.Fatal("commitment should be deleted for new-format agent identity")
	}
}

func TestDeleteAgentIdentityKeepsLegacyCommitmentWithoutReverseIndex(t *testing.T) {
	k, ctx := newPrivacyTestKeeper(t)

	commitment := make([]byte, 32)
	for i := range commitment {
		commitment[i] = byte(255 - i)
	}

	k.RegisterIdentity(ctx, commitment)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.AgentIdentityKey("agent1"), []byte{1})

	k.DeleteAgentIdentity(ctx, "agent1")

	if k.HasAgentIdentity(ctx, "agent1") {
		t.Fatal("legacy agent identity marker should be deleted")
	}
	if !k.IsIdentityRegistered(ctx, commitment) {
		t.Fatal("legacy commitment should remain because reverse index is unavailable")
	}
}
