package keeper

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

const (
	// BlocksPerYear at 5s/block = 6,307,200
	BlocksPerYear int64 = 6_307_200

	// HalvingInterval = 4 years in blocks
	HalvingInterval int64 = BlocksPerYear * 4

	// BaseBlockReward in aaxon: ~12.367 AXON/block for year 1-4 → ~78M/year
	// 78M * 1e18 / 6_307_200 ≈ 12.367e18 aaxon/block
	BaseBlockRewardStr = "12367000000000000000"

	// MaxBlockRewardSupplyStr: hard cap = 650,000,000 AXON = 650M × 10^18 aaxon
	// Whitepaper §8.2: 区块奖励（验证者挖矿）65% = 650,000,000 AXON
	MaxBlockRewardSupplyStr = "650000000000000000000000000"

	// M4: pool ratios — Proposer 20%, Validator 55%, Reputation 25%
	ProposerSharePercent       = 20
	ValidatorPoolSharePercent  = 55
	ReputationPoolSharePercent = 25
)

// DistributeBlockRewards is kept for backward compatibility but is now a no-op.
// Block reward minting and proposer distribution is handled by AccumulateBlockReward
// in abci.go. Validator and AI pool distribution happens at epoch boundaries via
// DistributeAccumulatedBlockRewards.
func (k Keeper) DistributeBlockRewards(ctx sdk.Context) {
	// Intentionally empty — replaced by AccumulateBlockReward (F5 fix).
	// This stub prevents compile errors if any external code still calls it.
}

// --- Supply cap tracking ---

func (k Keeper) GetTotalBlockRewardsMinted(ctx sdk.Context) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.TotalBlockRewardsMintedKey))
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	var amount sdkmath.Int
	if err := amount.Unmarshal(bz); err != nil {
		return sdkmath.ZeroInt()
	}
	return amount
}

func (k Keeper) SetTotalBlockRewardsMinted(ctx sdk.Context, total sdkmath.Int) {
	bz, err := total.Marshal()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal TotalBlockRewardsMinted: %v", err))
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.TotalBlockRewardsMintedKey), bz)
}

func (k Keeper) addTotalBlockRewardsMinted(ctx sdk.Context, amount sdkmath.Int) {
	total := k.GetTotalBlockRewardsMinted(ctx).Add(amount)
	k.SetTotalBlockRewardsMinted(ctx, total)
}

// calculateBlockReward returns the per-block reward accounting for halvings.
func calculateBlockReward(blockHeight int64) sdkmath.Int {
	baseReward, ok := new(big.Int).SetString(BaseBlockRewardStr, 10)
	if !ok {
		return sdkmath.ZeroInt()
	}

	halvings := blockHeight / HalvingInterval
	if halvings >= 64 {
		return sdkmath.ZeroInt()
	}

	// Right-shift to apply halving: reward = base >> halvings
	reward := new(big.Int).Rsh(baseReward, uint(halvings))
	if reward.Sign() <= 0 {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromBigInt(reward)
}

// distributeProposerReward sends 20% to the block proposer and returns any
// amount that could not be delivered so it can stay in the v2 validator pool.
func (k Keeper) distributeProposerReward(ctx sdk.Context, amount sdkmath.Int) sdkmath.Int {
	if amount.IsZero() {
		return sdkmath.ZeroInt()
	}

	proposerConsAddr := ctx.BlockHeader().ProposerAddress
	if len(proposerConsAddr) == 0 {
		return amount
	}

	// Look up the validator by consensus address to get the operator address
	validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(proposerConsAddr))
	if err != nil {
		k.Logger(ctx).Error("failed to find proposer validator", "error", err)
		return amount
	}

	// Convert validator operator address to account address for reward
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		k.Logger(ctx).Error("failed to parse validator operator address", "error", err)
		return amount
	}
	accAddr := sdk.AccAddress(valAddr)

	coins := sdk.NewCoins(sdk.NewCoin("aaxon", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddr, coins); err != nil {
		k.Logger(ctx).Error("failed to send proposer reward", "error", err)
		return amount
	}
	return sdkmath.ZeroInt()
}

// distributeValidatorRewards distributes 55% to active bonded validators weighted
// by v2 MiningPower. Falls back to bonded staking tokens when no eligible agents.
func (k Keeper) distributeValidatorRewards(ctx sdk.Context, totalAmount sdkmath.Int) sdkmath.Int {
	if totalAmount.IsZero() {
		return sdkmath.ZeroInt()
	}

	type validatorWeight struct {
		accAddr sdk.AccAddress
		weight  *big.Int
	}

	var validators []validatorWeight
	totalWeight := new(big.Int)

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status != types.AgentStatus_AGENT_STATUS_ONLINE {
			return false
		}
		if !k.isActiveValidatorAddress(ctx, agent.Address) {
			return false
		}

		mp := k.GetMiningPower(ctx, agent.Address)
		if mp <= 0 {
			mp = 1
		}
		w := big.NewInt(mp)
		totalWeight.Add(totalWeight, w)

		addr, err := sdk.AccAddressFromBech32(agent.Address)
		if err != nil {
			return false
		}
		validators = append(validators, validatorWeight{accAddr: addr, weight: w})
		return false
	})

	if len(validators) == 0 {
		bondedVals, err := k.stakingKeeper.GetBondedValidatorsByPower(ctx)
		if err == nil && len(bondedVals) > 0 {
			for _, val := range bondedVals {
				tokens := val.GetTokens().BigInt()
				if tokens.Sign() <= 0 {
					continue
				}
				totalWeight.Add(totalWeight, tokens)

				valAddr, err := sdk.ValAddressFromBech32(val.OperatorAddress)
				if err != nil {
					continue
				}
				validators = append(validators, validatorWeight{
					accAddr: sdk.AccAddress(valAddr),
					weight:  tokens,
				})
			}
		}
	}

	if totalWeight.Sign() <= 0 || len(validators) == 0 {
		return totalAmount
	}

	totalBig := totalAmount.BigInt()
	distributed := sdkmath.ZeroInt()

	for _, v := range validators {
		share := new(big.Int).Mul(totalBig, v.weight)
		share.Div(share, totalWeight)
		reward := sdkmath.NewIntFromBigInt(share)
		if reward.IsZero() {
			continue
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, v.accAddr, sdk.NewCoins(sdk.NewCoin("aaxon", reward))); err != nil {
			k.Logger(ctx).Error("failed to send validator reward", "address", v.accAddr.String(), "error", err)
			continue
		}
		distributed = distributed.Add(reward)
	}

	remainder := totalAmount.Sub(distributed)
	return remainder
}

// distributeReputationRewards distributes 25% (M4 Reputation Pool) to all registered
// Agents proportional to their ReputationScore. Unlike the old AI Performance Pool,
// this is open to ALL registered agents (not just validators) and weights by reputation
// only (not stake), incentivizing reputation accumulation.
func (k Keeper) distributeReputationRewards(ctx sdk.Context, totalAmount sdkmath.Int) sdkmath.Int {
	if totalAmount.IsZero() {
		return sdkmath.ZeroInt()
	}

	type repWeight struct {
		address string
		weight  int64
	}

	var agents []repWeight
	totalWeight := int64(0)

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
			return false
		}
		totalRep := k.GetTotalReputation(ctx, agent.Address)
		if totalRep <= 0 {
			return false
		}
		repScore := k.calcReputationScoreMillis(ctx, totalRep)
		if repScore <= 0 {
			return false
		}
		agents = append(agents, repWeight{address: agent.Address, weight: repScore})
		totalWeight += repScore
		return false
	})

	if totalWeight <= 0 || len(agents) == 0 {
		return totalAmount
	}

	totalBig := totalAmount.BigInt()
	distributed := sdkmath.ZeroInt()
	totalWeightBig := big.NewInt(totalWeight)

	for _, a := range agents {
		share := new(big.Int).Mul(totalBig, big.NewInt(a.weight))
		share.Div(share, totalWeightBig)
		reward := sdkmath.NewIntFromBigInt(share)
		if reward.IsZero() {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(a.address)
		if err != nil {
			continue
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin("aaxon", reward))); err != nil {
			k.Logger(ctx).Error("failed to send reputation reward", "address", a.address, "error", err)
			continue
		}
		distributed = distributed.Add(reward)
	}

	remainder := totalAmount.Sub(distributed)
	return remainder
}

// calcReputationScoreMillis converts milliscored total reputation to the
// ReputationScore multiplier (×1000) using the same log formula as mining power:
//
//	RepScore = 1 + beta * ln(1 + rep) / ln(rMax + 1)
//
// Reads beta and rMax from governance params. Returns value in range [1000, ~2500] (×1000).
func (k Keeper) calcReputationScoreMillis(ctx sdk.Context, totalRepMillis int64) int64 {
	if totalRepMillis <= 0 {
		return 1000
	}

	params := k.GetParams(ctx)
	beta := DefaultBeta
	rMax := DefaultRMax
	if params.Beta != "" {
		if b, err := sdkmath.LegacyNewDecFromStr(params.Beta); err == nil {
			beta = b
		}
	}
	if params.RMax > 0 {
		rMax = int64(params.RMax)
	}

	rep := totalRepMillis / 1000
	if rep > rMax {
		rep = rMax
	}
	if rep <= 0 {
		return 1000
	}

	repDec := sdkmath.LegacyNewDec(1 + rep)
	lnRep := ApproxLn(repDec)
	logDen := ln101
	if rMax != 100 {
		logDen = ApproxLn(sdkmath.LegacyNewDec(rMax + 1))
	}
	score := decOne.Add(beta.Mul(lnRep).Quo(logDen))

	result := score.MulInt64(1000).TruncateInt64()
	if result < 1000 {
		result = 1000
	}
	return result
}
