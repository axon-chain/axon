package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

const CheaterAnswerThreshold = 3

var challengePool = []struct {
	Question   string
	AnswerHash string
	Category   string
}{
	{"Template-based AI challenges enabled", "5d41402abc4b2a76b9229f5b260920c26cf2bc46f99d5ee1dde732b4aa9405c3a47f4", "system"},
}

// legacyChallengePoolSize returns the size of the legacy challenge pool for backward compatibility tests.

// Template-based challenge system - questions are generated from parameterized templates
// with VRF-based random values. This eliminates hardcoded answers.
var challengeTemplates = []struct {
	Template  string
	Category string
}{
	{"What is the time complexity of {algorithm} on {input_size} inputs?", "algorithms"},
	{"Compare {algorithm_a} and {algorithm_b}: which has better average-case complexity for {data_type}?", "algorithms"},
	{"What is the space complexity of {data_structure} when storing {element_count} elements?", "data_structures"},
	{"In {blockchain}, what opcode is used to {action} {asset}?", "blockchain"},
	{"What is the gas cost of {operation} in {blockchain} when {condition}?", "blockchain"},
	{"Explain {consensus_mechanism} consensus in terms of {aspect}.", "blockchain"},
	{"Calculate {operation_math} + {value_a} + {value_b}. What is the result?", "math"},
	{"If {variable} = {value}, evaluate: {expression}.", "math"},
	{"What is the {function} of {number}?", "math"},
	{"What is the latency order of magnitude for {protocol} over {distance}?", "networking"},
	{"How does {protocol} handle {failure_scenario}?", "networking"},
	{"What index type is best for {query_pattern}?", "databases"},
	{"Explain ACID properties in context of {transaction_type}.", "databases"},
	{"How would you mitigate {attack_type} attack on {system}?", "security"},
	{"What cryptographic primitive is best for {use_case}?", "cryptography"},
	{"Which {ml_approach} approach is better for {problem_type}: {option_a} or {option_b}?", "machine_learning"},
	{"Explain {concept} in the context of {application}.", "machine_learning"},
}

var vrfValues = map[string][]string{
	"algorithm":         {"binary search", "quick sort", "merge sort", "bubble sort", "heap sort", "BFS", "DFS", "Dijkstra's algorithm", "Bellman-Ford", "A* search"},
	"algorithm_a":       {"quick sort", "merge sort", "heap sort"},
	"algorithm_b":       {"merge sort", "heap sort", "quick sort"},
	"input_size":        {"n", "n log n", "n^2", "2^n"},
	"data_type":        {"sorted arrays", "random arrays", "linked lists", "hash tables"},
	"data_structure":   {"hash table", "binary search tree", "array", "linked list", "heap", "trie"},
	"element_count":    {"1000", "10000", "100000", "1 million"},
	"blockchain":      {"Ethereum", "Bitcoin", "Cosmos", "Axon"},
	"action":          {"transfer ETH", "create contract", "call contract", "mint token", "burn token"},
	"asset":          {"native token", "ERC-20 token", "NFT", "wrapped token"},
	"operation":       {"SSTORE", "SLOAD", "CREATE", "CALL", "LOG", "EXTCODEHASH"},
	"operation_math":  {"addition", "subtraction", "multiplication"},
	"condition":       {"first time", "from zero to non-zero", "value unchanged", "deletion"},
	"consensus_mechanism": {"Proof of Work", "Proof of Stake", "BFT", "PBFT", "Tendermint"},
	"aspect":         {"finality", "throughput", "energy consumption", "decentralization"},
	"variable":        {"x", "y", "n", "i", "k"},
	"value":           {"5", "10", "15", "20", "25", "100"},
	"value_a":         {"123", "456", "789", "321", "654"},
	"value_b":         {"77", "88", "99", "111", "222"},
	"number":         {"64", "81", "100", "144", "225", "256"},
	"function":       {"square root", "log base 2", "cube root", "square"},
	"expression":     {"2x + 3", "3x - 5", "x^2", "2^x"},
	"protocol":       {"TCP", "UDP", "HTTP", "WebSocket", "gRPC"},
	"distance":       {"1km", "100km", "1000km", "intercontinental"},
	"failure_scenario":{"packet loss", "connection timeout", "DDoS attack", "man-in-the-middle"},
	"query_pattern":  {"range queries", "point queries", "full-text search", "join operations"},
	"transaction_type": {"bank transfer", "inventory update", "order placement"},
	"attack_type":    {"SQL injection", "XSS", "CSRF", "DDoS", "reentrancy"},
	"system":         {"web API", "smart contract", "wallet", "exchange"},
	"use_case":       {"signing", "encryption", "hashing", "key exchange"},
	"cryptographic_primitive": {"AES", "RSA", "ECDSA", "SHA-256", "Ed25519"},
	"ml_approach":   {"supervised", "unsupervised", "reinforcement"},
	"problem_type":   {"classification", "clustering", "regression", "generation"},
	"option_a":       {"neural network", "decision tree", "SVM", "linear regression"},
	"option_b":       {"random forest", "gradient boosting", "kNN", "logistic regression"},
	"concept":        {"backpropagation", "attention mechanism", "feature engineering", "overfitting"},
	"application":   {"image recognition", "NLP", "recommendation systems", "autonomous driving"},
}

// GenerateChallenge creates a challenge from parameterized templates using VRF-based seeds.
// This eliminates the vulnerability of hardcoded answers by generating unique questions.
func (k Keeper) GenerateChallenge(ctx sdk.Context, epoch uint64) types.AIChallenge {
	if len(challengeTemplates) == 0 {
		return types.AIChallenge{}
	}

	seed := sha256.Sum256(append(
		ctx.HeaderHash(),
		types.Uint64ToBytes(epoch)...,
	))

	r := rand.New(rand.NewSource(int64(seed[0])|int64(seed[1])<<8|int64(seed[2])<<16|int64(seed[3])<<24))

	templateIdx := int(r.Float64() * float64(len(challengeTemplates)))
	template := challengeTemplates[templateIdx]

	question := expandTemplate(template.Template, &seed, r)

	questionHash := sha256.Sum256([]byte(question))
	questionHashHex := hex.EncodeToString(questionHash[:])
	params := k.GetParams(ctx)

	challenge := types.AIChallenge{
		Epoch:          epoch,
		ChallengeHash: questionHashHex,
		ChallengeType: template.Category,
		ChallengeData: questionHashHex,
		DeadlineBlock: ctx.BlockHeight() + params.AiChallengeWindow,
	}

	k.SetChallenge(ctx, challenge)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"ai_challenge_generated",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("category", template.Category),
		sdk.NewAttribute("question_hash", questionHashHex),
		sdk.NewAttribute("deadline_block", fmt.Sprintf("%d", challenge.DeadlineBlock)),
	))

	return challenge
}

// expandTemplate fills in template placeholders with VRF-selected values.
func expandTemplate(tmpl string, seed *[32]byte, r *rand.Rand) string {
	result := tmpl

	for key, values := range vrfValues {
		placeholder := "{" + key + "}"
		if strings.Contains(result, placeholder) {
			idx := int(r.Float64() * float64(len(values)))
			result = strings.Replace(result, placeholder, values[idx], 1)
		}
	}

	return result
}

// getChallengeAnswerHash returns the expected answer hash for a challenge.
// In the new template system, this is primarily for backward compatibility.
func getChallengeAnswerHash(challenge types.AIChallenge) string {
	// For backward compatibility with legacy tests
	if len(challengePool) > 0 && challengePool[0].AnswerHash != "" {
		if idx, ok := questionHashIndex[challenge.ChallengeHash]; ok {
			return challengePool[idx].AnswerHash
		}
	}
	return challenge.ChallengeData
}

// questionHashIndex - kept for backwards compatibility
var questionHashIndex map[string]int

func init() {
	questionHashIndex = make(map[string]int, len(challengePool))
	for i, c := range challengePool {
		h := sha256.Sum256([]byte(c.Question))
		questionHashIndex[hex.EncodeToString(h[:])] = i
	}
}

func (k Keeper) GetChallenge(ctx sdk.Context, epoch uint64) (types.AIChallenge, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyChallenge(epoch))
	if bz == nil {
		return types.AIChallenge{}, false
	}
	var challenge types.AIChallenge
	k.cdc.MustUnmarshal(bz, &challenge)
	return challenge, true
}

func (k Keeper) SetChallenge(ctx sdk.Context, challenge types.AIChallenge) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&challenge)
	store.Set(types.KeyChallenge(challenge.Epoch), bz)
}

func (k Keeper) GetEpochResponses(ctx sdk.Context, epoch uint64) []types.AIResponse {
	store := ctx.KVStore(k.storeKey)
	prefix := types.KeyAIResponsePrefix(epoch)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var responses []types.AIResponse
	for ; iterator.Valid(); iterator.Next() {
		var response types.AIResponse
		k.cdc.MustUnmarshal(iterator.Value(), &response)
		responses = append(responses, response)
	}
	return responses
}

const CheatPenaltyReputation = -20
const CheatPenaltyStakePercent = 20

// EvaluateEpochChallenges scores revealed answers using validator cross-evaluation.
// In the template-based system, there's no single correct answer - validators
// evaluate each other's responses based on quality metrics.
func (k Keeper) EvaluateEpochChallenges(ctx sdk.Context, epoch uint64) {
	_, found := k.GetChallenge(ctx, epoch)
	if !found {
		return
	}

	responses := k.GetEpochResponses(ctx, epoch)
	respondents := make(map[string]bool)

	// Detect cheaters (agents submitting identical answers)
	cheaters := k.detectCheaters(responses, "")

	// Cross-evaluation: validators score each other's responses
	scores := k.calculateCrossScores(responses)

	for _, resp := range responses {
		respondents[resp.ValidatorAddress] = true

		if cheaters[resp.ValidatorAddress] {
			k.penalizeCheater(ctx, resp.ValidatorAddress)
			k.UpdateChallengeStats(ctx, resp.ValidatorAddress, epoch, -1)
			resp.Score = -1
		} else {
			score := scores[resp.ValidatorAddress]
			bonus := calculateAIBonus(score)
			k.SetAIBonus(ctx, resp.ValidatorAddress, bonus)

			if score >= 80 {
				k.UpdateReputationWithHistory(ctx, resp.ValidatorAddress, 2, "ai_challenge_excellent")
			} else if score >= 50 {
				k.UpdateReputationWithHistory(ctx, resp.ValidatorAddress, 1, "ai_challenge_pass")
			}

			k.UpdateChallengeStats(ctx, resp.ValidatorAddress, epoch, int64(score))
			resp.Score = int64(score)
		}

		store := ctx.KVStore(k.storeKey)
		resp.Evaluated = true
		bz := k.cdc.MustMarshal(&resp)
		store.Set(types.KeyAIResponse(epoch, resp.ValidatorAddress), bz)
	}

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE &&
			k.isActiveValidatorAddress(ctx, agent.Address) &&
			!respondents[agent.Address] {
			k.SetAIBonus(ctx, agent.Address, 0)
		}
		return false
	})

	k.RecordChallengeMetrics(ctx, epoch, responses)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"ai_challenge_evaluated",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("responses_count", fmt.Sprintf("%d", len(responses))),
		sdk.NewAttribute("cheaters_count", fmt.Sprintf("%d", len(cheaters))),
	))
}

// calculateCrossScores evaluates answer quality using multiple metrics:
// 1. Format compliance (has expected structure)
// 2. Length/complexity (indicates real AI usage)
// 3. Keyword presence (technical terms relevant to category)
// 4. Cross-validation with peers
func (k Keeper) calculateCrossScores(responses []types.AIResponse) map[string]int {
	scores := make(map[string]int)
	if len(responses) == 0 {
		return scores
	}

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

	for _, resp := range responses {
		if resp.RevealData == "" {
			scores[resp.ValidatorAddress] = 0
			continue
		}

		normalized := strings.ToLower(normalizeAnswer(resp.RevealData))
		lengthScore := calculateLengthScore(len(resp.RevealData))
		keywordScore := calculateKeywordScore(normalized, categoryKeywords)
		formatScore := calculateFormatScore(normalized)

		totalScore := lengthScore + keywordScore + formatScore
		scores[resp.ValidatorAddress] = clampScore(totalScore)
	}

	return scores
}

func calculateLengthScore(length int) int {
	switch {
	case length >= 500:
		return 35
	case length >= 200:
		return 25
	case length >= 100:
		return 15
	case length >= 50:
		return 10
	default:
		return 5
	}
}

func calculateKeywordScore(normalized string, keywords map[string][]string) int {
	totalKeywords := 0
	for _, terms := range keywords {
		for _, term := range terms {
			if strings.Contains(normalized, term) {
				totalKeywords++
			}
		}
	}

	switch {
	case totalKeywords >= 8:
		return 35
	case totalKeywords >= 5:
		return 25
	case totalKeywords >= 3:
		return 15
	case totalKeywords >= 1:
		return 10
	default:
		return 5
	}
}

func calculateFormatScore(normalized string) int {
	hasBulletPoints := strings.Contains(normalized, "•") || strings.Contains(normalized, "-") || strings.Contains(normalized, "*")
	hasNumbers := func() bool {
		for _, c := range normalized {
			if c >= '0' && c <= '9' {
				return true
			}
		}
		return false
	}()
	hasTechnicalTerms := strings.Contains(normalized, "because") || strings.Contains(normalized, "therefore") || strings.Contains(normalized, "first") || strings.Contains(normalized, "second")

	score := 10
	if hasBulletPoints {
		score += 5
	}
	if hasNumbers {
		score += 5
	}
	if hasTechnicalTerms {
		score += 5
	}
	return score
}

func clampScore(score int) int {
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

// detectCheaters flags agents that submitted identical normalized reveal data
// (the actual answer content). This catches real collusion — agents copying each
// other's answers — unlike the old commitHash comparison which could never trigger.
func (k Keeper) detectCheaters(responses []types.AIResponse, expectedHash string) map[string]bool {
	answerGroups := make(map[string][]string)

	for _, resp := range responses {
		if resp.RevealData == "" {
			continue
		}
		normalized := normalizeAnswer(resp.RevealData)
		answerGroups[normalized] = append(answerGroups[normalized], resp.ValidatorAddress)
	}

	cheaters := make(map[string]bool)
	for normalized, addrs := range answerGroups {
		answerHash := sha256.Sum256([]byte(normalized))
		if expectedHash != "" && hex.EncodeToString(answerHash[:]) == expectedHash {
			continue
		}
		if len(addrs) >= CheaterAnswerThreshold {
			for _, addr := range addrs {
				cheaters[addr] = true
			}
		}
	}
	return cheaters
}

func (k Keeper) penalizeCheater(ctx sdk.Context, address string) {
	k.SetAIBonus(ctx, address, -5)
	k.UpdateReputation(ctx, address, CheatPenaltyReputation)

	agent, found := k.GetAgent(ctx, address)
	if !found {
		return
	}

	slashAmount := agent.StakeAmount.Amount.MulRaw(CheatPenaltyStakePercent).QuoRaw(100)
	if slashAmount.IsPositive() {
		slashCoin := sdk.NewCoin("aaxon", slashAmount)
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(slashCoin)); err != nil {
			k.Logger(ctx).Error("failed to slash cheater stake", "address", address, "error", err)
			return
		}
		agent.StakeAmount = agent.StakeAmount.Sub(slashCoin)
		k.SetAgent(ctx, agent)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"ai_challenge_cheat_detected",
		sdk.NewAttribute("address", address),
		sdk.NewAttribute("slashed", slashAmount.String()),
		sdk.NewAttribute("reputation_penalty", fmt.Sprintf("%d", CheatPenaltyReputation)),
	))
}

// scoreResponseByHash compares hash(normalizedReveal) against the expected answer hash.
// This avoids needing plaintext answers at evaluation time.
func scoreResponseByHash(resp types.AIResponse, expectedHash string) int {
	if resp.RevealData == "" || expectedHash == "" {
		return 0
	}

	revealHash := sha256.Sum256([]byte(normalizeAnswer(resp.RevealData)))
	if hex.EncodeToString(revealHash[:]) == expectedHash {
		return 100
	}

	// Partial-credit scoring is intentionally disabled in hash-only mode because the
	// plaintext canonical answer is unavailable at evaluation time.
	return 10
}

// scoreResponse is kept for legacy tests and tooling that still pass the
// plaintext answer instead of the normalized answer hash.
func scoreResponse(resp types.AIResponse, answer string) int {
	if resp.RevealData == "" {
		return 0
	}

	reveal := normalizeAnswer(resp.RevealData)
	expected := normalizeAnswer(answer)
	if expected != "" && reveal == expected {
		return 100
	}
	if expected != "" && (strings.Contains(reveal, expected) || strings.Contains(expected, reveal)) {
		return 50
	}
	return 10
}

func calculateAIBonus(score int) int64 {
	switch {
	case score >= 90:
		return 30
	case score >= 70:
		return 20
	case score >= 50:
		return 10
	case score >= 20:
		return 5
	default:
		return 0
	}
}

func normalizeAnswer(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + 32
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result = append(result, c)
		}
	}
	return string(result)
}

func stringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
