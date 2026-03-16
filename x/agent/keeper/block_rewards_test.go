package keeper_test

import (
	"math/big"
	"testing"

	"github.com/axon-chain/axon/x/agent/keeper"
)

const (
	blocksPerYear   int64 = 6_307_200
	halvingInterval int64 = blocksPerYear * 4
)

// ---------------------------------------------------------------------------
// calculateBlockReward — additional scenarios
// ---------------------------------------------------------------------------

func TestBlockRewardAtBlockZero(t *testing.T) {
	reward := keeper.ExportCalculateBlockReward(0)
	if reward.IsZero() {
		t.Error("calculateBlockReward(0) should return non-zero (halving 0)")
	}
}

func TestBlockRewardAtBlockOne(t *testing.T) {
	reward := keeper.ExportCalculateBlockReward(1)
	if reward.IsZero() {
		t.Error("calculateBlockReward(1) should return non-zero")
	}
}

func TestBlockRewardConsistentWithinEra(t *testing.T) {
	r1 := keeper.ExportCalculateBlockReward(100)
	r2 := keeper.ExportCalculateBlockReward(1_000_000)
	r3 := keeper.ExportCalculateBlockReward(halvingInterval - 1)

	if !r1.Equal(r2) || !r2.Equal(r3) {
		t.Errorf("reward should be constant within era 0: %s, %s, %s", r1, r2, r3)
	}
}

func TestBlockRewardHalvingRatioExact(t *testing.T) {
	era0 := keeper.ExportCalculateBlockReward(halvingInterval - 1)
	era1 := keeper.ExportCalculateBlockReward(halvingInterval)

	if era0.IsZero() || era1.IsZero() {
		t.Fatal("rewards for era 0 and era 1 should be non-zero")
	}

	// era1 should be exactly era0 / 2 (integer division via right-shift)
	expected := era0.QuoRaw(2)
	if !era1.Equal(expected) {
		t.Errorf("era1 reward = %s, want %s (half of era0 %s)", era1, expected, era0)
	}
}

func TestBlockRewardMultipleHalvings(t *testing.T) {
	tests := []struct {
		name     string
		height   int64
		halvings int64
	}{
		{"Year 1-4 (era 0)", 100, 0},
		{"Year 5-8 (era 1)", halvingInterval, 1},
		{"Year 9-12 (era 2)", halvingInterval * 2, 2},
		{"Year 13-16 (era 3)", halvingInterval * 3, 3},
		{"Year 17-20 (era 4)", halvingInterval * 4, 4},
	}

	baseReward := keeper.ExportCalculateBlockReward(1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := keeper.ExportCalculateBlockReward(tt.height)
			if reward.IsZero() {
				t.Fatalf("reward at height %d should be non-zero", tt.height)
			}

			// Expected = base >> halvings
			expectedBig := new(big.Int).Rsh(baseReward.BigInt(), uint(tt.halvings))
			if reward.BigInt().Cmp(expectedBig) != 0 {
				t.Errorf("height %d: got %s, want %s", tt.height, reward, expectedBig)
			}
		})
	}
}

func TestBlockRewardZeroAfterManyHalvings(t *testing.T) {
	tests := []int64{64, 100, 200}
	for _, halvings := range tests {
		reward := keeper.ExportCalculateBlockReward(halvingInterval * halvings)
		if !reward.IsZero() {
			t.Errorf("expected zero reward after %d halvings, got %s", halvings, reward)
		}
	}
}

func TestBlockRewardPositiveForFirst63Eras(t *testing.T) {
	for era := int64(0); era < 63; era++ {
		height := halvingInterval * era
		reward := keeper.ExportCalculateBlockReward(height)
		if reward.IsZero() {
			t.Errorf("reward at era %d (height %d) should be positive", era, height)
			break
		}
	}
}

func TestBlockRewardDecreasing(t *testing.T) {
	prev := keeper.ExportCalculateBlockReward(1)
	for era := int64(1); era < 10; era++ {
		cur := keeper.ExportCalculateBlockReward(halvingInterval * era)
		if !cur.LT(prev) {
			t.Errorf("era %d reward (%s) should be less than era %d reward (%s)",
				era, cur, era-1, prev)
		}
		prev = cur
	}
}

// ---------------------------------------------------------------------------
// calculateContributionPerBlock — additional scenarios
// ---------------------------------------------------------------------------

func TestContributionPerBlockAtBlockZero(t *testing.T) {
	reward := keeper.ExportCalculateContributionPerBlock(0)
	if reward.IsZero() {
		t.Error("contribution at block 0 should be non-zero")
	}
}

func TestContributionPerBlockConsistentWithinEra(t *testing.T) {
	r1 := keeper.ExportCalculateContributionPerBlock(100)
	r2 := keeper.ExportCalculateContributionPerBlock(1_000_000)
	r3 := keeper.ExportCalculateContributionPerBlock(halvingInterval - 1)

	if !r1.Equal(r2) || !r2.Equal(r3) {
		t.Errorf("contribution should be constant within era 0: %s, %s, %s", r1, r2, r3)
	}
}

func TestContributionPerBlockPhaseRatios(t *testing.T) {
	// Whitepaper §8.4 declining schedule: 35M → 25M → 15M → 5M per year
	phase0 := keeper.ExportCalculateContributionPerBlock(1)                  // Year 1-4: 35M/yr
	phase1 := keeper.ExportCalculateContributionPerBlock(blocksPerYear * 4)  // Year 5-8: 25M/yr
	phase2 := keeper.ExportCalculateContributionPerBlock(blocksPerYear * 8)  // Year 9-12: 15M/yr
	phase3 := keeper.ExportCalculateContributionPerBlock(blocksPerYear * 12) // Year 12+: 5M/yr

	if phase0.IsZero() || phase1.IsZero() || phase2.IsZero() || phase3.IsZero() {
		t.Fatal("all contribution phases should be non-zero")
	}

	// Each phase should be less than the previous
	if !phase1.LT(phase0) {
		t.Errorf("phase1 (%s) should be less than phase0 (%s)", phase1, phase0)
	}
	if !phase2.LT(phase1) {
		t.Errorf("phase2 (%s) should be less than phase1 (%s)", phase2, phase1)
	}
	if !phase3.LT(phase2) {
		t.Errorf("phase3 (%s) should be less than phase2 (%s)", phase3, phase2)
	}
}

func TestContributionPerBlockTailPhase(t *testing.T) {
	// Year 12+ all have same rate (5M/yr tail)
	r50 := keeper.ExportCalculateContributionPerBlock(blocksPerYear * 50)
	r100 := keeper.ExportCalculateContributionPerBlock(blocksPerYear * 100)
	if r50.IsZero() {
		t.Error("contribution at year 50 should be non-zero (tail phase)")
	}
	if !r50.Equal(r100) {
		t.Errorf("tail phase should be constant: year50=%s, year100=%s", r50, r100)
	}
}

func TestContributionPerBlockDecreasing(t *testing.T) {
	// Test that rewards decrease at each phase boundary
	boundaries := []int64{0, 4, 8, 12}
	prev := keeper.ExportCalculateContributionPerBlock(1)
	for i := 1; i < len(boundaries); i++ {
		cur := keeper.ExportCalculateContributionPerBlock(blocksPerYear * boundaries[i])
		if !cur.LT(prev) {
			t.Errorf("phase %d contribution (%s) should be less than phase %d (%s)",
				i, cur, i-1, prev)
		}
		prev = cur
	}
}

// ---------------------------------------------------------------------------
// reputationBonusPercent — boundary value tests for block reward weighting
// ---------------------------------------------------------------------------

func TestReputationBonusTierTransitions(t *testing.T) {
	transitions := []struct {
		below  uint64
		at     uint64
		expect int64
	}{
		{29, 30, 5},
		{49, 50, 10},
		{69, 70, 15},
		{89, 90, 20},
	}
	for _, tr := range transitions {
		belowBonus := keeper.ReputationBonusPercentForTest(tr.below)
		atBonus := keeper.ReputationBonusPercentForTest(tr.at)
		if atBonus != tr.expect {
			t.Errorf("rep %d bonus = %d, want %d", tr.at, atBonus, tr.expect)
		}
		if atBonus <= belowBonus {
			t.Errorf("rep %d bonus (%d) should be greater than rep %d bonus (%d)",
				tr.at, atBonus, tr.below, belowBonus)
		}
	}
}

// ---------------------------------------------------------------------------
// Block reward share constants
// ---------------------------------------------------------------------------

func TestBlockRewardSharesSumTo100(t *testing.T) {
	total := keeper.ProposerSharePercent + keeper.ValidatorPoolSharePercent + keeper.ReputationPoolSharePercent
	if total != 100 {
		t.Errorf("reward shares sum to %d%%, want 100%%", total)
	}
}

func TestBlockRewardShareValues(t *testing.T) {
	if keeper.ProposerSharePercent != 20 {
		t.Errorf("ProposerSharePercent = %d, want 20", keeper.ProposerSharePercent)
	}
	if keeper.ValidatorPoolSharePercent != 55 {
		t.Errorf("ValidatorPoolSharePercent = %d, want 55", keeper.ValidatorPoolSharePercent)
	}
	if keeper.ReputationPoolSharePercent != 25 {
		t.Errorf("ReputationPoolSharePercent = %d, want 25", keeper.ReputationPoolSharePercent)
	}
}

func TestContributionRewardCapScalesWithStakeShare(t *testing.T) {
	pool := big.NewInt(1_000_000)
	totalStake := big.NewInt(1_000)

	stakeA := big.NewInt(600)
	stakeB := big.NewInt(400)

	capA := keeper.ContributionRewardCapForTest(pool, stakeA, totalStake)
	capB := keeper.ContributionRewardCapForTest(pool, stakeB, totalStake)

	expectedA := big.NewInt(12_000) // 1,000,000 * 2% * 60%
	expectedB := big.NewInt(8_000)  // 1,000,000 * 2% * 40%

	if capA.Cmp(expectedA) != 0 {
		t.Fatalf("capA = %s, want %s", capA, expectedA)
	}
	if capB.Cmp(expectedB) != 0 {
		t.Fatalf("capB = %s, want %s", capB, expectedB)
	}
}

func TestContributionRewardCapIsSybilInvariant(t *testing.T) {
	pool := big.NewInt(1_000_000)
	totalStake := big.NewInt(1_000)

	combined := keeper.ContributionRewardCapForTest(pool, big.NewInt(600), totalStake)
	splitOne := keeper.ContributionRewardCapForTest(pool, big.NewInt(300), totalStake)
	splitTwo := keeper.ContributionRewardCapForTest(pool, big.NewInt(300), totalStake)

	splitTotal := new(big.Int).Add(splitOne, splitTwo)
	if splitTotal.Cmp(combined) != 0 {
		t.Fatalf("split cap total = %s, want %s", splitTotal, combined)
	}
}
