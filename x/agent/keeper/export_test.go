package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func GetChallengePoolSize() []struct{} {
	return make([]struct{}, len(challengePool))
}

func GetLegacyChallengePoolSize() int {
	return len(challengePool)
}

func GetChallengeTemplateSize() int {
	return len(challengeTemplates)
}

func NormalizeAnswerForTest(s string) string {
	return normalizeAnswer(s)
}

func ScoreResponseForTest(reveal, answer string) int {
	return scoreResponse(types.AIResponse{RevealData: reveal}, answer)
}

func CalculateAIBonusForTest(score int) int64 {
	return calculateAIBonus(score)
}

func ExportCalculateBlockReward(blockHeight int64) sdkmath.Int {
	return calculateBlockReward(blockHeight)
}

func ExportCalculateContributionPerBlock(blockHeight int64) sdkmath.Int {
	return calculateContributionPerBlock(blockHeight)
}

func ReputationBonusPercentForTest(reputation uint64) int64 {
	return reputationBonusPercent(reputation)
}

func DetectCheatersForTest(k Keeper, responses []types.AIResponse, expectedHash string) map[string]bool {
	return k.detectCheaters(responses, expectedHash)
}

func ContributionRewardCapForTest(poolAmount, agentStake, totalEligibleStake *big.Int) *big.Int {
	return contributionRewardCap(poolAmount, agentStake, totalEligibleStake, 200)
}

func IsActiveValidatorAddressForTest(k Keeper, ctx sdk.Context, address string) bool {
	return k.isActiveValidatorAddress(ctx, address)
}

func CalculateCrossScoresForTest(k Keeper, responses []types.AIResponse) map[string]int {
	return k.calculateCrossScores(responses)
}

func CalculateLengthScoreForTest(length int) int {
	return calculateLengthScore(length)
}

func CalculateKeywordScoreForTest(normalized string) int {
	categoryKeywords := map[string][]string{
		"algorithms":        {"o(n)", "complexity", "sort", "search", "tree", "graph", "recursive", "iteration", "divide", "conquer"},
		"data_structures":   {"array", "list", "stack", "queue", "hash", "map", "tree", "heap", "trie", "node"},
		"blockchain":        {"block", "chain", "consensus", "pow", "pos", "bft", "tendermint", "merkle", "gas", "nonce"},
		"math":              {"log", "sqrt", "prime", "fibonacci", "equation", "derivative", "integral", "matrix", "vector"},
		"networking":        {"tcp", "udp", "http", "dns", "latency", "bandwidth", "protocol", "socket", "packet"},
		"databases":         {"sql", "index", "join", "acid", "transaction", "query", "schema", "normalization"},
		"security":          {"encryption", "hash", "signature", "attack", "vulnerability", "authentication", "authorization"},
		"cryptography":     {"aes", "rsa", "ecc", "ecdsa", "sha", "key", "cipher", "encrypt", "decrypt"},
		"machine_learning": {"neural", "gradient", "training", "loss", "epoch", "batch", "accuracy", "precision"},
	}
	return calculateKeywordScore(normalized, categoryKeywords)
}
