package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

// CheaterAnswerThreshold is the minimum number of agents submitting the same
// normalized answer before they are flagged as colluding cheaters.
const CheaterAnswerThreshold = 3

// challengePool stores questions with hashed answers only.
// No plaintext answers exist in source — validators must know the answers
// independently. Answer hashes are SHA-256(normalizeAnswer(answer)).
// To update the pool, use: go run tools/gen-answer-hashes/main.go
var challengePool = []struct {
	Question   string
	AnswerHash string
	Category   string
}{
	{"What is the time complexity of binary search?", "ab7c4b4440c85e8ce97d1e0bbab416b20d206f1cae76878afbd34e33ccd98db6", "algorithms"},
	{"In Ethereum, what opcode is used to transfer ETH to another address?", "7edb360f06acaef2cc80dba16cf563f199d347db4443da04da0c8173e3f9e4ed", "blockchain"},
	{"What consensus algorithm does CometBFT use?", "33acc48717c272e9a576248d2af56f7f591650ebe147ecf953f2e15e970e3aed", "blockchain"},
	{"What is the derivative of x^3 with respect to x?", "3e1d20bb6e2e6309c7d73a22a75331dc4ba9e9926f6d2544d0bd57d8174e4d8b", "math"},
	{"In Go, what keyword is used to launch a concurrent goroutine?", "4cd0e21a9a0795a14ec9aa5f0e7d1abff0492565770e43eafdf1e3e8afed1f33", "programming"},
	{"What data structure uses LIFO (Last In First Out)?", "6ee08e6eb3bc6f45bc99fcd39fcc479286a1beb0c04d39f204c61762378075d6", "data_structures"},
	{"What is the SHA-256 hash length in bits?", "51e8ea280b44e16934d4d611901f3d3afc41789840acdff81942c2f65009cd52", "cryptography"},
	{"What layer of the OSI model does TCP operate at?", "6694ea8075001f6628da20f1afdafc74a763e2098bafc633b057534792db6aad", "networking"},
	{"In a Merkle tree, what is stored in leaf nodes?", "f2b2355832773f0168ee64d7fa9b167b39c835d01430d7d36ebcd8efc0ad3283", "data_structures"},
	{"What EIP introduced EIP-1559 fee mechanism?", "658d44d65e8b9a3b5fa5667ef22b33f69a40569a3f3b300c0e1e3ee9ca7e68fd", "blockchain"},
	{"What is the base case needed for in recursive functions?", "bf632b0c64c01081793bd5c50806cb407f678ae04d3df1eb64b8e5fdc1545b1f", "algorithms"},
	{"What type of encryption uses the same key for encrypt and decrypt?", "ac9fc833c96e7347f2b1dae832ae54795a310aaeb83e18c06ddf28d57fe33ff2", "cryptography"},
	{"In SQL, what clause filters groups after aggregation?", "12ac05147efdd0b67f567003c7d4e231b74ab82cc728dabc86093e8d89af2c29", "databases"},
	{"What is a smart contract's equivalent of a constructor in Solidity?", "e3c1703abf8a6b3df04dd30e184fe02bb6724f46830b544562ff083169371068", "blockchain"},
	{"Name the sorting algorithm with best-case O(n) and worst-case O(n^2).", "b719b1a6763b170d16d00e115f21b27177eb7072819c1a50531e7058fb9b452d", "algorithms"},
	{"What HTTP method is idempotent and used to update resources?", "373cb2c6d4fe2778441d4f0266505b699fa518d002e5793b87f9b48836de3f62", "networking"},
	{"In BFT consensus, what fraction of nodes can be faulty?", "792d23fa7dc0adfcc53a8be57db0eac06bf1f4e0a46ee1a55465035dcd83f7cc", "blockchain"},
	{"What does CAP theorem state about distributed systems?", "3829c9300cee83099f62df43694c0fc29eb2b1ed874bab38f207172fc3266081", "distributed_systems"},
	{"What is the purpose of a nonce in blockchain transactions?", "639f3b55787bd98953fbd36fe9ca56c562ceed0e4f1b6385cc410ec44f08bede", "blockchain"},
	{"What Cosmos SDK module handles token transfers?", "4381dc2ab14285160c808659aee005d51255add7264b318d07c7417292c7442c", "blockchain"},
	{"What is the space complexity of a hash table?", "3963250267e2da8912df02a9bcd55ab975aad3347d90f5d4ec1df3d7fd245be3", "data_structures"},
	{"Name the pattern where an object notifies dependents of state changes.", "e41f1d834710df8c7bc2befedb7b92339973a92288d8581b3696604c0d04c5f7", "design_patterns"},
	{"What is the maximum block gas limit set in Axon genesis?", "2ddb67b8a8c259ffaff61a5abdd38f5b5d6f1c6e2af4344c85b17b77af2451cc", "axon"},
	{"In proof of stake, what prevents nothing-at-stake attacks?", "31af10958c444c19b6bc71c82b07c30cd24661f75920f93975a31fab1455d050", "blockchain"},
	{"What encoding does Cosmos SDK use for addresses?", "bda2ebcbf0bd6bc4ee1c330a64a9ff95e839cc2d25c593a01e704996bc1e869c", "blockchain"},
	{"What is the halting problem about?", "25f255bfe8a08bcbaffd2208961b3196757743dfafcb2a6e8adc1fe2954e6c05", "theory"},
	{"What protocol does gRPC use for transport?", "46179e1c8e2c3f841a9da7599aabad957dfd14d96704a68f24280290e381261a", "networking"},
	{"Name the principle: a class should have only one reason to change.", "dac229411941d57b25ab9b3cf2762605e27d7079e58f96f2d630e2ecab7b83d3", "design_patterns"},
	{"What is the gas cost of SSTORE in Ethereum when setting a zero to non-zero value?", "876c9b16254e157d1eb645390dcfae6f29b9d3cd394e73a91de8ee5d0e67ee43", "blockchain"},
	{"What type of database is LevelDB?", "264b8327c2695fd0cfa367bf03be6f03a87ebbc3650702542b6f6ef2a7a1686f", "databases"},
	{"What algorithm finds the shortest path in a weighted graph with non-negative edges?", "10682236c9a24dcea388854f756a0997282cce0ca7dd3658bee6a1a600062a5c", "algorithms"},
	{"What is the worst-case time complexity of quicksort?", "5558f13adcff8bc461cd7e20840b1bfab5e90f4e558924dde505765b9c87925b", "algorithms"},
	{"What search algorithm explores all neighbors at the current depth before moving deeper?", "a1b086dd3d736b47fb2ab2a290cb27657fb9774655f268fc078f9fd28f7ccb00", "algorithms"},
	{"What algorithm finds the minimum spanning tree by greedily adding the cheapest edge that does not form a cycle?", "862079b2a455c3efcdbca2c62f623af9aabb54106b9bba9d71a7bd0601e25bc9", "algorithms"},
	{"What is the time complexity of merge sort?", "b5f4d76448877bea0926a76e334c3d6a04d3306e0b454901d0b4f6bcee9836ac", "algorithms"},
	{"What algorithmic technique solves problems by breaking them into overlapping subproblems?", "cb6cc1389a1fadedc61ff3d3d10b7644c18704f968835f09ae29101bbb0b1093", "algorithms"},
	{"What sorting algorithm has O(n log n) worst case and is in-place?", "ad9f3f17fe1f5f7adc438ceb0ea309757213122cc063d8652a697687b383da57", "algorithms"},
	{"What Ethereum token standard defines non-fungible tokens?", "39d40b1bb24173cd7ad81b104f8aadf2bca503d7a3db5db8b6ccd9146a74cfa5", "blockchain"},
	{"What mechanism in Cosmos enables cross-chain communication?", "3263c2fa090392fd141e13070edcdfe7617a1c3d91648dfd912a4b6a541a8613", "blockchain"},
	{"What Ethereum token standard is used for fungible tokens?", "845516ec99f3fcde192ba507bb82cc802ebdf343f5aa5a3f4bb314e21467aace", "blockchain"},
	{"What is the name of the Ethereum bytecode execution environment?", "603871c2ddd41c26ee77495e2e31e6de7f9957e0dea3b0f09abf8a5ee17a0d4a", "blockchain"},
	{"What type of node stores the full blockchain history?", "2ae51dff4dd6d2853fdc28abde809e96407973305803e9ad8736643db410975e", "blockchain"},
	{"What mechanism allows token holders to vote on protocol changes?", "bf9422a2a04aad2180fad0913f1cdf8a9d3780fd074817d2569050d0a8095d3a", "blockchain"},
	{"What elliptic curve does Bitcoin use for digital signatures?", "383b27532153f353fa4cc689239f7365dfe924ebcf67807eb6916307a4e2701e", "cryptography"},
	{"What key exchange protocol lets two parties establish a shared secret over an insecure channel?", "d35e8c6bf652bf67252d5625fa330a400252175a51e9610299c21471caf762cd", "cryptography"},
	{"What does AES stand for?", "51c80c18389613730fac44454abe8b6322a8b0eb4a64a4c26faacf36388a1d34", "cryptography"},
	{"What is the block size of AES in bits?", "2747b7c718564ba5f066f0523b03e17f6a496b06851333d2d59ab6d863225848", "cryptography"},
	{"What type of cryptographic scheme allows verification without revealing the underlying data?", "f382e21334df3237fa4be5f8c4d93a26d1ad3125817ceb770c9ec9e21ecd6b8b", "cryptography"},
	{"What algorithm is widely used for public-key cryptography based on integer factorization?", "f5f69168bba3cfa1e2a80dff839b48db36df36fa876c1cd9d7d508f3ab308744", "cryptography"},
	{"What protocol resolves domain names to IP addresses?", "dd75a9d6fb309c4399fe425cd5f90ff95eba135d6924fb91766ee5d3726b168a", "networking"},
	{"What port does HTTPS use by default?", "6d05621ab7cb7b4fb796ca2ffbe1a141e0d4319d3deb6a05322b9de85d69b923", "networking"},
	{"What transport protocol is connectionless?", "571e437548ffbac2cccfa26d7026aa7bd84186d79ca5ab7a5924d9026359b9e0", "networking"},
	{"What protocol is used to securely access a remote shell?", "7f5a55cf3f88be936fb9440249cb449f3067ccee4b525d0027dc9278a29c32c1", "networking"},
	{"What HTTP status code means resource not found?", "6b3c238ebcf1f3c07cf0e556faa82c6b8fe96840ff4b6b7e9962a2d855843a0b", "networking"},
	{"What network device operates at layer 3 of the OSI model?", "74c95604043427f0bee1d0e16bfa53afd537f736ad0073c4cc4e1ccb3a82b5dc", "networking"},
	{"What SQL command removes a table and its schema entirely?", "d90ee9ccf6bea1d2942a7b21319338198dec2a746f8a0d0771621f00da2e0864", "databases"},
	{"What SQL keyword removes duplicate rows from query results?", "54c9e847ed79c5c2d4a4b32373e1c3ba5cff331219091e6a3dab219caeae80b9", "databases"},
	{"What property ensures a database transaction is all-or-nothing?", "3931c975909268ae950c3d126f70efd9158ed6168241ed66713a5ff47d7a2d4d", "databases"},
	{"In SQL, what type of JOIN returns all rows from the left table?", "4b61f23363dc4b08dea0e81b9487b0f24adb6c8124aa41c58ec0d4c60c3d7699", "databases"},
	{"What SQL command is used to add new rows to a table?", "1e22560cee2c4b727c6a117792e04a6769efbe2395f8e2528c603a153a446477", "databases"},
	{"What type of database management system guarantees ACID properties?", "5449d70e4205b9bcec11e0093209f247e31439854b0633dc2235254e010e9a0f", "databases"},
	{"What design pattern ensures a class has only one instance?", "e234ee35f82c5dd3f2f20ef0835225e326805e56b5fe83ee5e56c0f899f4901d", "design_patterns"},
	{"What design pattern provides a surrogate object to control access to another object?", "1241936d4dd3aad68fe7bfbdfe854b935926bc678fc72377e15166078916227a", "design_patterns"},
	{"What design pattern lets you compose objects into tree structures?", "ad1e26066637d18f500318563e09c716fe3b0f6cf646ce32fb4c79f995aa15f5", "design_patterns"},
	{"What design pattern defines a family of algorithms and makes them interchangeable?", "73dff70e25ad51ca765a584eef93a1cd909527e2ef4860e0b12f0a7e7ae1979a", "design_patterns"},
	{"What design pattern converts the interface of a class into another expected interface?", "ae1eae1d76e5b7c865c4122ce366a08025842566d2d96c75cc13e6353a73db0d", "design_patterns"},
	{"What is the sum of interior angles of a triangle in degrees?", "7b69759630f869f2723875f873935fed29d2d12b10ef763c1c33b8e0004cb405", "math"},
	{"What is the next Fibonacci number after 5, 8, 13?", "6f4b6612125fb3a0daecd2799dfd6c9c299424fd920f9b308110a2c1fbd8f443", "math"},
	{"What is log base 2 of 1024?", "4a44dc15364204a80fe80e9039455cc1608281820fe2b24f1e5233ade6af1dd5", "math"},
	{"What is the square root of 144?", "6b51d431df5d7f141cbececcf79edf3dd861c3b4069f0b11661a3eefacbba918", "math"},
	{"What is the value of pi rounded to two decimal places?", "2efff1261c25d94dd6698ea1047f5c0a7107ca98b0a6c2427ee6614143500215", "math"},
	{"What is 2 raised to the power of 10?", "e39eef82f61b21e2e7f762fcc4307358f165757f2e77ec855d6992f7e0191932", "math"},
	{"In Python, what keyword is used to define a generator function?", "8bbcba42d48dcffe6ea6efa3f5656ec8b8e0d5c931091c17a12516eb5d193e41", "programming"},
	{"In Java, what keyword prevents a class from being subclassed?", "2443630b4620165c8b173e7265e17526fe2787ae594364dd6d839ad58f2fc007", "programming"},
	{"What programming paradigm treats computation as evaluation of mathematical functions?", "3b637864e75ab14ebba72b7f39d73be2ec309d571e807fc1170290353b214d78", "programming"},
	{"In Python, what built-in function returns the length of a container?", "71fa9faaa6f884aa11f4cea21477b2204a48a4fa7f05cecad00a1250eeeffb4c", "programming"},
	{"What does API stand for?", "147d60cea94fc104db8beb2505378be6b0de6dbeeae4d497bd7a2b29790078e5", "programming"},
	{"In Rust, what system prevents data races at compile time?", "01b8016fdce455c4343951d02b110bd9c02ce8456f43f8893c8b4ccbb1ca54aa", "programming"},
	{"What distributed consensus algorithm uses a leader and log replication?", "76a35072df72591a656e69cab6f6fa99aa386fd5ace35c9042851eb324ec16b5", "distributed_systems"},
	{"What technique splits a database across multiple machines by key range?", "a5756b8dad17b6db2d7f23d0e01d18f1e3a9da25b4d95acb4efe5bd4f85f7f80", "distributed_systems"},
	{"What type of clock assigns a counter to events for partial ordering?", "39261de510c553cbb0eba883fc63b817a3e37dd3cdd9bb64773ac01f080b4972", "distributed_systems"},
	{"What consistency model guarantees that a read returns the most recent write?", "d7ac9cbaf1cc9bcd4d222ccf4e4cc7adcc5381b254d733dc9e01e777ced4acd7", "distributed_systems"},
	{"What protocol ensures all nodes in a distributed system agree on a single value?", "c983c585ac3c40d920834f96200066352ff58e323da4dadae1d948fb27e63f82", "distributed_systems"},
	{"What complexity class contains problems solvable in polynomial time?", "148de9c5a7a44d19e56cd9ae1a554bf67847afb0c58f6e12fa29ac7ddfca9940", "theory"},
	{"What complexity class contains problems verifiable in polynomial time?", "0a2f133eb9f7ca028a20aa3fcd7f6cb8a05a0e89c166e69e2fbd480a00be447d", "theory"},
	{"What information-theoretic quantity measures uncertainty in a random variable?", "67671a2f53dd910a8b35840edb6a0a1e751ae5532178ca7f025b823eee317992", "theory"},
	{"What is a problem called if no algorithm can decide it for all inputs?", "91f30934bc09ccc82ca215955da88acfc1f3d82759438b69e480b7ed08d32e57", "theory"},
	{"What type of automaton recognizes regular languages?", "407358029486e5ff59f88f3035d0f0aa1ef212c97f35c38dbd547b53979e3b63", "theory"},
	{"What is the smallest token denomination in Axon?", "752eb36644326c095da984432728579d3693ca137a0e0118260176feba6183c5", "axon"},
	{"What module in Axon handles AI agent registration?", "d4f0bc5a29de06b510f9aa428f1eedba926012b591fef7a518e776a7c9bd1824", "axon"},
	{"What SDK framework does Axon build upon?", "897654ebaf871429f46f0c0d56df45f95f3b0ee8fc5e9a2dac473be6c7654d33", "axon"},
	{"What consensus engine does Axon use?", "adac33cc3e985b27bdce81e708fba8b9b7cbf0a4b7fe0c92c12f121e91556d0e", "axon"},
	{"What activation function outputs values between 0 and 1?", "fa15bd82f01453acc8c2ebacd9b43eb898411e65024c88d01a501177e2a42df5", "machine_learning"},
	{"What technique reduces overfitting by randomly disabling neurons during training?", "1b4c0a293bc93fb96e930a71ce8f599c0a97f0b6e5ff929ed674b9bad828739f", "machine_learning"},
	{"What type of neural network is primarily used for image recognition?", "4aa08612a8f6e9e429c9a46055665f59a034a792dcbbbfb850c69430439aa433", "machine_learning"},
	{"What optimization algorithm iteratively updates parameters using the gradient of the loss?", "36200518ca7c15ba442112374b7f54cc6928a8a163e1076be421fa977893fe8b", "machine_learning"},
	{"What unsupervised learning algorithm partitions data into k groups?", "8650c839f7c371e7ec86db3004414ed23eb49b97f3523b279e03a43383163d56", "machine_learning"},
	{"What metric measures the area under the receiver operating characteristic curve?", "dc3743da64c5b837fa5aba22256f26a75e4f6089a82d6ef2163d2eb249f7c510", "machine_learning"},
	{"What scheduling algorithm gives each process equal time slices in rotation?", "87c7e8c457a3f6e8d3aa072224512385ee59e5393f86de08ed4568a17531631c", "operating_systems"},
	{"What memory management technique divides memory into fixed-size pages?", "26ca59cbe63ebf7f690310861c5839ba1b9eefad7de66123534af716d47f8845", "operating_systems"},
	{"What is the first process started by the Linux kernel?", "bb54068aea85faa7e487530083366be9962390af822e4c71ef1aca7033c83e66", "operating_systems"},
	{"What system call creates a new process in Unix?", "8d0c7ac992af2f4913b47d425af0fe08dca35d157cbf27e40c203589b983d1fd", "operating_systems"},
	{"What condition occurs when two or more processes each wait for the other to release a resource?", "4f6c0efa1a0f9935d690952daeb37c167e7732678bd34eca7503833ca53893ea", "operating_systems"},
	{"What hardware component translates virtual addresses to physical addresses?", "a7a47d953b9f346794729b66cd4fc48823171c78d882a0f14b504b2ac63e35c7", "operating_systems"},
	{"What attack injects malicious SQL through user input?", "645b82b573fe64356e51765a9c44fe4eb496483fe8b6b722b77d3e8e222cdbb9", "security"},
	{"What security protocol replaced SSL for encrypted web communication?", "b7e651cbb43ba0ca3498759c8c3596c3a11a199004cd9e5a198d50d4585ec8c5", "security"},
	{"What type of attack floods a server with traffic to make it unavailable?", "deeb92f091caa8e2404885e30da06e8507eee571e81b062ef6723c4ec0b8ecf0", "security"},
	{"What attack tricks a user's browser into making an unwanted request to another site?", "7ce12ba8782a32f74357cefb81edb8c20ea4d755115ecb4063348b8cc9d41f34", "security"},
	{"What attack intercepts communication between two parties without their knowledge?", "cca6b60b9a61ab32ea452e67f407bfac350adaf49eabcd6aa37550712238d34a", "security"},
	{"What security principle states users should have only the minimum permissions required?", "196839c141461caa4701edf8b507c840ae507c6cca48818612e4218596e71d09", "security"},
}

// questionHashIndex maps question hash → pool index for fast lookup.
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
// ChallengeData stores only the question hash so the on-chain state does not
// reveal which question was selected until after evaluation.
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
		ChallengeData: questionHashHex,
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

// getChallengeAnswerHash returns the expected answer hash for a challenge.
func getChallengeAnswerHash(challenge types.AIChallenge) string {
	if idx, ok := questionHashIndex[challenge.ChallengeHash]; ok {
		return challengePool[idx].AnswerHash
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

const CheatPenaltyReputation = -20
const CheatPenaltyStakePercent = 20

// EvaluateEpochChallenges scores revealed answers using hash comparison and
// detects cheating via duplicate normalized reveal data (F2 fix).
func (k Keeper) EvaluateEpochChallenges(ctx sdk.Context, epoch uint64) {
	challenge, found := k.GetChallenge(ctx, epoch)
	if !found {
		return
	}

	expectedHash := getChallengeAnswerHash(challenge)
	responses := k.GetEpochResponses(ctx, epoch)
	respondents := make(map[string]bool)
	cheaterExcludeHash := expectedHash
	if !k.IsV110UpgradeActivated(ctx) {
		cheaterExcludeHash = ""
	}
	cheaters := k.detectCheaters(responses, cheaterExcludeHash)

	for _, resp := range responses {
		respondents[resp.ValidatorAddress] = true

		if cheaters[resp.ValidatorAddress] {
			k.penalizeCheater(ctx, resp.ValidatorAddress)
			resp.Score = -1
		} else {
			score := scoreResponseByHash(resp, expectedHash)
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
		if agent.Status == types.AgentStatus_AGENT_STATUS_ONLINE &&
			k.isActiveValidatorAddress(ctx, agent.Address) &&
			!respondents[agent.Address] {
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
