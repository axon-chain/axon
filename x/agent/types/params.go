package types

import (
	"fmt"

	"cosmossdk.io/math"
)

const (
	// DeregisterCooldownBlocks is 7 days at 5s/block = 120960 blocks.
	DeregisterCooldownBlocks int64 = 120960

	ReputationGainEpochOnline      int64 = 1
	ReputationGainHeartbeatStreak  int64 = 1
	ReputationGainActiveEpoch      int64 = 1
	ReputationLossOffline          int64 = -5
	ReputationLossSlashing         int64 = -50
	ReputationLossNoHeartbeatEpoch int64 = -1

	// MaxDailyRegistrations per address per ~24h window (whitepaper §10.5)
	MaxDailyRegistrations uint64 = 3

	// AIBonus bounds (whitepaper §7.3): -5% ~ +30%
	MinAIBonus int64 = -5
	MaxAIBonus int64 = 30
)

func DefaultParams() Params {
	return Params{
		MinRegisterStake:   100,
		RegisterBurnAmount: 20,
		ContractDeployBurn: 10,
		InitialReputation:  10,
		MaxReputation:      100,
		HeartbeatInterval:  100,
		HeartbeatTimeout:   720,
		EpochLength:        720,
		AiChallengeWindow:  50,
		ContributionCapBps: 200,

		// Mining formula
		Alpha: "0.5",
		Beta:  "1.5",
		RMax:  100,

		// Reputation tiers
		L1Cap:           40,
		L2Cap:           30,
		L1DecayPerEpoch: "0.1",
		L2DecayPerEpoch: "0.05",

		// L2 evaluation
		L2MinReporterRep:      30,
		L2MinAccountAge:       120960,
		L2BudgetPerAgent:      "0.1",
		L2BudgetCap:           100,
		L2AbuseThreshold:      50,
		L2MutualReportPenalty: "0.1",
		L2NoEvidenceWeight:    "0.3",

		// Reward distribution (basis points, sum = 10000)
		ProposerRewardBps: 2000,
		ValidatorPoolBps:  5500,
		ReputationPoolBps: 2500,

		// Anti-Sybil
		ContributionCapMultiplier: "3.0",
	}
}

func (p Params) Validate() error {
	if p.MinRegisterStake == 0 {
		return ErrInsufficientStake
	}
	if p.RegisterBurnAmount > p.MinRegisterStake {
		return fmt.Errorf("RegisterBurnAmount (%d) must not exceed MinRegisterStake (%d)", p.RegisterBurnAmount, p.MinRegisterStake)
	}
	if p.MaxReputation == 0 {
		return fmt.Errorf("MaxReputation must be > 0")
	}
	if p.EpochLength == 0 {
		return fmt.Errorf("EpochLength must be > 0")
	}
	if p.HeartbeatInterval <= 0 {
		return fmt.Errorf("HeartbeatInterval must be > 0, got %d", p.HeartbeatInterval)
	}
	if p.HeartbeatTimeout <= 0 {
		return fmt.Errorf("HeartbeatTimeout must be > 0, got %d", p.HeartbeatTimeout)
	}
	if p.HeartbeatTimeout <= p.HeartbeatInterval {
		return fmt.Errorf("HeartbeatTimeout (%d) must be > HeartbeatInterval (%d)", p.HeartbeatTimeout, p.HeartbeatInterval)
	}
	if p.AiChallengeWindow <= 0 {
		return fmt.Errorf("AiChallengeWindow must be > 0, got %d", p.AiChallengeWindow)
	}
	if p.InitialReputation > p.MaxReputation {
		return fmt.Errorf("InitialReputation (%d) must not exceed MaxReputation (%d)", p.InitialReputation, p.MaxReputation)
	}
	if p.ContributionCapBps > 10000 {
		return fmt.Errorf("ContributionCapBps (%d) must not exceed 10000 (100%%)", p.ContributionCapBps)
	}
	if p.RMax < 50 || p.RMax > 200 {
		return fmt.Errorf("RMax (%d) must be in [50, 200]", p.RMax)
	}
	if p.L1Cap < 20 || p.L1Cap > 60 {
		return fmt.Errorf("L1Cap (%d) must be in [20, 60]", p.L1Cap)
	}
	if p.L2Cap < 10 || p.L2Cap > 50 {
		return fmt.Errorf("L2Cap (%d) must be in [10, 50]", p.L2Cap)
	}
	rewardSum := p.ProposerRewardBps + p.ValidatorPoolBps + p.ReputationPoolBps
	if rewardSum != 10000 {
		return fmt.Errorf("reward BPS must sum to 10000, got %d", rewardSum)
	}
	if err := validateDecRange(p.Alpha, "Alpha", 0.3, 0.7); err != nil {
		return err
	}
	if err := validateDecRange(p.Beta, "Beta", 0.5, 3.0); err != nil {
		return err
	}
	if err := validateDecRange(p.L1DecayPerEpoch, "L1DecayPerEpoch", 0, 1.0); err != nil {
		return err
	}
	if err := validateDecRange(p.L2DecayPerEpoch, "L2DecayPerEpoch", 0, 0.5); err != nil {
		return err
	}
	if err := validateDecRange(p.L2BudgetPerAgent, "L2BudgetPerAgent", 0.01, 1.0); err != nil {
		return err
	}
	if err := validateDecRange(p.L2MutualReportPenalty, "L2MutualReportPenalty", 0, 0.5); err != nil {
		return err
	}
	if err := validateDecRange(p.L2NoEvidenceWeight, "L2NoEvidenceWeight", 0, 1.0); err != nil {
		return err
	}
	if err := validateDecRange(p.ContributionCapMultiplier, "ContributionCapMultiplier", 1.0, 10.0); err != nil {
		return err
	}
	if p.L2MinReporterRep < 10 || p.L2MinReporterRep > 80 {
		return fmt.Errorf("L2MinReporterRep (%d) must be in [10, 80]", p.L2MinReporterRep)
	}
	if p.L2BudgetCap < 10 || p.L2BudgetCap > 1000 {
		return fmt.Errorf("L2BudgetCap (%d) must be in [10, 1000]", p.L2BudgetCap)
	}
	if p.L2AbuseThreshold < 10 || p.L2AbuseThreshold > 200 {
		return fmt.Errorf("L2AbuseThreshold (%d) must be in [10, 200]", p.L2AbuseThreshold)
	}
	if p.L2MinAccountAge < 0 {
		return fmt.Errorf("L2MinAccountAge must be >= 0")
	}
	if err := p.validateAlphaBetaCross(); err != nil {
		return err
	}
	return nil
}

// Alpha × Beta <= 2.0 cross-constraint (deterministic via LegacyDec)
func (p Params) validateAlphaBetaCross() error {
	if p.Alpha == "" || p.Beta == "" {
		return nil
	}
	alpha, err := math.LegacyNewDecFromStr(p.Alpha)
	if err != nil {
		return nil
	}
	beta, err := math.LegacyNewDecFromStr(p.Beta)
	if err != nil {
		return nil
	}
	limit := math.LegacyNewDec(2)
	if alpha.Mul(beta).GT(limit) {
		return fmt.Errorf("Alpha × Beta (%s) must be <= 2.0", alpha.Mul(beta).String())
	}
	return nil
}

func validateDecRange(s, name string, min, max float64) error {
	if s == "" {
		return nil
	}
	v, err := math.LegacyNewDecFromStr(s)
	if err != nil {
		return fmt.Errorf("%s: invalid decimal %q", name, s)
	}
	minDec := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", min))
	maxDec := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", max))
	if v.LT(minDec) || v.GT(maxDec) {
		return fmt.Errorf("%s (%s) must be in [%s, %s]", name, s, minDec.String(), maxDec.String())
	}
	return nil
}
