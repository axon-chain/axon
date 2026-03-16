package types

import (
	"fmt"
	"math/big"
)

const (
	DefaultMaxShieldAmount      = "10000000000000000000000" // 10000 × 10^18 aaxon
	DefaultPoolCapRatio         = 10                        // 10% = 0.10
	DefaultPrivacyGasMultiplier = 200                       // 2.00 in basis points / 100
	DefaultVKRegistrationFee    = "100000000000000000000"   // 100 × 10^18 aaxon
	DefaultMaxKnownRoots        = 100
	TreeDepth                   = 32
)

type Params struct {
	MaxShieldAmount      string `json:"max_shield_amount"`
	PoolCapRatio         uint64 `json:"pool_cap_ratio"`         // basis points (10 = 0.10)
	PrivacyGasMultiplier uint64 `json:"privacy_gas_multiplier"` // ×100
	VKRegistrationFee    string `json:"vk_registration_fee"`
	MaxKnownRoots        uint64 `json:"max_known_roots"`
}

func DefaultParams() Params {
	return Params{
		MaxShieldAmount:      DefaultMaxShieldAmount,
		PoolCapRatio:         uint64(DefaultPoolCapRatio),
		PrivacyGasMultiplier: uint64(DefaultPrivacyGasMultiplier),
		VKRegistrationFee:    DefaultVKRegistrationFee,
		MaxKnownRoots:        uint64(DefaultMaxKnownRoots),
	}
}

func (p Params) Validate() error {
	if p.MaxKnownRoots == 0 {
		return fmt.Errorf("MaxKnownRoots must be > 0")
	}
	if p.MaxKnownRoots > 10000 {
		return fmt.Errorf("MaxKnownRoots must be <= 10000")
	}
	if p.PoolCapRatio > 100 {
		return fmt.Errorf("PoolCapRatio must be <= 100 (100%%)")
	}
	if p.PrivacyGasMultiplier < 100 || p.PrivacyGasMultiplier > 1000 {
		return fmt.Errorf("PrivacyGasMultiplier (%d) must be in [100, 1000]", p.PrivacyGasMultiplier)
	}
	if p.MaxShieldAmount == "" {
		return fmt.Errorf("MaxShieldAmount must not be empty")
	}
	if amt, ok := new(big.Int).SetString(p.MaxShieldAmount, 10); !ok || amt.Sign() <= 0 {
		return fmt.Errorf("MaxShieldAmount must be a positive integer, got %q", p.MaxShieldAmount)
	}
	if p.VKRegistrationFee == "" {
		return fmt.Errorf("VKRegistrationFee must not be empty")
	}
	if fee, ok := new(big.Int).SetString(p.VKRegistrationFee, 10); !ok || fee.Sign() <= 0 {
		return fmt.Errorf("VKRegistrationFee must be a positive integer, got %q", p.VKRegistrationFee)
	}
	return nil
}
