package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

// challengePool stores questions with plaintext answers in source.
// On-chain state only stores the question hash (not the question itself),
// so the current epoch's challenge cannot be identified until after evaluation.
var challengePool = []struct {
	Question string
	Answer   string
	Category string
}{
	{"What is the time complexity of binary search?", "O(log n)", "algorithms"},
	{"In Ethereum, what opcode is used to transfer ETH to another address?", "CALL", "blockchain"},
	{"What consensus algorithm does CometBFT use?", "PBFT", "blockchain"},
	{"What is the derivative of x^3 with respect to x?", "3x^2", "math"},
	{"In Go, what keyword is used to launch a concurrent goroutine?", "go", "programming"},
	{"What data structure uses LIFO (Last In First Out)?", "stack", "data_structures"},
	{"What is the SHA-256 hash length in bits?", "256", "cryptography"},
	{"What layer of the OSI model does TCP operate at?", "transport", "networking"},
	{"In a Merkle tree, what is stored in leaf nodes?", "hash of data", "data_structures"},
	{"What EIP introduced EIP-1559 fee mechanism?", "EIP-1559", "blockchain"},
	{"What is the base case needed for in recursive functions?", "termination", "algorithms"},
	{"What type of encryption uses the same key for encrypt and decrypt?", "symmetric", "cryptography"},
	{"In SQL, what clause filters groups after aggregation?", "HAVING", "databases"},
	{"What is a smart contract's equivalent of a constructor in Solidity?", "constructor", "blockchain"},
	{"Name the sorting algorithm with best-case O(n) and worst-case O(n^2).", "insertion sort", "algorithms"},
	{"What HTTP method is idempotent and used to update resources?", "PUT", "networking"},
	{"In BFT consensus, what fraction of nodes can be faulty?", "less than 1/3", "blockchain"},
	{"What does CAP theorem state about distributed systems?", "cannot have all three: consistency, availability, partition tolerance", "distributed_systems"},
	{"What is the purpose of a nonce in blockchain transactions?", "prevent replay attacks", "blockchain"},
	{"What Cosmos SDK module handles token transfers?", "bank", "blockchain"},
	{"What is the space complexity of a hash table?", "O(n)", "data_structures"},
	{"Name the pattern where an object notifies dependents of state changes.", "observer", "design_patterns"},
	{"What is the maximum block gas limit set in Axon genesis?", "40000000", "axon"},
	{"In proof of stake, what prevents nothing-at-stake attacks?", "slashing", "blockchain"},
	{"What encoding does Cosmos SDK use for addresses?", "bech32", "blockchain"},
	{"What is the halting problem about?", "undecidability of program termination", "theory"},
	{"What protocol does gRPC use for transport?", "HTTP/2", "networking"},
	{"Name the principle: a class should have only one reason to change.", "single responsibility", "design_patterns"},
	{"What is the gas cost of SSTORE in Ethereum when setting a zero to non-zero value?", "20000", "blockchain"},
	{"What type of database is LevelDB?", "key-value", "databases"},
	{"What algorithm finds the shortest path in a weighted graph with non-negative edges?", "dijkstra", "algorithms"},
	{"What is the worst-case time complexity of quicksort?", "O(n^2)", "algorithms"},
	{"What search algorithm explores all neighbors at the current depth before moving deeper?", "BFS", "algorithms"},
	{"What algorithm finds the minimum spanning tree by greedily adding the cheapest edge that does not form a cycle?", "kruskal", "algorithms"},
	{"What is the time complexity of merge sort?", "O(n log n)", "algorithms"},
	{"What algorithmic technique solves problems by breaking them into overlapping subproblems?", "dynamic programming", "algorithms"},
	{"What sorting algorithm has O(n log n) worst case and is in-place?", "heapsort", "algorithms"},
	{"What Ethereum token standard defines non-fungible tokens?", "ERC-721", "blockchain"},
	{"What mechanism in Cosmos enables cross-chain communication?", "IBC", "blockchain"},
	{"What Ethereum token standard is used for fungible tokens?", "ERC-20", "blockchain"},
	{"What is the name of the Ethereum bytecode execution environment?", "EVM", "blockchain"},
	{"What type of node stores the full blockchain history?", "full node", "blockchain"},
	{"What mechanism allows token holders to vote on protocol changes?", "governance", "blockchain"},
	{"What elliptic curve does Bitcoin use for digital signatures?", "secp256k1", "cryptography"},
	{"What key exchange protocol lets two parties establish a shared secret over an insecure channel?", "diffie-hellman", "cryptography"},
	{"What does AES stand for?", "advanced encryption standard", "cryptography"},
	{"What is the block size of AES in bits?", "128", "cryptography"},
	{"What type of cryptographic scheme allows verification without revealing the underlying data?", "zero knowledge proof", "cryptography"},
	{"What algorithm is widely used for public-key cryptography based on integer factorization?", "RSA", "cryptography"},
	{"What protocol resolves domain names to IP addresses?", "DNS", "networking"},
	{"What port does HTTPS use by default?", "443", "networking"},
	{"What transport protocol is connectionless?", "UDP", "networking"},
	{"What protocol is used to securely access a remote shell?", "SSH", "networking"},
	{"What HTTP status code means resource not found?", "404", "networking"},
	{"What network device operates at layer 3 of the OSI model?", "router", "networking"},
	{"What SQL command removes a table and its schema entirely?", "DROP", "databases"},
	{"What SQL keyword removes duplicate rows from query results?", "DISTINCT", "databases"},
	{"What property ensures a database transaction is all-or-nothing?", "atomicity", "databases"},
	{"In SQL, what type of JOIN returns all rows from the left table?", "LEFT JOIN", "databases"},
	{"What SQL command is used to add new rows to a table?", "INSERT", "databases"},
	{"What type of database management system guarantees ACID properties?", "relational", "databases"},
	{"What design pattern ensures a class has only one instance?", "singleton", "design_patterns"},
	{"What design pattern provides a surrogate object to control access to another object?", "proxy", "design_patterns"},
	{"What design pattern lets you compose objects into tree structures?", "composite", "design_patterns"},
	{"What design pattern defines a family of algorithms and makes them interchangeable?", "strategy", "design_patterns"},
	{"What design pattern converts the interface of a class into another expected interface?", "adapter", "design_patterns"},
	{"What is the sum of interior angles of a triangle in degrees?", "180", "math"},
	{"What is the next Fibonacci number after 5, 8, 13?", "21", "math"},
	{"What is log base 2 of 1024?", "10", "math"},
	{"What is the square root of 144?", "12", "math"},
	{"What is the value of pi rounded to two decimal places?", "3.14", "math"},
	{"What is 2 raised to the power of 10?", "1024", "math"},
	{"In Python, what keyword is used to define a generator function?", "yield", "programming"},
	{"In Java, what keyword prevents a class from being subclassed?", "final", "programming"},
	{"What programming paradigm treats computation as evaluation of mathematical functions?", "functional", "programming"},
	{"In Python, what built-in function returns the length of a container?", "len", "programming"},
	{"What does API stand for?", "application programming interface", "programming"},
	{"In Rust, what system prevents data races at compile time?", "ownership", "programming"},
	{"What distributed consensus algorithm uses a leader and log replication?", "raft", "distributed_systems"},
	{"What technique splits a database across multiple machines by key range?", "sharding", "distributed_systems"},
	{"What type of clock assigns a counter to events for partial ordering?", "lamport clock", "distributed_systems"},
	{"What consistency model guarantees that a read returns the most recent write?", "linearizability", "distributed_systems"},
	{"What protocol ensures all nodes in a distributed system agree on a single value?", "consensus", "distributed_systems"},
	{"What complexity class contains problems solvable in polynomial time?", "P", "theory"},
	{"What complexity class contains problems verifiable in polynomial time?", "NP", "theory"},
	{"What information-theoretic quantity measures uncertainty in a random variable?", "entropy", "theory"},
	{"What is a problem called if no algorithm can decide it for all inputs?", "undecidable", "theory"},
	{"What type of automaton recognizes regular languages?", "finite automaton", "theory"},
	{"What is the smallest token denomination in Axon?", "aaxon", "axon"},
	{"What module in Axon handles AI agent registration?", "agent", "axon"},
	{"What SDK framework does Axon build upon?", "cosmos sdk", "axon"},
	{"What consensus engine does Axon use?", "cometbft", "axon"},
	{"What activation function outputs values between 0 and 1?", "sigmoid", "machine_learning"},
	{"What technique reduces overfitting by randomly disabling neurons during training?", "dropout", "machine_learning"},
	{"What type of neural network is primarily used for image recognition?", "CNN", "machine_learning"},
	{"What optimization algorithm iteratively updates parameters using the gradient of the loss?", "gradient descent", "machine_learning"},
	{"What unsupervised learning algorithm partitions data into k groups?", "k-means", "machine_learning"},
	{"What metric measures the area under the receiver operating characteristic curve?", "AUC", "machine_learning"},
	{"What scheduling algorithm gives each process equal time slices in rotation?", "round robin", "operating_systems"},
	{"What memory management technique divides memory into fixed-size pages?", "paging", "operating_systems"},
	{"What is the first process started by the Linux kernel?", "init", "operating_systems"},
	{"What system call creates a new process in Unix?", "fork", "operating_systems"},
	{"What condition occurs when two or more processes each wait for the other to release a resource?", "deadlock", "operating_systems"},
	{"What hardware component translates virtual addresses to physical addresses?", "MMU", "operating_systems"},
	{"What attack injects malicious SQL through user input?", "SQL injection", "security"},
	{"What security protocol replaced SSL for encrypted web communication?", "TLS", "security"},
	{"What type of attack floods a server with traffic to make it unavailable?", "DDoS", "security"},
	{"What attack tricks a user's browser into making an unwanted request to another site?", "CSRF", "security"},
	{"What attack intercepts communication between two parties without their knowledge?", "man in the middle", "security"},
	{"What security principle states users should have only the minimum permissions required?", "least privilege", "security"},
}

// questionHashIndex is a lookup table from question hash → pool index,
// built once at init to avoid repeated hashing during challenge evaluation.
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

// GenerateChallenge creates a deterministic challenge for the epoch.
// ChallengeData stores only the question hash (not plaintext) so the on-chain
// state does not reveal which question was selected until after evaluation.
func (k Keeper) GenerateChallenge(ctx sdk.Context, epoch uint64) types.AIChallenge {
	poolSize := uint64(len(challengePool))
	if poolSize == 0 {
		return types.AIChallenge{}
	}

	seed := sha256.Sum256(append(
		ctx.HeaderHash(),
		types.Uint64ToBytes(epoch)...,
	))
	index := types.BytesToUint64(seed[:8]) % poolSize
	selected := challengePool[index]

	questionHash := sha256.Sum256([]byte(selected.Question))
	questionHashHex := hex.EncodeToString(questionHash[:])
	params := k.GetParams(ctx)

	challenge := types.AIChallenge{
		Epoch:         epoch,
		ChallengeHash: questionHashHex,
		ChallengeType: selected.Category,
		ChallengeData: questionHashHex, // hash only — no plaintext question on-chain
		DeadlineBlock: ctx.BlockHeight() + params.AiChallengeWindow,
	}

	k.SetChallenge(ctx, challenge)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"ai_challenge_generated",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("category", selected.Category),
		sdk.NewAttribute("question_hash", questionHashHex),
		sdk.NewAttribute("deadline_block", fmt.Sprintf("%d", challenge.DeadlineBlock)),
	))

	return challenge
}

// getChallengeAnswer looks up the answer by matching ChallengeHash against
// the precomputed question hash index (no plaintext question comparison).
func getChallengeAnswer(challenge types.AIChallenge) string {
	if idx, ok := questionHashIndex[challenge.ChallengeHash]; ok {
		return challengePool[idx].Answer
	}
	return ""
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

// CheatPenaltyReputation is the reputation penalty for cheating.
const CheatPenaltyReputation = -20

// CheatPenaltyStakePercent is the percentage of stake slashed for cheating.
const CheatPenaltyStakePercent = 20

// EvaluateEpochChallenges scores responses and detects cheating (whitepaper §8.6 path 5).
// Cheating = multiple validators submit identical commit hashes OR identical
// normalized reveal data (collusion/copying — audit H-2 fix).
func (k Keeper) EvaluateEpochChallenges(ctx sdk.Context, epoch uint64) {
	challenge, found := k.GetChallenge(ctx, epoch)
	if !found {
		return
	}

	correctAnswer := getChallengeAnswer(challenge)
	responses := k.GetEpochResponses(ctx, epoch)
	respondents := make(map[string]bool)
	cheaters := k.detectCheaters(responses)

	for _, resp := range responses {
		respondents[resp.ValidatorAddress] = true

		if cheaters[resp.ValidatorAddress] {
			k.penalizeCheater(ctx, resp.ValidatorAddress)
			resp.Score = -1
		} else {
			score := scoreResponse(resp, correctAnswer)
			bonus := calculateAIBonus(score)
			k.SetAIBonus(ctx, resp.ValidatorAddress, bonus)

			if score >= 80 {
				k.UpdateReputation(ctx, resp.ValidatorAddress, 2)
			} else if score >= 50 {
				k.UpdateReputation(ctx, resp.ValidatorAddress, 1)
			}
			resp.Score = int64(score)
		}

		store := ctx.KVStore(k.storeKey)
		resp.Evaluated = true
		bz := k.cdc.MustMarshal(&resp)
		store.Set(types.KeyAIResponse(epoch, resp.ValidatorAddress), bz)
	}

	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE && !respondents[agent.Address] {
			k.SetAIBonus(ctx, agent.Address, 0)
		}
		return false
	})

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"ai_challenge_evaluated",
		sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("responses_count", fmt.Sprintf("%d", len(responses))),
		sdk.NewAttribute("cheaters_count", fmt.Sprintf("%d", len(cheaters))),
	))
}

// detectCheaters finds validators that share identical commit hashes (collusion).
// Since commit = SHA256(address + ":" + answer), honest validators with the same
// answer produce different hashes. Identical commits indicate key/answer sharing.
func (k Keeper) detectCheaters(responses []types.AIResponse) map[string]bool {
	commitCounts := make(map[string][]string)

	for _, resp := range responses {
		if resp.CommitHash != "" {
			commitCounts[resp.CommitHash] = append(commitCounts[resp.CommitHash], resp.ValidatorAddress)
		}
	}

	cheaters := make(map[string]bool)
	for _, addrs := range commitCounts {
		if len(addrs) > 1 {
			for _, addr := range addrs {
				cheaters[addr] = true
			}
		}
	}
	return cheaters
}

// penalizeCheater slashes 20% of stake, reputation -20, AIBonus = -5.
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

func scoreResponse(resp types.AIResponse, correctAnswer string) int {
	if resp.RevealData == "" {
		return 0
	}

	normalizedReveal := normalizeAnswer(resp.RevealData)
	normalizedAnswer := normalizeAnswer(correctAnswer)

	if normalizedReveal == normalizedAnswer {
		return 100
	}

	// Allow partial match if reveal length is reasonable relative to answer.
	// Prevents gaming by concatenating all possible answers into a single reveal.
	maxRevealLen := len(normalizedAnswer) * 3
	if maxRevealLen < 64 {
		maxRevealLen = 64
	}
	if len(normalizedReveal) >= 3 && len(normalizedAnswer) > 0 &&
		len(normalizedReveal) <= maxRevealLen {
		if stringContains(normalizedReveal, normalizedAnswer) || stringContains(normalizedAnswer, normalizedReveal) {
			return 50
		}
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
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
