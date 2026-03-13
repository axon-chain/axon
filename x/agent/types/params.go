package types

import "fmt"

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
	return nil
}
