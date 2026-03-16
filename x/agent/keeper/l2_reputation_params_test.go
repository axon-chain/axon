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

	"github.com/axon-chain/axon/x/agent/types"
)

func newL2ReputationTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore := store.NewCommitMultiStore(dbm.NewMemDB(), log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("load latest version: %v", err)
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	k := NewKeeper(cdc, storeKey, nil, nil)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 720}, false, log.NewNopLogger())
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("set default params: %v", err)
	}
	return k, ctx
}

func setTestAgent(k Keeper, ctx sdk.Context, address string, registeredAt int64) {
	k.SetAgent(ctx, types.Agent{
		Address:          address,
		AgentId:          address,
		Status:           types.AgentStatus_AGENT_STATUS_ONLINE,
		StakeAmount:      sdk.NewInt64Coin("aaxon", 1),
		BurnedAtRegister: sdk.NewInt64Coin("aaxon", 0),
		RegisteredAt:     registeredAt,
		LastHeartbeat:    registeredAt,
		Capabilities:     []string{"test"},
		Model:            "test-model",
	})
}

func TestSubmitL2ReportUsesConfiguredThresholdsAndWeights(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L2MinReporterRep = 10
	params.L2MinAccountAge = 0
	params.L2NoEvidenceWeight = "0"
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	setTestAgent(k, ctx, "reporter", ctx.BlockHeight())
	setTestAgent(k, ctx, "target", 1)
	k.SetL1Score(ctx, "reporter", 10_000)

	if err := k.SubmitL2Report(ctx, "reporter", "target", 1, "", "no evidence"); err != nil {
		t.Fatalf("submit report: %v", err)
	}

	details := k.GetL2ReportDetails(ctx, k.GetCurrentEpoch(ctx), "target")
	if len(details) != 1 {
		t.Fatalf("expected 1 stored report, got %d", len(details))
	}
	if details[0].Weight != 0 {
		t.Fatalf("expected zero weight from configured no-evidence weight, got %d", details[0].Weight)
	}
}

func TestSubmitL2ReportRejectsBelowConfiguredMinReporterRep(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L2MinReporterRep = 35
	params.L2MinAccountAge = 0
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	setTestAgent(k, ctx, "reporter", 1)
	setTestAgent(k, ctx, "target", 1)
	k.SetL1Score(ctx, "reporter", 34_000)

	if err := k.SubmitL2Report(ctx, "reporter", "target", 1, "", "too low"); err == nil || err.Error() != "reputation too low" {
		t.Fatalf("expected reputation too low, got %v", err)
	}
}

func TestSubmitL2ReportRejectsTooNewUsingConfiguredAccountAge(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	ctx = ctx.WithBlockHeight(104)

	params := types.DefaultParams()
	params.L2MinReporterRep = 10
	params.L2MinAccountAge = 5
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	setTestAgent(k, ctx, "reporter", 100)
	setTestAgent(k, ctx, "target", 1)
	k.SetL1Score(ctx, "reporter", 10_000)

	if err := k.SubmitL2Report(ctx, "reporter", "target", 1, "", "too new"); err == nil || err.Error() != "account too new" {
		t.Fatalf("expected account too new, got %v", err)
	}
}

func TestComputeRawL2DeltasUsesConfiguredAbuseThreshold(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L2AbuseThreshold = 10
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	for i := 0; i < 11; i++ {
		k.storeL2Report(ctx, L2Report{
			Reporter: "spammer",
			Target:   sdk.AccAddress([]byte{byte(i + 1)}).String(),
			Score:    1,
			Weight:   1000,
			Epoch:    1,
		})
	}

	raw := k.computeRawL2Deltas(ctx, 1)
	if len(raw) != 0 {
		t.Fatalf("expected spammer reports to be zeroed by configured abuse threshold, got %+v", raw)
	}
}

func TestComputeRawL2DeltasUsesConfiguredMutualPenalty(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L2AbuseThreshold = 100
	params.L2MutualReportPenalty = "0.5"
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	k.storeL2Report(ctx, L2Report{Reporter: "alice", Target: "bob", Score: 1, Weight: 1000, Epoch: 1})
	k.storeL2Report(ctx, L2Report{Reporter: "bob", Target: "alice", Score: 1, Weight: 1000, Epoch: 1})
	k.storeL2Report(ctx, L2Report{Reporter: "carol", Target: "bob", Score: -1, Weight: 1000, Epoch: 1})

	raw := k.computeRawL2Deltas(ctx, 1)
	if got := raw["bob"]; got != -333 {
		t.Fatalf("expected bob delta -333 with configured mutual penalty, got %d", got)
	}
}

func TestBootstrapLegacyReputationSeedsDualLayerScores(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L1Cap = 40
	params.L2Cap = 30
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	k.BootstrapLegacyReputation(ctx, "agent", 50)

	if got := k.GetL1Score(ctx, "agent"); got != 40_000 {
		t.Fatalf("expected l1 score 40000, got %d", got)
	}
	if got := k.GetL2Score(ctx, "agent"); got != 10_000 {
		t.Fatalf("expected l2 score 10000, got %d", got)
	}
	if got := k.GetTotalReputation(ctx, "agent"); got != 50_000 {
		t.Fatalf("expected total reputation 50000, got %d", got)
	}
}

func TestApplyReputationDecayRespectsZeroParams(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L1DecayPerEpoch = "0"
	params.L2DecayPerEpoch = "0"
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	k.SetL1Score(ctx, "agent", 10_000)
	k.SetL2Score(ctx, "agent", 5_000)

	k.ApplyReputationDecay(ctx)

	if got := k.GetL1Score(ctx, "agent"); got != 10_000 {
		t.Fatalf("expected l1 score to remain 10000, got %d", got)
	}
	if got := k.GetL2Score(ctx, "agent"); got != 5_000 {
		t.Fatalf("expected l2 score to remain 5000, got %d", got)
	}
}
