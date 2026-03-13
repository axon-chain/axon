package keeper_test

import (
	"testing"

	"github.com/axon-chain/axon/x/agent/keeper"
	"github.com/axon-chain/axon/x/agent/types"
)

// ---------------------------------------------------------------------------
// Reputation Constants (whitepaper §7.1)
// ---------------------------------------------------------------------------

func TestReputationGainConstants(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  int64
	}{
		{"ReputationGainEpochOnline", types.ReputationGainEpochOnline, 1},
		{"ReputationGainHeartbeatStreak", types.ReputationGainHeartbeatStreak, 1},
		{"ReputationGainActiveEpoch", types.ReputationGainActiveEpoch, 1},
	}
	for _, tt := range tests {
		if tt.value != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.want)
		}
	}
}

func TestReputationLossConstants(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  int64
	}{
		{"ReputationLossOffline", types.ReputationLossOffline, -5},
		{"ReputationLossSlashing", types.ReputationLossSlashing, -50},
		{"ReputationLossNoHeartbeatEpoch", types.ReputationLossNoHeartbeatEpoch, -1},
	}
	for _, tt := range tests {
		if tt.value != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.want)
		}
	}
}

func TestReputationLossOfflineIsNegative(t *testing.T) {
	if types.ReputationLossOffline >= 0 {
		t.Error("ReputationLossOffline must be negative")
	}
	if types.ReputationLossSlashing >= 0 {
		t.Error("ReputationLossSlashing must be negative")
	}
	if types.ReputationLossNoHeartbeatEpoch >= 0 {
		t.Error("ReputationLossNoHeartbeatEpoch must be negative")
	}
}

func TestReputationSlashingIsHarshest(t *testing.T) {
	if types.ReputationLossSlashing > types.ReputationLossOffline {
		t.Errorf("slashing (%d) should be harsher than offline (%d)",
			types.ReputationLossSlashing, types.ReputationLossOffline)
	}
}

// ---------------------------------------------------------------------------
// Reputation Bonus Edge Cases — beyond what deflation_test.go covers
// ---------------------------------------------------------------------------

func TestReputationBonusExactBoundaries(t *testing.T) {
	tests := []struct {
		rep  uint64
		want int64
	}{
		{0, 0},
		{1, 0},
		{29, 0},
		{30, 5},  // inclusive lower bound
		{50, 10}, // >=50 enters 10% tier
		{51, 10},
		{70, 15}, // >=70 enters 15% tier
		{71, 15},
		{90, 20}, // >=90 enters 20% tier
		{91, 20},
		{100, 20},
	}
	for _, tt := range tests {
		got := keeper.ReputationBonusPercentForTest(tt.rep)
		if got != tt.want {
			t.Errorf("reputationBonusPercent(%d) = %d, want %d", tt.rep, got, tt.want)
		}
	}
}

func TestReputationBonusBeyondMax(t *testing.T) {
	bonus150 := keeper.ReputationBonusPercentForTest(150)
	bonus100 := keeper.ReputationBonusPercentForTest(100)
	if bonus150 != bonus100 {
		t.Errorf("reputation 150 bonus (%d) should equal reputation 100 bonus (%d) — values above max should stay in top tier",
			bonus150, bonus100)
	}
}

func TestReputationBonusMonotonic(t *testing.T) {
	prev := keeper.ReputationBonusPercentForTest(0)
	for rep := uint64(1); rep <= 100; rep++ {
		cur := keeper.ReputationBonusPercentForTest(rep)
		if cur < prev {
			t.Errorf("bonus decreased from rep %d (%d%%) to rep %d (%d%%)",
				rep-1, prev, rep, cur)
		}
		prev = cur
	}
}

func TestReputationBonusNonNegative(t *testing.T) {
	for rep := uint64(0); rep <= 100; rep++ {
		bonus := keeper.ReputationBonusPercentForTest(rep)
		if bonus < 0 {
			t.Errorf("reputationBonusPercent(%d) = %d, should be non-negative", rep, bonus)
		}
	}
}

func TestReputationBonusMaxCap(t *testing.T) {
	maxBonus := keeper.ReputationBonusPercentForTest(100)
	if maxBonus != 20 {
		t.Errorf("max reputation bonus = %d%%, want 20%%", maxBonus)
	}
}

// ---------------------------------------------------------------------------
// Params: Reputation bounds
// ---------------------------------------------------------------------------

func TestDefaultParamsReputationBounds(t *testing.T) {
	p := types.DefaultParams()
	if p.MaxReputation != 100 {
		t.Errorf("MaxReputation = %d, want 100", p.MaxReputation)
	}
	if p.InitialReputation != 10 {
		t.Errorf("InitialReputation = %d, want 10", p.InitialReputation)
	}
	if p.InitialReputation > p.MaxReputation {
		t.Error("InitialReputation should not exceed MaxReputation")
	}
}

func TestDeregisterCooldownConstant(t *testing.T) {
	if types.DeregisterCooldownBlocks != 120960 {
		t.Errorf("DeregisterCooldownBlocks = %d, want 120960", types.DeregisterCooldownBlocks)
	}
}
