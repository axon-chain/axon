package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

var (
	DefaultAlpha       = math.LegacyMustNewDecFromStr("0.5")
	DefaultBeta        = math.LegacyMustNewDecFromStr("1.5")
	DefaultRMax  int64 = 100

	decOne = math.LegacyOneDec()
	decTwo = math.LegacyNewDec(2)

	// ln(101) pre-computed to 18-digit precision for determinism
	ln101 = math.LegacyMustNewDecFromStr("4.615120516934242563")

	maxMiningPower = math.LegacyNewDec(1_000_000)
)

// ApproxSqrt computes √x via Newton's method (20 iterations).
// Panics if x < 0. Returns zero for x == 0.
func ApproxSqrt(x math.LegacyDec) math.LegacyDec {
	if x.IsNegative() {
		panic("ApproxSqrt: negative input")
	}
	if x.IsZero() {
		return math.LegacyZeroDec()
	}

	// Initial guess: x/2, clamped to at least 1
	guess := x.Quo(decTwo)
	if guess.IsZero() {
		guess = decOne
	}

	for i := 0; i < 20; i++ {
		// guess = (guess + x/guess) / 2
		guess = guess.Add(x.Quo(guess)).Quo(decTwo)
	}
	return guess
}

// ApproxLn computes ln(x) for x > 0 using the identity
//
//	ln(x) = 2 * arctanh((x-1)/(x+1)) = 2 * Σ_{k=0}^{N} (t^(2k+1))/(2k+1)
//
// where t = (x-1)/(x+1), |t| < 1 for x > 0.
// Uses 30 terms for accuracy at 18 decimal places.
func ApproxLn(x math.LegacyDec) math.LegacyDec {
	if !x.IsPositive() {
		panic("ApproxLn: non-positive input")
	}
	if x.Equal(decOne) {
		return math.LegacyZeroDec()
	}

	// Range reduction: ln(x) = ln(m * 2^e) = ln(m) + e*ln(2)
	// Bring x into [0.5, 2) for better series convergence
	var exponent int64
	reduced := x
	ln2 := math.LegacyMustNewDecFromStr("0.693147180559945309")

	for reduced.GT(decTwo) {
		reduced = reduced.Quo(decTwo)
		exponent++
	}
	half := math.LegacyMustNewDecFromStr("0.5")
	for reduced.LT(half) {
		reduced = reduced.Mul(decTwo)
		exponent--
	}

	t := reduced.Sub(decOne).Quo(reduced.Add(decOne))
	t2 := t.Mul(t)

	sum := math.LegacyZeroDec()
	tPow := t // t^1

	for k := 0; k < 30; k++ {
		denom := math.LegacyNewDec(int64(2*k + 1))
		sum = sum.Add(tPow.Quo(denom))
		tPow = tPow.Mul(t2)
	}

	result := sum.Mul(decTwo)
	if exponent != 0 {
		result = result.Add(ln2.MulInt64(exponent))
	}
	return result
}

// ApproxPow computes base^exp = e^(exp * ln(base)) using Taylor series for exp().
// base must be positive. Returns 1 for exp==0, base for exp==1.
func ApproxPow(base, exp math.LegacyDec) math.LegacyDec {
	if base.IsZero() {
		return math.LegacyZeroDec()
	}
	if exp.IsZero() {
		return decOne
	}
	if exp.Equal(decOne) {
		return base
	}
	// Special-case sqrt for alpha=0.5
	if exp.Equal(DefaultAlpha) {
		return ApproxSqrt(base)
	}

	lnBase := ApproxLn(base)
	y := exp.Mul(lnBase) // y = exp * ln(base)

	// e^y via Taylor series: Σ y^k / k!, 40 terms
	sum := decOne
	term := decOne
	for k := int64(1); k <= 40; k++ {
		term = term.Mul(y).Quo(math.LegacyNewDec(k))
		sum = sum.Add(term)
	}
	return sum
}

// CalcMiningPower computes:
//
//	MiningPower = StakeScore × ReputationScore
//	StakeScore  = stake ^ alpha
//	ReputationScore = 1 + beta * ln(1 + clamp(reputation, 0, rMax)) / ln(rMax + 1)
//
// reputation is clamped to [0, rMax]. stake=0 yields 0.
func CalcMiningPower(stake math.Int, reputation int64, alpha, beta math.LegacyDec, rMax int64) math.LegacyDec {
	if stake.IsZero() || !stake.IsPositive() {
		return math.LegacyZeroDec()
	}
	if rMax <= 0 {
		rMax = DefaultRMax
	}

	stakeDec := math.LegacyNewDecFromInt(stake)
	stakeScore := ApproxPow(stakeDec, alpha)

	rep := reputation
	if rep < 0 {
		rep = 0
	}
	if rep > rMax {
		rep = rMax
	}

	// ReputationScore = 1 + beta * ln(1 + rep) / ln(rMax + 1)
	repDec := math.LegacyNewDec(1 + rep)
	lnRep := ApproxLn(repDec)
	logDen := ln101
	if rMax != 100 {
		logDen = ApproxLn(math.LegacyNewDec(rMax + 1))
	}
	reputationScore := decOne.Add(beta.Mul(lnRep).Quo(logDen))

	return stakeScore.Mul(reputationScore)
}

// ComputeAllMiningPowers iterates over all agents and returns normalized
// mining power in CometBFT int64 range [1, 1_000_000].
// Reads alpha/beta/rMax from governance params so governance changes take effect.
func (k Keeper) ComputeAllMiningPowers(ctx sdk.Context) map[string]int64 {
	params := k.GetParams(ctx)

	alpha := DefaultAlpha
	beta := DefaultBeta
	rMax := DefaultRMax

	if params.Alpha != "" {
		if a, err := math.LegacyNewDecFromStr(params.Alpha); err == nil {
			alpha = a
		}
	}
	if params.Beta != "" {
		if b, err := math.LegacyNewDecFromStr(params.Beta); err == nil {
			beta = b
		}
	}
	if params.RMax > 0 {
		rMax = params.RMax
	}

	raw := make(map[string]math.LegacyDec)
	maxRaw := math.LegacyZeroDec()

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		stake := agent.StakeAmount.Amount
		if !stake.IsPositive() {
			return false
		}

		totalRepMillis := k.GetTotalReputation(ctx, agent.Address)
		rep := totalRepMillis / 1000
		if rep > rMax {
			rep = rMax
		}

		power := CalcMiningPower(stake, rep, alpha, beta, rMax)
		if power.IsPositive() {
			raw[agent.Address] = power
			if power.GT(maxRaw) {
				maxRaw = power
			}
		}
		return false
	})

	result := make(map[string]int64, len(raw))
	if maxRaw.IsZero() {
		return result
	}

	for addr, p := range raw {
		// Normalize: scaled = p / maxRaw * (maxMiningPower - 1) + 1
		normalized := p.Quo(maxRaw).Mul(maxMiningPower.Sub(decOne)).Add(decOne)

		val := normalized.TruncateInt64()
		if val < 1 {
			val = 1
		}
		if val > 1_000_000 {
			val = 1_000_000
		}
		result[addr] = val
	}

	return result
}

const MiningPowerKeyPrefix = "MiningPower/"

// StoreMiningPowers persists computed mining powers for use in validator updates.
func (k Keeper) StoreMiningPowers(ctx sdk.Context, powers map[string]int64) {
	store := ctx.KVStore(k.storeKey)
	for addr, power := range powers {
		key := []byte(MiningPowerKeyPrefix + addr)
		bz := make([]byte, 8)
		bz[0] = byte(power >> 56)
		bz[1] = byte(power >> 48)
		bz[2] = byte(power >> 40)
		bz[3] = byte(power >> 32)
		bz[4] = byte(power >> 24)
		bz[5] = byte(power >> 16)
		bz[6] = byte(power >> 8)
		bz[7] = byte(power)
		store.Set(key, bz)
	}
}

// GetMiningPower returns the stored mining power for an address.
func (k Keeper) GetMiningPower(ctx sdk.Context, addr string) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(MiningPowerKeyPrefix + addr))
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return int64(bz[0])<<56 | int64(bz[1])<<48 | int64(bz[2])<<40 | int64(bz[3])<<32 |
		int64(bz[4])<<24 | int64(bz[5])<<16 | int64(bz[6])<<8 | int64(bz[7])
}
