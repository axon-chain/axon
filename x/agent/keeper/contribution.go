package keeper

import (
	"encoding/binary"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

const (
	// ContributionBaseYearlyStr = ~35M AXON/year in aaxon for Year 1-4
	ContributionBaseYearlyStr = "35000000000000000000000000"

	// MaxContributionSupplyStr: hard cap = 350,000,000 AXON = 350M × 10^18 aaxon
	// Whitepaper §9.4: Agent 贡献奖励 35% = 350,000,000 AXON
	MaxContributionSupplyStr = "350000000000000000000000000"

	// ContributionPhaseBlocks = 4 years per phase
	ContributionPhaseBlocks int64 = BlocksPerYear * 4

	// MinReputationForReward — agents with rep < 20 don't participate
	MinReputationForReward uint64 = 20

	// MinRegistrationBlocks — agents registered < 7 days (120960 blocks) don't participate
	MinRegistrationBlocks int64 = 120960

	// Contribution weights
	WeightDeployContract int64 = 50
	WeightContractCalled int64 = 30
	WeightTxActivity     int64 = 10
	WeightHighReputation int64 = 10

	maxContributionCounter uint64 = 10000
)

// MintContributionRewards mints tokens for the contribution pool each block.
// Hard-capped at 350M AXON total (whitepaper §9.4).
func (k Keeper) MintContributionRewards(ctx sdk.Context) {
	blockHeight := ctx.BlockHeight()
	if blockHeight <= 1 {
		return
	}

	perBlock := calculateContributionPerBlock(blockHeight)
	if perBlock.IsZero() {
		return
	}

	// Enforce 350M hard cap
	maxSupply, _ := new(big.Int).SetString(MaxContributionSupplyStr, 10)
	totalMinted := k.GetTotalContributionMinted(ctx)
	remaining := sdkmath.NewIntFromBigInt(new(big.Int).Sub(maxSupply, totalMinted.BigInt()))
	if !remaining.IsPositive() {
		return
	}
	if perBlock.GT(remaining) {
		perBlock = remaining
	}

	coin := sdk.NewCoin("aaxon", perBlock)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		k.Logger(ctx).Error("failed to mint contribution rewards", "error", err)
		return
	}

	k.addTotalContributionMinted(ctx, perBlock)
	k.addToContributionPool(ctx, coin)
}

// --- Supply cap tracking ---

func (k Keeper) GetTotalContributionMinted(ctx sdk.Context) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.TotalContributionMintedKey))
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	var amount sdkmath.Int
	if err := amount.Unmarshal(bz); err != nil {
		return sdkmath.ZeroInt()
	}
	return amount
}

func (k Keeper) SetTotalContributionMinted(ctx sdk.Context, total sdkmath.Int) {
	bz, err := total.Marshal()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal TotalContributionMinted: %v", err))
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.TotalContributionMintedKey), bz)
}

func (k Keeper) addTotalContributionMinted(ctx sdk.Context, amount sdkmath.Int) {
	total := k.GetTotalContributionMinted(ctx).Add(amount)
	k.SetTotalContributionMinted(ctx, total)
}

// calculateContributionPerBlock returns per-block contribution reward matching
// the whitepaper §8.4 custom declining schedule (NOT equal halving):
//
//	Year 1-4:  35M AXON/year
//	Year 5-8:  25M AXON/year
//	Year 9-12: 15M AXON/year
//	Year 12+:   5M AXON/year (long tail until 350M cap)
func calculateContributionPerBlock(blockHeight int64) sdkmath.Int {
	year := blockHeight / BlocksPerYear
	var yearlyStr string
	switch {
	case year < 4:
		yearlyStr = "35000000000000000000000000" // 35M AXON
	case year < 8:
		yearlyStr = "25000000000000000000000000" // 25M AXON
	case year < 12:
		yearlyStr = "15000000000000000000000000" // 15M AXON
	default:
		yearlyStr = "5000000000000000000000000" // 5M AXON
	}

	yearly, ok := new(big.Int).SetString(yearlyStr, 10)
	if !ok {
		return sdkmath.ZeroInt()
	}

	perBlock := new(big.Int).Div(yearly, big.NewInt(BlocksPerYear))
	if perBlock.Sign() <= 0 {
		return sdkmath.ZeroInt()
	}

	return sdkmath.NewIntFromBigInt(perBlock)
}

// DistributeContributionRewards distributes accumulated contribution pool at epoch boundary.
func (k Keeper) DistributeContributionRewards(ctx sdk.Context, epoch uint64) {
	pool := k.getContributionPool(ctx)
	if pool.IsZero() {
		return
	}

	type scoredAgent struct {
		address string
		score   int64
		stake   *big.Int
	}

	var agents []scoredAgent
	totalScore := int64(0)
	totalEligibleStake := new(big.Int)
	currentBlock := ctx.BlockHeight()

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_SUSPENDED {
			return false
		}

		// Filter: reputation >= 20
		if agent.Reputation < MinReputationForReward {
			return false
		}

		// Filter: registered >= 7 days
		if currentBlock-agent.RegisteredAt < MinRegistrationBlocks {
			return false
		}

		score := k.calculateContributionScore(ctx, epoch, agent)
		if score <= 0 {
			return false
		}

		stake := agent.StakeAmount.Amount.BigInt()
		if stake.Sign() <= 0 {
			return false
		}

		agents = append(agents, scoredAgent{address: agent.Address, score: score, stake: new(big.Int).Set(stake)})
		totalScore += score
		totalEligibleStake.Add(totalEligibleStake, stake)
		return false
	})

	if totalScore <= 0 || len(agents) == 0 || totalEligibleStake.Sign() <= 0 {
		return
	}

	params := k.GetParams(ctx)
	capBps := int64(params.GetContributionCapBps())

	poolBig := pool.Amount.BigInt()
	distributed := sdkmath.ZeroInt()

	for _, a := range agents {
		share := new(big.Int).Mul(poolBig, big.NewInt(a.score))
		share.Div(share, big.NewInt(totalScore))

		maxPerAgent := contributionRewardCap(poolBig, a.stake, totalEligibleStake, capBps)

		if share.Cmp(maxPerAgent) > 0 {
			share.Set(maxPerAgent)
		}

		reward := sdkmath.NewIntFromBigInt(share)
		if reward.IsZero() {
			continue
		}

		addr, err := sdk.AccAddressFromBech32(a.address)
		if err != nil {
			continue
		}

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin("aaxon", reward))); err != nil {
			k.Logger(ctx).Error("failed to send contribution reward", "address", a.address, "error", err)
			continue
		}
		distributed = distributed.Add(reward)
	}

	remaining := pool.Sub(sdk.NewCoin("aaxon", distributed))
	if remaining.IsPositive() {
		k.setContributionPool(ctx, remaining)
	} else {
		k.setContributionPool(ctx, sdk.NewInt64Coin("aaxon", 0))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"contribution_rewards_distributed",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("pool", pool.String()),
		sdk.NewAttribute("distributed", distributed.String()),
		sdk.NewAttribute("agents", fmt.Sprintf("%d", len(agents))),
	))
}

// calculateContributionScore computes a weighted score for an agent.
func (k Keeper) calculateContributionScore(ctx sdk.Context, epoch uint64, agent types.Agent) int64 {
	score := int64(0)

	// Contract deployments
	deploys := k.getCounter(ctx, types.KeyDeployCount(epoch, agent.Address))
	if k.IsV110UpgradeActivated(ctx) && deploys > maxContributionCounter {
		deploys = maxContributionCounter
	}
	score += int64(deploys) * WeightDeployContract

	// Contracts called by others (popularity)
	calls := k.getCounter(ctx, types.KeyContractCall(epoch, agent.Address))
	if k.IsV110UpgradeActivated(ctx) && calls > maxContributionCounter {
		calls = maxContributionCounter
	}
	score += int64(calls) * WeightContractCalled

	// Transaction activity
	activity := k.GetEpochActivity(ctx, epoch, agent.Address)
	activityCapped := activity
	if activityCapped > 100 {
		activityCapped = 100
	}
	score += int64(activityCapped) * WeightTxActivity

	// High reputation bonus (> 70)
	if agent.Reputation > 70 {
		score += int64(agent.Reputation-70) * WeightHighReputation
	}

	// Online bonus
	if agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE {
		score += 5
	}

	return score
}

func contributionRewardCap(poolAmount, agentStake, totalEligibleStake *big.Int, capBps int64) *big.Int {
	if poolAmount == nil || agentStake == nil || totalEligibleStake == nil {
		return new(big.Int)
	}
	if poolAmount.Sign() <= 0 || agentStake.Sign() <= 0 || totalEligibleStake.Sign() <= 0 {
		return new(big.Int)
	}
	if capBps <= 0 {
		capBps = 200
	}

	capAmount := new(big.Int).Mul(poolAmount, big.NewInt(capBps))
	capAmount.Mul(capAmount, agentStake)
	capAmount.Div(capAmount, big.NewInt(10000))
	capAmount.Div(capAmount, totalEligibleStake)
	return capAmount
}

// Tracking helpers for contribution metrics

func (k Keeper) IncrementDeployCount(ctx sdk.Context, address string) {
	epoch := k.GetCurrentEpoch(ctx)
	key := types.KeyDeployCount(epoch, address)
	k.incrementCounter(ctx, key)
}

func (k Keeper) IncrementContractCalls(ctx sdk.Context, address string) {
	epoch := k.GetCurrentEpoch(ctx)
	key := types.KeyContractCall(epoch, address)
	k.incrementCounter(ctx, key)
}

func (k Keeper) incrementCounter(ctx sdk.Context, key []byte) {
	store := ctx.KVStore(k.storeKey)
	count := uint64(0)
	bz := store.Get(key)
	if bz != nil && len(bz) >= 8 {
		count = binary.BigEndian.Uint64(bz)
	}
	count++
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(key, bz)
}

func (k Keeper) getCounter(ctx sdk.Context, key []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz == nil || len(bz) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// Contribution pool management

func (k Keeper) getContributionPool(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.ContributionPoolKey))
	if bz == nil {
		return sdk.NewInt64Coin("aaxon", 0)
	}
	var coin sdk.Coin
	k.cdc.MustUnmarshal(bz, &coin)
	return coin
}

func (k Keeper) setContributionPool(ctx sdk.Context, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&amount)
	store.Set([]byte(types.ContributionPoolKey), bz)
}

func (k Keeper) addToContributionPool(ctx sdk.Context, amount sdk.Coin) {
	current := k.getContributionPool(ctx)
	k.setContributionPool(ctx, current.Add(amount))
}

func (k Keeper) GetContributionPool(ctx sdk.Context) sdk.Coin {
	return k.getContributionPool(ctx)
}

func (k Keeper) SetContributionPool(ctx sdk.Context, coin sdk.Coin) {
	k.setContributionPool(ctx, coin)
}
