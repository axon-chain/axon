package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/axon-chain/axon/x/agent/keeper"
	"github.com/axon-chain/axon/x/agent/types"
)

// ---------------------------------------------------------------------------
// Challenge Pool Size
// ---------------------------------------------------------------------------

func TestChallengePoolMinimumSize(t *testing.T) {
	pool := keeper.GetChallengePoolSize()
	if len(pool) < 100 {
		t.Errorf("challenge pool has %d entries, want >= 100", len(pool))
	}
}

// ---------------------------------------------------------------------------
// scoreResponse — edge cases beyond keeper_test.go
// ---------------------------------------------------------------------------

func TestScoreResponseExactMatchCaseInsensitive(t *testing.T) {
	tests := []struct {
		reveal string
		answer string
		want   int
	}{
		{"o(log n)", "O(log n)", 100},
		{"O(LOG N)", "O(log n)", 100},
		{"  O(log n)  ", "O(log n)", 100},
		{"PBFT", "PBFT", 100},
		{"pbft", "PBFT", 100},
	}
	for _, tt := range tests {
		score := keeper.ScoreResponseForTest(tt.reveal, tt.answer)
		if score != tt.want {
			t.Errorf("scoreResponse(%q, %q) = %d, want %d", tt.reveal, tt.answer, score, tt.want)
		}
	}
}

func TestScoreResponsePartialMatch(t *testing.T) {
	tests := []struct {
		reveal string
		answer string
		want   int
	}{
		{"the answer is CALL obviously", "CALL", 50},
		{"I think it uses PBFT consensus", "PBFT", 50},
		{"stack is the data structure", "stack", 50},
	}
	for _, tt := range tests {
		score := keeper.ScoreResponseForTest(tt.reveal, tt.answer)
		if score != tt.want {
			t.Errorf("scoreResponse(%q, %q) = %d, want %d", tt.reveal, tt.answer, score, tt.want)
		}
	}
}

func TestScoreResponseNoMatch(t *testing.T) {
	score := keeper.ScoreResponseForTest("completely wrong answer", "CALL")
	if score != 10 {
		t.Errorf("no-match score = %d, want 10", score)
	}
}

func TestScoreResponseBothEmpty(t *testing.T) {
	score := keeper.ScoreResponseForTest("", "")
	if score != 0 {
		t.Errorf("both empty score = %d, want 0 (empty reveal always returns 0)", score)
	}
}

func TestScoreResponseEmptyReveal(t *testing.T) {
	score := keeper.ScoreResponseForTest("", "O(log n)")
	if score != 0 {
		t.Errorf("empty reveal should return 0, got %d", score)
	}
}

func TestScoreResponseEmptyAnswer(t *testing.T) {
	score := keeper.ScoreResponseForTest("some answer", "")
	if score < 10 {
		t.Errorf("non-empty reveal with empty answer: score = %d, expected >= 10", score)
	}
}

func TestScoreResponseAnswerContainedInReveal(t *testing.T) {
	score := keeper.ScoreResponseForTest("dijkstra", "dijkstra")
	if score != 100 {
		t.Errorf("exact match 'dijkstra' = %d, want 100", score)
	}

	score = keeper.ScoreResponseForTest("i think dijkstra is the answer", "dijkstra")
	if score != 50 {
		t.Errorf("partial match containing 'dijkstra' = %d, want 50", score)
	}
}

func TestScoreResponseRevealContainedInAnswer(t *testing.T) {
	// When the answer is longer and contains the reveal
	score := keeper.ScoreResponseForTest("DNS", "DNS")
	if score != 100 {
		t.Errorf("exact match DNS = %d, want 100", score)
	}
}

// ---------------------------------------------------------------------------
// normalizeAnswer — edge cases beyond keeper_test.go
// ---------------------------------------------------------------------------

func TestNormalizeAnswerWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"\tHello\tWorld\t", "helloworld"},
		{"\n\nTest\n\n", "test"},
		{"\r\nWindows\r\n", "windows"},
		{"  multiple   spaces  ", "multiplespaces"},
		{"", ""},
		{"   ", ""},
	}
	for _, tt := range tests {
		got := keeper.NormalizeAnswerForTest(tt.input)
		if got != tt.want {
			t.Errorf("normalizeAnswer(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeAnswerPreservesNonAlpha(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"O(n^2)", "o(n^2)"},
		{"ERC-721", "erc-721"},
		{"3x^2", "3x^2"},
		{"123/456", "123/456"},
		{"a+b=c", "a+b=c"},
	}
	for _, tt := range tests {
		got := keeper.NormalizeAnswerForTest(tt.input)
		if got != tt.want {
			t.Errorf("normalizeAnswer(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeAnswerSingleChar(t *testing.T) {
	if got := keeper.NormalizeAnswerForTest("A"); got != "a" {
		t.Errorf("normalizeAnswer(\"A\") = %q, want \"a\"", got)
	}
	if got := keeper.NormalizeAnswerForTest("z"); got != "z" {
		t.Errorf("normalizeAnswer(\"z\") = %q, want \"z\"", got)
	}
	if got := keeper.NormalizeAnswerForTest("5"); got != "5" {
		t.Errorf("normalizeAnswer(\"5\") = %q, want \"5\"", got)
	}
}

// ---------------------------------------------------------------------------
// calculateAIBonus — exact boundary values
// ---------------------------------------------------------------------------

func TestCalculateAIBonusExactBoundaries(t *testing.T) {
	tests := []struct {
		score int
		want  int64
	}{
		{0, 0},
		{19, 0},
		{20, 5}, // boundary: >= 20
		{49, 5},
		{50, 10}, // boundary: >= 50
		{69, 10},
		{70, 20}, // boundary: >= 70
		{89, 20},
		{90, 30}, // boundary: >= 90
		{99, 30},
		{100, 30},
	}
	for _, tt := range tests {
		got := keeper.CalculateAIBonusForTest(tt.score)
		if got != tt.want {
			t.Errorf("calculateAIBonus(%d) = %d, want %d", tt.score, got, tt.want)
		}
	}
}

func TestCalculateAIBonusNonNegative(t *testing.T) {
	for score := 0; score <= 100; score++ {
		bonus := keeper.CalculateAIBonusForTest(score)
		if bonus < 0 {
			t.Errorf("calculateAIBonus(%d) = %d, should be non-negative", score, bonus)
		}
	}
}

func TestCalculateAIBonusMonotonic(t *testing.T) {
	prev := keeper.CalculateAIBonusForTest(0)
	for score := 1; score <= 100; score++ {
		cur := keeper.CalculateAIBonusForTest(score)
		if cur < prev {
			t.Errorf("AI bonus decreased from score %d (%d) to score %d (%d)",
				score-1, prev, score, cur)
		}
		prev = cur
	}
}

// ---------------------------------------------------------------------------
// detectCheaters — scenarios beyond deflation_test.go
// ---------------------------------------------------------------------------

func TestDetectCheaters3PlusDuplicates(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", CommitHash: "hash1", RevealData: "PBFT"},
		{ValidatorAddress: "axon1bbb", CommitHash: "hash2", RevealData: "pbft"},
		{ValidatorAddress: "axon1ccc", CommitHash: "hash3", RevealData: " PBFT "},
		{ValidatorAddress: "axon1ddd", CommitHash: "hash4", RevealData: "RAFT"},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, "")

	if len(cheaters) != 3 {
		t.Errorf("expected 3 cheaters, got %d", len(cheaters))
	}
	for _, addr := range []string{"axon1aaa", "axon1bbb", "axon1ccc"} {
		if !cheaters[addr] {
			t.Errorf("%s should be flagged as cheater", addr)
		}
	}
	if cheaters["axon1ddd"] {
		t.Error("axon1ddd should NOT be flagged (unique answer)")
	}
}

func TestDetectCheatersAllSameAnswer(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", CommitHash: "hash1", RevealData: "Dijkstra"},
		{ValidatorAddress: "axon1bbb", CommitHash: "hash2", RevealData: "dijkstra"},
		{ValidatorAddress: "axon1ccc", CommitHash: "hash3", RevealData: "  DIJKSTRA"},
		{ValidatorAddress: "axon1ddd", CommitHash: "hash4", RevealData: "Dijkstra  "},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, "")
	if len(cheaters) != 4 {
		t.Errorf("expected all 4 flagged as cheaters, got %d", len(cheaters))
	}
}

func TestDetectCheatersSingleResponse(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", CommitHash: "only_one", RevealData: "a"},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, "")
	if len(cheaters) != 0 {
		t.Errorf("single response should have 0 cheaters, got %d", len(cheaters))
	}
}

func TestDetectCheatersEmptyList(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	cheaters := keeper.DetectCheatersForTest(k, []types.AIResponse{}, "")
	if len(cheaters) != 0 {
		t.Errorf("empty responses should have 0 cheaters, got %d", len(cheaters))
	}
}

func TestDetectCheatersMultipleGroups(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", CommitHash: "hash1", RevealData: "alpha"},
		{ValidatorAddress: "axon1bbb", CommitHash: "hash2", RevealData: "ALPHA"},
		{ValidatorAddress: "axon1ccc", CommitHash: "hash3", RevealData: " alpha "},
		{ValidatorAddress: "axon1ddd", CommitHash: "hash4", RevealData: "beta"},
		{ValidatorAddress: "axon1eee", CommitHash: "hash5", RevealData: "BETA"},
		{ValidatorAddress: "axon1fff", CommitHash: "hash6", RevealData: " beta "},
		{ValidatorAddress: "axon1ggg", CommitHash: "hash7", RevealData: "gamma"},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, "")
	if len(cheaters) != 6 {
		t.Errorf("expected 6 cheaters (2 groups of 3), got %d", len(cheaters))
	}
	if cheaters["axon1ggg"] {
		t.Error("axon1ggg should NOT be flagged (unique answer)")
	}
}

func TestDetectCheatersMixedEmptyAndDuplicate(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", CommitHash: "hash1", RevealData: ""},
		{ValidatorAddress: "axon1bbb", CommitHash: "hash2", RevealData: "merkle"},
		{ValidatorAddress: "axon1ccc", CommitHash: "hash3", RevealData: "MERKLE"},
		{ValidatorAddress: "axon1ddd", CommitHash: "hash4", RevealData: " merkle "},
		{ValidatorAddress: "axon1eee", CommitHash: "hash5", RevealData: ""},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, "")
	if len(cheaters) != 3 {
		t.Errorf("expected 3 cheaters (dup triplet, empty ignored), got %d", len(cheaters))
	}
	if !cheaters["axon1bbb"] || !cheaters["axon1ccc"] || !cheaters["axon1ddd"] {
		t.Error("the duplicate-answer triplet should be flagged")
	}
	if cheaters["axon1aaa"] || cheaters["axon1eee"] {
		t.Error("empty-reveal addresses should not be flagged")
	}
}

// ---------------------------------------------------------------------------
// Cheat Penalty Constants — verify relationships
// ---------------------------------------------------------------------------

func TestCheatPenaltyReputationIsNegative(t *testing.T) {
	if keeper.CheatPenaltyReputation >= 0 {
		t.Errorf("CheatPenaltyReputation = %d, should be negative", keeper.CheatPenaltyReputation)
	}
}

func TestCheatPenaltyStakePercentRange(t *testing.T) {
	if keeper.CheatPenaltyStakePercent <= 0 || keeper.CheatPenaltyStakePercent > 100 {
		t.Errorf("CheatPenaltyStakePercent = %d, should be in (0, 100]", keeper.CheatPenaltyStakePercent)
	}
}

func TestDetectCheatersSkipsCorrectAnswerGroup(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	normalized := keeper.NormalizeAnswerForTest("PBFT")
	hash := sha256.Sum256([]byte(normalized))
	expectedHash := hex.EncodeToString(hash[:])

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", RevealData: "PBFT"},
		{ValidatorAddress: "axon1bbb", RevealData: "pbft"},
		{ValidatorAddress: "axon1ccc", RevealData: " PBFT "},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, expectedHash)
	if len(cheaters) != 0 {
		t.Fatalf("expected no cheaters, got %d", len(cheaters))
	}
}

func TestDetectCheatersMixedCorrectAndWrongGroups(t *testing.T) {
	k, _, _ := setupTestKeeper(t)

	normalized := keeper.NormalizeAnswerForTest("PBFT")
	hash := sha256.Sum256([]byte(normalized))
	expectedHash := hex.EncodeToString(hash[:])

	responses := []types.AIResponse{
		{ValidatorAddress: "axon1aaa", RevealData: "PBFT"},
		{ValidatorAddress: "axon1bbb", RevealData: "pbft"},
		{ValidatorAddress: "axon1ccc", RevealData: " PBFT "},
		{ValidatorAddress: "axon1ddd", RevealData: "RAFT"},
		{ValidatorAddress: "axon1eee", RevealData: "raft"},
		{ValidatorAddress: "axon1fff", RevealData: " raft "},
	}

	cheaters := keeper.DetectCheatersForTest(k, responses, expectedHash)
	if len(cheaters) != 3 {
		t.Fatalf("expected 3 wrong-answer cheaters, got %d", len(cheaters))
	}
	for _, addr := range []string{"axon1ddd", "axon1eee", "axon1fff"} {
		if !cheaters[addr] {
			t.Fatalf("%s should be flagged", addr)
		}
	}
	for _, addr := range []string{"axon1aaa", "axon1bbb", "axon1ccc"} {
		if cheaters[addr] {
			t.Fatalf("%s should not be flagged", addr)
		}
	}
}
