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

	// ProposerShare = 25%
	ProposerSharePercent = 25
	// ValidatorPoolShare = 50%
	ValidatorPoolSharePercent = 50
	// AIPerformanceShare = 25%
	AIPerformanceSharePercent = 25
)

// DistributeBlockRewards mints and distributes block rewards.
// Hard-capped at 650M AXON total (whitepaper §8.2).
func (k Keeper) DistributeBlockRewards(ctx sdk.Context) {
	blockHeight := ctx.BlockHeight()
	if blockHeight <= 1 {
		return
	}

	reward := calculateBlockReward(blockHeight)
	if reward.IsZero() {
		return
	}

	// Enforce 650M hard cap
	maxSupply, _ := new(big.Int).SetString(MaxBlockRewardSupplyStr, 10)
	totalMinted := k.GetTotalBlockRewardsMinted(ctx)
	remaining := sdkmath.NewIntFromBigInt(new(big.Int).Sub(maxSupply, totalMinted.BigInt()))
	if !remaining.IsPositive() {
		return
	}
	if reward.GT(remaining) {
		reward = remaining
	}

	rewardCoin := sdk.NewCoin("aaxon", reward)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(rewardCoin)); err != nil {
		k.Logger(ctx).Error("failed to mint block rewards", "error", err)
		return
	}

	k.addTotalBlockRewardsMinted(ctx, reward)

	proposerReward := reward.Mul(sdkmath.NewInt(ProposerSharePercent)).Quo(sdkmath.NewInt(100))
	validatorReward := reward.Mul(sdkmath.NewInt(ValidatorPoolSharePercent)).Quo(sdkmath.NewInt(100))
	aiReward := reward.Sub(proposerReward).Sub(validatorReward)

	k.distributeProposerReward(ctx, proposerReward)
	k.distributeValidatorRewards(ctx, validatorReward)
	k.distributeAIPerformanceRewards(ctx, aiReward)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"block_rewards",
		sdk.NewAttribute("height", fmt.Sprintf("%d", blockHeight)),
		sdk.NewAttribute("total", rewardCoin.String()),
		sdk.NewAttribute("proposer", proposerReward.String()),
		sdk.NewAttribute("validators", validatorReward.String()),
		sdk.NewAttribute("ai_performance", aiReward.String()),
	))
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

// distributeProposerReward sends 25% to the block proposer.
func (k Keeper) distributeProposerReward(ctx sdk.Context, amount sdkmath.Int) {
	if amount.IsZero() {
		return
	}

	proposerConsAddr := ctx.BlockHeader().ProposerAddress
	if len(proposerConsAddr) == 0 {
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", amount))
		return
	}

	// Look up the validator by consensus address to get the operator address
	validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(proposerConsAddr))
	if err != nil {
		k.Logger(ctx).Error("failed to find proposer validator", "error", err)
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", amount))
		return
	}

	// Convert validator operator address to account address for reward
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		k.Logger(ctx).Error("failed to parse validator operator address", "error", err)
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", amount))
		return
	}
	accAddr := sdk.AccAddress(valAddr)

	coins := sdk.NewCoins(sdk.NewCoin("aaxon", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddr, coins); err != nil {
		k.Logger(ctx).Error("failed to send proposer reward", "error", err)
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", amount))
	}
}

// reputationBonusPercent returns the tiered reputation bonus per whitepaper §7.3.
func reputationBonusPercent(reputation uint64) int64 {
	switch {
	case reputation >= 90:
		return 20
	case reputation >= 70:
		return 15
	case reputation >= 50:
		return 10
	case reputation >= 30:
		return 5
	default:
		return 0
	}
}

// distributeValidatorRewards distributes 50% to active validators by weight.
// Weight = Stake × (100 + ReputationBonus + AIBonus) per whitepaper §7.3.
// Falls back to staking validators weighted by bonded tokens when no Agents are online.
func (k Keeper) distributeValidatorRewards(ctx sdk.Context, totalAmount sdkmath.Int) {
	if totalAmount.IsZero() {
		return
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
		stake := agent.StakeAmount.Amount.BigInt()
		repBonus := reputationBonusPercent(agent.Reputation)
		aiBonus := k.GetAIBonus(ctx, agent.Address)
		multiplier := int64(100) + repBonus + aiBonus
		if multiplier < 10 {
			multiplier = 10
		}
		w := new(big.Int).Mul(stake, big.NewInt(multiplier))
		totalWeight.Add(totalWeight, w)

		addr, err := sdk.AccAddressFromBech32(agent.Address)
		if err != nil {
			return false
		}
		validators = append(validators, validatorWeight{accAddr: addr, weight: w})
		return false
	})

	// Fallback: if no online Agents, distribute to bonded staking validators
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
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", totalAmount))
		return
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
	if remainder.IsPositive() {
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", remainder))
	}
}

// distributeAIPerformanceRewards distributes 25% by AI challenge scores.
// If no agents have AI scores, the reward goes to the reward pool for later epoch distribution.
func (k Keeper) distributeAIPerformanceRewards(ctx sdk.Context, totalAmount sdkmath.Int) {
	if totalAmount.IsZero() {
		return
	}

	type aiWeight struct {
		address string
		bonus   int64
	}

	var agents []aiWeight
	totalBonus := int64(0)

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status != types.AgentStatus_AGENT_STATUS_ONLINE {
			return false
		}
		bonus := k.GetAIBonus(ctx, agent.Address)
		if bonus <= 0 {
			return false
		}
		agents = append(agents, aiWeight{address: agent.Address, bonus: bonus})
		totalBonus += bonus
		return false
	})

	if totalBonus <= 0 || len(agents) == 0 {
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", totalAmount))
		return
	}

	totalBig := totalAmount.BigInt()
	distributed := sdkmath.ZeroInt()

	for _, a := range agents {
		share := new(big.Int).Mul(totalBig, big.NewInt(a.bonus))
		share.Div(share, big.NewInt(totalBonus))
		reward := sdkmath.NewIntFromBigInt(share)
		if reward.IsZero() {
			continue
		}

		addr, err := sdk.AccAddressFromBech32(a.address)
		if err != nil {
			continue
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin("aaxon", reward))); err != nil {
			k.Logger(ctx).Error("failed to send AI performance reward", "address", a.address, "error", err)
			continue
		}
		distributed = distributed.Add(reward)
	}

	remainder := totalAmount.Sub(distributed)
	if remainder.IsPositive() {
		k.AddToRewardPool(ctx, sdk.NewCoin("aaxon", remainder))
	}
}
