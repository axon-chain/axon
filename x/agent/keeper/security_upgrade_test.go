package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axon-chain/axon/x/agent/types"
)

func TestSubmitL2ReportUsesFullWeightOnlyForIndexedEvidence(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L2MinReporterRep = 10
	params.L2MinAccountAge = 0
	params.L2NoEvidenceWeight = "0.3"
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	setTestAgent(k, ctx, "reporter", 1)
	setTestAgent(k, ctx, "target", 1)
	k.SetL1Score(ctx, "reporter", 10_000)

	evidenceHash := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
	k.RecordEvidenceTxHash(ctx, evidenceHash)

	if err := k.SubmitL2Report(ctx, "reporter", "target", 1, evidenceHash.Hex(), "indexed"); err != nil {
		t.Fatalf("submit report: %v", err)
	}

	details := k.GetL2ReportDetails(ctx, k.GetCurrentEpoch(ctx), "target")
	if len(details) != 1 {
		t.Fatalf("expected 1 stored report, got %d", len(details))
	}
	if details[0].Weight != 316 {
		t.Fatalf("expected full evidence weight 316, got %d", details[0].Weight)
	}
}

func TestSubmitL2ReportFallsBackWhenEvidenceMissingOrInvalid(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	params := types.DefaultParams()
	params.L2MinReporterRep = 10
	params.L2MinAccountAge = 0
	params.L2NoEvidenceWeight = "0.3"
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	setTestAgent(k, ctx, "reporter1", 1)
	setTestAgent(k, ctx, "target1", 1)
	k.SetL1Score(ctx, "reporter1", 10_000)
	if err := k.SubmitL2Report(ctx, "reporter1", "target1", 1, "xyz", "invalid"); err != nil {
		t.Fatalf("submit invalid evidence report: %v", err)
	}

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 720)
	setTestAgent(k, ctx, "reporter2", 1)
	setTestAgent(k, ctx, "target2", 1)
	k.SetL1Score(ctx, "reporter2", 10_000)
	missingHash := common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")
	if err := k.SubmitL2Report(ctx, "reporter2", "target2", 1, missingHash.Hex(), "missing"); err != nil {
		t.Fatalf("submit missing evidence report: %v", err)
	}

	details1 := k.GetL2ReportDetails(ctx.WithBlockHeight(720), 1, "target1")
	if len(details1) != 1 || details1[0].Weight != 94 {
		t.Fatalf("expected invalid evidence weight 94, got %+v", details1)
	}

	details2 := k.GetL2ReportDetails(ctx, k.GetCurrentEpoch(ctx), "target2")
	if len(details2) != 1 || details2[0].Weight != 94 {
		t.Fatalf("expected missing evidence weight 94, got %+v", details2)
	}
}

func TestCleanupOldEpochDataRemovesOnlyExpiredEpoch(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	store := ctx.KVStore(k.storeKey)

	expiredEpoch := uint64(3)
	recentEpoch := uint64(4)

	store.Set(types.KeyChallenge(expiredEpoch), []byte("expired-challenge"))
	store.Set(types.KeyChallenge(recentEpoch), []byte("recent-challenge"))
	store.Set(types.KeyAIResponse(expiredEpoch, "agent1"), []byte("expired-response"))
	store.Set(types.KeyAIResponse(recentEpoch, "agent1"), []byte("recent-response"))
	store.Set(types.KeyEpochActivity(expiredEpoch, "agent1"), types.Uint64ToBytes(1))
	store.Set(types.KeyEpochActivity(recentEpoch, "agent1"), types.Uint64ToBytes(1))
	store.Set(types.KeyDeployCount(expiredEpoch, "agent1"), types.Uint64ToBytes(1))
	store.Set(types.KeyDeployCount(recentEpoch, "agent1"), types.Uint64ToBytes(1))
	store.Set(types.KeyContractCall(expiredEpoch, "agent1"), types.Uint64ToBytes(1))
	store.Set(types.KeyContractCall(recentEpoch, "agent1"), types.Uint64ToBytes(1))

	k.cleanupOldEpochData(ctx, expiredEpoch)

	if store.Has(types.KeyChallenge(expiredEpoch)) || store.Has(types.KeyAIResponse(expiredEpoch, "agent1")) {
		t.Fatal("expired epoch keys should be deleted")
	}
	if !store.Has(types.KeyChallenge(recentEpoch)) || !store.Has(types.KeyAIResponse(recentEpoch, "agent1")) {
		t.Fatal("recent epoch keys should remain")
	}
}

func TestCleanupOldDailyRegDataRemovesEntriesOlderThanOneDay(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	store := ctx.KVStore(k.storeKey)

	store.Set(dailyRegisterKey("agent1", 1), types.Uint64ToBytes(1))
	store.Set(dailyRegisterKey("agent1", 2), types.Uint64ToBytes(1))
	store.Set(dailyRegisterKey("agent1", 3), types.Uint64ToBytes(1))

	k.cleanupOldDailyRegData(ctx, 3)

	if store.Has(dailyRegisterKey("agent1", 1)) {
		t.Fatal("day 1 entry should be deleted")
	}
	if !store.Has(dailyRegisterKey("agent1", 2)) || !store.Has(dailyRegisterKey("agent1", 3)) {
		t.Fatal("current day and previous day entries should remain")
	}
}

func TestCleanupOldDailyRegDataHandlesSlashInBinaryDaySuffix(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	store := ctx.KVStore(k.storeKey)

	trickyDay := int64(47) // 0x2f, contains slash byte in the binary suffix
	store.Set(dailyRegisterKey("agent1", trickyDay), types.Uint64ToBytes(1))

	k.cleanupOldDailyRegData(ctx, trickyDay+2)

	if store.Has(dailyRegisterKey("agent1", trickyDay)) {
		t.Fatal("tricky binary-suffix day entry should be deleted")
	}
}

func TestCleanupOldDailyRegDataBatchesLargeBacklog(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	store := ctx.KVStore(k.storeKey)

	// Simulate 2600 agents × 10 old days = 26000 entries
	for i := 0; i < 2600; i++ {
		addr := "agent" + strings.Repeat("0", 4-len(fmt.Sprintf("%d", i))) + fmt.Sprintf("%d", i)
		for day := int64(1); day <= 10; day++ {
			store.Set(dailyRegisterKey(addr, day), types.Uint64ToBytes(1))
		}
	}

	// First call: should hit the cap and return false (incomplete)
	done := k.cleanupOldDailyRegData(ctx, 12) // cutoff = 11, deletes days 1-10
	if done {
		t.Fatal("expected batched cleanup to be incomplete on first call (26000 > 5000 cap)")
	}

	// Keep calling until complete
	rounds := 1
	for !done {
		done = k.cleanupOldDailyRegData(ctx, 12)
		rounds++
		if rounds > 20 {
			t.Fatal("cleanup did not complete within 20 rounds")
		}
	}
	t.Logf("cleanup completed in %d rounds for 26000 entries", rounds)

	// Verify all old entries are gone
	for day := int64(1); day <= 10; day++ {
		if store.Has(dailyRegisterKey("agent0000", day)) {
			t.Fatalf("day %d entry should be deleted", day)
		}
	}
}

func TestCalculateContributionScoreCapsLargeCounters(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	agent := types.Agent{
		Address:    "agent1",
		Reputation: 80,
		Status:     types.AgentStatus_AGENT_STATUS_ONLINE,
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyDeployCount(1, agent.Address), types.Uint64ToBytes(^uint64(0)))
	store.Set(types.KeyContractCall(1, agent.Address), types.Uint64ToBytes(^uint64(0)))
	store.Set(types.KeyEpochActivity(1, agent.Address), types.Uint64ToBytes(250))

	score := k.calculateContributionScore(ctx, 1, agent)
	if score != 801105 {
		t.Fatalf("expected capped contribution score 801105, got %d", score)
	}
}

func TestV110SenderCompatFlagRoundTrip(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	if !k.IsV110UpgradeActivated(ctx) {
		t.Fatal("non-mainnet contexts should activate v1.1 behavior by default")
	}

	ctx = ctx.WithChainID(mainnetChainID).WithBlockHeight(V110UpgradeHeight - 1)
	if k.IsV110UpgradeActivated(ctx) {
		t.Fatal("mainnet should stay on legacy behavior before upgrade height")
	}

	ctx = ctx.WithBlockHeight(V110UpgradeHeight)
	if !k.IsV110UpgradeActivated(ctx) {
		t.Fatal("mainnet should activate v1.1 behavior at upgrade height")
	}
}

func TestLastDailyRegCleanupDayRoundTrip(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	if got := k.GetLastDailyRegCleanupDay(ctx); got != -1 {
		t.Fatalf("expected unset cleanup day -1, got %d", got)
	}
	k.SetLastDailyRegCleanupDay(ctx, 7)
	if got := k.GetLastDailyRegCleanupDay(ctx); got != 7 {
		t.Fatalf("expected cleanup day 7, got %d", got)
	}
}

func TestRecordEvidenceTxHashStoresNormalizedHex(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	hash := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	k.RecordEvidenceTxHash(ctx, hash)
	if !k.HasEvidenceTxHash(ctx, hash.Hex()) {
		t.Fatal("stored tx hash should be queryable")
	}
	if !k.HasEvidenceTxHash(ctx, strings.ToUpper(hash.Hex())) {
		t.Fatal("lookup should normalize evidence hash case")
	}
}

// ---------------------------------------------------------------------------
// C1 Fix: Verify pre-upgrade scoring is NOT broken by cheater-detection change.
// This is the critical regression test: before the fix, expectedHash was
// cleared for both detectCheaters AND scoreResponseByHash, making all
// pre-upgrade AI scores 0 — which would break historical block replay.
// ---------------------------------------------------------------------------

func TestEvaluateEpochChallengesScoresCorrectlyBeforeUpgrade(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	// Simulate mainnet before upgrade height
	ctx = ctx.WithChainID(mainnetChainID).WithBlockHeight(V110UpgradeHeight - 100)

	if k.IsV110UpgradeActivated(ctx) {
		t.Fatal("precondition: should be pre-upgrade")
	}

	// Pick a known challenge from the pool (index 0)
	epoch := uint64(1)
	params := k.GetParams(ctx)
	params.EpochLength = 720
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	// Generate challenge for epoch 1
	challenge := k.GenerateChallenge(ctx, epoch)
	expectedHash := getChallengeAnswerHash(challenge)
	if expectedHash == "" {
		t.Fatal("precondition: expectedHash must be non-empty for scoring")
	}

	// Store a response with arbitrary wrong reveal data — scoring should produce
	// score=10 (wrong answer), NOT score=0 (broken empty-hash path).
	wrongResp := types.AIResponse{
		ValidatorAddress: "axon1validator1",
		RevealData:       "some_wrong_answer",
		CommitHash:       "commit1",
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&wrongResp)
	store.Set(types.KeyAIResponse(epoch, wrongResp.ValidatorAddress), bz)

	// Register the validator as agent
	setTestAgent(k, ctx, wrongResp.ValidatorAddress, 1)

	// Evaluate
	k.EvaluateEpochChallenges(ctx, epoch)

	// Read back the evaluated response
	var evaluated types.AIResponse
	evalBz := store.Get(types.KeyAIResponse(epoch, wrongResp.ValidatorAddress))
	if evalBz == nil {
		t.Fatal("evaluated response should exist")
	}
	k.cdc.MustUnmarshal(evalBz, &evaluated)

	// CRITICAL: score must be 10 (wrong answer scored against real hash),
	// NOT 0 (which would mean expectedHash was incorrectly cleared).
	if evaluated.Score == 0 {
		t.Fatal("REGRESSION: pre-upgrade scoring returned 0 — expectedHash was " +
			"incorrectly cleared for scoreResponseByHash (C1 bug)")
	}
	if evaluated.Score != 10 {
		t.Fatalf("expected score 10 for wrong answer, got %d", evaluated.Score)
	}

	// Verify bonus was set (score=10 → bonus=0, but SetAIBonus was called)
	bonus := k.GetAIBonus(ctx, wrongResp.ValidatorAddress)
	if bonus != 0 {
		t.Fatalf("expected bonus 0 for score=10, got %d", bonus)
	}
}

func TestEvaluateEpochChallengesScoresCorrectlyAfterUpgrade(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)

	// Simulate mainnet at upgrade height
	ctx = ctx.WithChainID(mainnetChainID).WithBlockHeight(V110UpgradeHeight)

	if !k.IsV110UpgradeActivated(ctx) {
		t.Fatal("precondition: should be post-upgrade")
	}

	epoch := uint64(1)
	params := k.GetParams(ctx)
	params.EpochLength = 720
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	challenge := k.GenerateChallenge(ctx, epoch)
	expectedHash := getChallengeAnswerHash(challenge)
	if expectedHash == "" {
		t.Fatal("precondition: expectedHash must be non-empty")
	}

	wrongResp := types.AIResponse{
		ValidatorAddress: "axon1validator1",
		RevealData:       "wrong_answer",
		CommitHash:       "commit1",
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&wrongResp)
	store.Set(types.KeyAIResponse(epoch, wrongResp.ValidatorAddress), bz)
	setTestAgent(k, ctx, wrongResp.ValidatorAddress, 1)

	k.EvaluateEpochChallenges(ctx, epoch)

	var evaluated types.AIResponse
	k.cdc.MustUnmarshal(store.Get(types.KeyAIResponse(epoch, wrongResp.ValidatorAddress)), &evaluated)

	if evaluated.Score != 10 {
		t.Fatalf("expected score 10 for wrong answer post-upgrade, got %d", evaluated.Score)
	}
}

// Test that correct-answer groups are NOT penalized after upgrade (F1 fix)
// but ARE penalized before upgrade (preserving old behavior for replay).
func TestDetectCheatersBehaviorAcrossUpgradeBoundary(t *testing.T) {
	k, _ := newL2ReputationTestKeeper(t)

	// Construct responses: 3 validators give the same correct answer
	normalized := normalizeAnswer("pbft")
	hash := sha256.Sum256([]byte(normalized))
	correctHash := hex.EncodeToString(hash[:])

	responses := []types.AIResponse{
		{ValidatorAddress: "v1", RevealData: "PBFT"},
		{ValidatorAddress: "v2", RevealData: "pbft"},
		{ValidatorAddress: "v3", RevealData: " PBFT "},
	}

	// Pre-upgrade: expectedHash="" → correct answer group IS flagged
	preUpgradeCheaters := k.detectCheaters(responses, "")
	if len(preUpgradeCheaters) != 3 {
		t.Fatalf("pre-upgrade: expected 3 cheaters (old behavior), got %d", len(preUpgradeCheaters))
	}

	// Post-upgrade: expectedHash=correctHash → correct answer group is SKIPPED
	postUpgradeCheaters := k.detectCheaters(responses, correctHash)
	if len(postUpgradeCheaters) != 0 {
		t.Fatalf("post-upgrade: expected 0 cheaters (correct answers excluded), got %d", len(postUpgradeCheaters))
	}
}

// Simulate a full historical replay scenario: same binary processes blocks
// before AND after the upgrade height, producing consistent state.
func TestHistoricalReplayConsistency(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	ctx = ctx.WithChainID(mainnetChainID)

	params := k.GetParams(ctx)
	params.EpochLength = 720
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("set params: %v", err)
	}

	// --- Phase 1: Pre-upgrade epoch ---
	preUpgradeHeight := V110UpgradeHeight - 500
	ctx = ctx.WithBlockHeight(preUpgradeHeight)
	epoch1 := uint64(1)

	challenge1 := k.GenerateChallenge(ctx, epoch1)
	expectedHash1 := getChallengeAnswerHash(challenge1)

	resp1 := types.AIResponse{
		ValidatorAddress: "axon1replay1",
		RevealData:       "some_answer",
		CommitHash:       "c1",
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyAIResponse(epoch1, resp1.ValidatorAddress), k.cdc.MustMarshal(&resp1))
	setTestAgent(k, ctx, resp1.ValidatorAddress, 1)

	k.EvaluateEpochChallenges(ctx, epoch1)

	var eval1 types.AIResponse
	k.cdc.MustUnmarshal(store.Get(types.KeyAIResponse(epoch1, resp1.ValidatorAddress)), &eval1)

	// Pre-upgrade scoring must produce non-zero score (wrong answer = 10)
	if eval1.Score == 0 && expectedHash1 != "" {
		t.Fatal("REPLAY FAILURE: pre-upgrade epoch scored 0, would diverge from v1.0.0 state")
	}

	// --- Phase 2: Post-upgrade epoch ---
	postUpgradeHeight := V110UpgradeHeight + 100
	ctx = ctx.WithBlockHeight(postUpgradeHeight)
	epoch2 := uint64(2)

	challenge2 := k.GenerateChallenge(ctx, epoch2)
	expectedHash2 := getChallengeAnswerHash(challenge2)
	_ = expectedHash2

	resp2 := types.AIResponse{
		ValidatorAddress: "axon1replay2",
		RevealData:       "another_answer",
		CommitHash:       "c2",
	}
	store.Set(types.KeyAIResponse(epoch2, resp2.ValidatorAddress), k.cdc.MustMarshal(&resp2))
	setTestAgent(k, ctx, resp2.ValidatorAddress, 1)

	k.EvaluateEpochChallenges(ctx, epoch2)

	var eval2 types.AIResponse
	k.cdc.MustUnmarshal(store.Get(types.KeyAIResponse(epoch2, resp2.ValidatorAddress)), &eval2)

	// Post-upgrade scoring also works
	if eval2.Score == 0 {
		t.Fatal("post-upgrade epoch scored 0, scoring broken")
	}

	// Both phases produced consistent non-zero results
	t.Logf("replay consistency: pre-upgrade score=%d, post-upgrade score=%d", eval1.Score, eval2.Score)
}

func TestCleanupEvidenceTxHashesDeletesExpiredPrimaryAndHeightIndex(t *testing.T) {
	k, ctx := newL2ReputationTestKeeper(t)
	ctx = ctx.WithBlockHeight(100)

	expiredHash := common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	recentHash := common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	k.RecordEvidenceTxHash(ctx, expiredHash)
	ctx = ctx.WithBlockHeight(101)
	k.RecordEvidenceTxHash(ctx, recentHash)

	k.cleanupEvidenceTxHashes(ctx, 100)

	if k.HasEvidenceTxHash(ctx, expiredHash.Hex()) {
		t.Fatal("expired evidence hash should be deleted")
	}
	if !k.HasEvidenceTxHash(ctx, recentHash.Hex()) {
		t.Fatal("recent evidence hash should remain")
	}
}
