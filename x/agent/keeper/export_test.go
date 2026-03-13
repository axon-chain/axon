package keeper

import (
	sdkmath "cosmossdk.io/math"

	"github.com/axon-chain/axon/x/agent/types"
)

func GetChallengePoolSize() []struct{} {
	return make([]struct{}, len(challengePool))
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

func DetectCheatersForTest(k Keeper, responses []types.AIResponse) map[string]bool {
	return k.detectCheaters(responses)
}
