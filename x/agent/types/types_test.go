package types_test

import (
	"math"
	"testing"

	"github.com/axon-chain/axon/x/agent/types"
)

// ---------------------------------------------------------------------------
// Uint64ToBytes / BytesToUint64 round-trip
// ---------------------------------------------------------------------------

func TestUint64BytesRoundTrip(t *testing.T) {
	values := []uint64{
		0, 1, 2, 127, 128, 255, 256,
		1<<16 - 1, 1 << 16,
		1<<32 - 1, 1 << 32,
		math.MaxUint64 - 1, math.MaxUint64,
	}
	for _, v := range values {
		bz := types.Uint64ToBytes(v)
		if len(bz) != 8 {
			t.Errorf("Uint64ToBytes(%d) returned %d bytes, want 8", v, len(bz))
		}
		got := types.BytesToUint64(bz)
		if got != v {
			t.Errorf("round-trip failed: input=%d, got=%d", v, got)
		}
	}
}

func TestUint64ToBytesBigEndian(t *testing.T) {
	bz := types.Uint64ToBytes(1)
	if bz[7] != 1 {
		t.Errorf("expected big-endian encoding, last byte should be 1, got %d", bz[7])
	}
	for i := 0; i < 7; i++ {
		if bz[i] != 0 {
			t.Errorf("expected leading zeros in big-endian encoding, byte[%d] = %d", i, bz[i])
		}
	}
}

func TestUint64ToBytesZero(t *testing.T) {
	bz := types.Uint64ToBytes(0)
	for i, b := range bz {
		if b != 0 {
			t.Errorf("Uint64ToBytes(0) byte[%d] = %d, want 0", i, b)
		}
	}
}

func TestUint64BytesOrdering(t *testing.T) {
	a := types.Uint64ToBytes(100)
	b := types.Uint64ToBytes(200)
	if string(a) >= string(b) {
		t.Error("big-endian encoding should preserve ordering: 100 < 200")
	}
}

// ---------------------------------------------------------------------------
// Key functions
// ---------------------------------------------------------------------------

func TestKeyAgent(t *testing.T) {
	key := types.KeyAgent("axon1test")
	want := types.AgentKeyPrefix + "axon1test"
	if string(key) != want {
		t.Errorf("KeyAgent = %q, want %q", key, want)
	}
}

func TestKeyAgentDifferentAddresses(t *testing.T) {
	k1 := types.KeyAgent("axon1aaa")
	k2 := types.KeyAgent("axon1bbb")
	if string(k1) == string(k2) {
		t.Error("different addresses should produce different keys")
	}
}

func TestKeyDeregisterQueue(t *testing.T) {
	key := types.KeyDeregisterQueue("axon1test")
	want := types.DeregisterQueueKeyPrefix + "axon1test"
	if string(key) != want {
		t.Errorf("KeyDeregisterQueue = %q, want %q", key, want)
	}
}

func TestKeyChallenge(t *testing.T) {
	key := types.KeyChallenge(42)
	if len(key) != len(types.ChallengeKeyPrefix)+8 {
		t.Errorf("KeyChallenge length = %d, want %d", len(key), len(types.ChallengeKeyPrefix)+8)
	}
	prefix := string(key[:len(types.ChallengeKeyPrefix)])
	if prefix != types.ChallengeKeyPrefix {
		t.Errorf("KeyChallenge prefix = %q, want %q", prefix, types.ChallengeKeyPrefix)
	}
}

func TestKeyChallengeDifferentEpochs(t *testing.T) {
	k1 := types.KeyChallenge(1)
	k2 := types.KeyChallenge(2)
	if string(k1) == string(k2) {
		t.Error("different epochs should produce different challenge keys")
	}
}

func TestKeyChallengePool(t *testing.T) {
	key := types.KeyChallengePool(10)
	if len(key) != len(types.ChallengePoolKeyPrefix)+8 {
		t.Errorf("KeyChallengePool length = %d, want %d", len(key), len(types.ChallengePoolKeyPrefix)+8)
	}
}

func TestKeyAIResponse(t *testing.T) {
	key := types.KeyAIResponse(5, "axon1val")
	if len(key) == 0 {
		t.Fatal("KeyAIResponse should not be empty")
	}
	prefix := string(key[:len(types.AIResponseKeyPrefix)])
	if prefix != types.AIResponseKeyPrefix {
		t.Errorf("KeyAIResponse prefix = %q, want %q", prefix, types.AIResponseKeyPrefix)
	}
}

func TestKeyAIResponseDifferentValidators(t *testing.T) {
	k1 := types.KeyAIResponse(1, "axon1aaa")
	k2 := types.KeyAIResponse(1, "axon1bbb")
	if string(k1) == string(k2) {
		t.Error("different validators should produce different AI response keys")
	}
}

func TestKeyAIResponseDifferentEpochs(t *testing.T) {
	k1 := types.KeyAIResponse(1, "axon1aaa")
	k2 := types.KeyAIResponse(2, "axon1aaa")
	if string(k1) == string(k2) {
		t.Error("different epochs should produce different AI response keys")
	}
}

func TestKeyAIResponsePrefix(t *testing.T) {
	prefix := types.KeyAIResponsePrefix(5)
	fullKey := types.KeyAIResponse(5, "axon1val")

	if len(prefix) > len(fullKey) {
		t.Fatal("prefix should be shorter than full key")
	}
	if string(fullKey[:len(prefix)]) != string(prefix) {
		t.Error("full key should start with its prefix")
	}
}

func TestKeyEpochActivity(t *testing.T) {
	key := types.KeyEpochActivity(3, "axon1test")
	if len(key) == 0 {
		t.Fatal("KeyEpochActivity should not be empty")
	}
	prefix := string(key[:len(types.EpochActivityKeyPrefix)])
	if prefix != types.EpochActivityKeyPrefix {
		t.Errorf("KeyEpochActivity prefix = %q, want %q", prefix, types.EpochActivityKeyPrefix)
	}
}

func TestKeyAIBonus(t *testing.T) {
	key := types.KeyAIBonus("axon1val")
	want := types.AIBonusKeyPrefix + "axon1val"
	if string(key) != want {
		t.Errorf("KeyAIBonus = %q, want %q", key, want)
	}
}

func TestKeyDeployCount(t *testing.T) {
	key := types.KeyDeployCount(1, "axon1val")
	if len(key) == 0 {
		t.Fatal("KeyDeployCount should not be empty")
	}
	prefix := string(key[:len(types.DeployCountKeyPrefix)])
	if prefix != types.DeployCountKeyPrefix {
		t.Errorf("KeyDeployCount prefix = %q, want %q", prefix, types.DeployCountKeyPrefix)
	}
}

func TestKeyContractCall(t *testing.T) {
	key := types.KeyContractCall(1, "axon1val")
	if len(key) == 0 {
		t.Fatal("KeyContractCall should not be empty")
	}
	prefix := string(key[:len(types.ContractCallKeyPrefix)])
	if prefix != types.ContractCallKeyPrefix {
		t.Errorf("KeyContractCall prefix = %q, want %q", prefix, types.ContractCallKeyPrefix)
	}
}

// ---------------------------------------------------------------------------
// Key prefix uniqueness — no prefix is a prefix of another
// ---------------------------------------------------------------------------

func TestKeyPrefixesUnique(t *testing.T) {
	prefixes := []struct {
		name  string
		value string
	}{
		{"AgentKeyPrefix", types.AgentKeyPrefix},
		{"DeregisterQueueKeyPrefix", types.DeregisterQueueKeyPrefix},
		{"ChallengeKeyPrefix", types.ChallengeKeyPrefix},
		{"ChallengePoolKeyPrefix", types.ChallengePoolKeyPrefix},
		{"AIResponseKeyPrefix", types.AIResponseKeyPrefix},
		{"ContributionKeyPrefix", types.ContributionKeyPrefix},
		{"EpochActivityKeyPrefix", types.EpochActivityKeyPrefix},
		{"AIBonusKeyPrefix", types.AIBonusKeyPrefix},
		{"DeployCountKeyPrefix", types.DeployCountKeyPrefix},
		{"ContractCallKeyPrefix", types.ContractCallKeyPrefix},
	}

	for i := 0; i < len(prefixes); i++ {
		for j := i + 1; j < len(prefixes); j++ {
			if prefixes[i].value == prefixes[j].value {
				t.Errorf("duplicate prefix: %s and %s both = %q",
					prefixes[i].name, prefixes[j].name, prefixes[i].value)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// DefaultParams
// ---------------------------------------------------------------------------

func TestDefaultParamsValues(t *testing.T) {
	p := types.DefaultParams()

	if p.MinRegisterStake != 100 {
		t.Errorf("MinRegisterStake = %d, want 100", p.MinRegisterStake)
	}
	if p.RegisterBurnAmount != 20 {
		t.Errorf("RegisterBurnAmount = %d, want 20", p.RegisterBurnAmount)
	}
	if p.ContractDeployBurn != 10 {
		t.Errorf("ContractDeployBurn = %d, want 10", p.ContractDeployBurn)
	}
	if p.InitialReputation != 10 {
		t.Errorf("InitialReputation = %d, want 10", p.InitialReputation)
	}
	if p.MaxReputation != 100 {
		t.Errorf("MaxReputation = %d, want 100", p.MaxReputation)
	}
	if p.HeartbeatInterval != 100 {
		t.Errorf("HeartbeatInterval = %d, want 100", p.HeartbeatInterval)
	}
	if p.HeartbeatTimeout != 720 {
		t.Errorf("HeartbeatTimeout = %d, want 720", p.HeartbeatTimeout)
	}
	if p.EpochLength != 720 {
		t.Errorf("EpochLength = %d, want 720", p.EpochLength)
	}
	if p.AiChallengeWindow != 50 {
		t.Errorf("AiChallengeWindow = %d, want 50", p.AiChallengeWindow)
	}
}

func TestDefaultParamsValid(t *testing.T) {
	p := types.DefaultParams()
	if err := p.Validate(); err != nil {
		t.Errorf("default params should be valid: %v", err)
	}
}

func TestParamsZeroStakeInvalid(t *testing.T) {
	p := types.DefaultParams()
	p.MinRegisterStake = 0
	if err := p.Validate(); err == nil {
		t.Error("params with zero MinRegisterStake should be invalid")
	}
}

func TestParamsMinStakeOne(t *testing.T) {
	p := types.DefaultParams()
	p.MinRegisterStake = 1
	p.RegisterBurnAmount = 0
	if err := p.Validate(); err != nil {
		t.Errorf("params with MinRegisterStake=1, RegisterBurnAmount=0 should be valid: %v", err)
	}
}

func TestParamsBurnExceedsStake(t *testing.T) {
	p := types.DefaultParams()
	p.MinRegisterStake = 10
	p.RegisterBurnAmount = 20
	if err := p.Validate(); err == nil {
		t.Error("params with RegisterBurnAmount > MinRegisterStake should be invalid")
	}
}

func TestParamsZeroEpochLength(t *testing.T) {
	p := types.DefaultParams()
	p.EpochLength = 0
	if err := p.Validate(); err == nil {
		t.Error("params with EpochLength=0 should be invalid")
	}
}

func TestParamsHeartbeatTimeoutLessThanInterval(t *testing.T) {
	p := types.DefaultParams()
	p.HeartbeatTimeout = 50
	p.HeartbeatInterval = 100
	if err := p.Validate(); err == nil {
		t.Error("params with HeartbeatTimeout <= HeartbeatInterval should be invalid")
	}
}

// ---------------------------------------------------------------------------
// Genesis
// ---------------------------------------------------------------------------

func TestDefaultGenesisNotNil(t *testing.T) {
	gs := types.DefaultGenesis()
	if gs == nil {
		t.Fatal("DefaultGenesis should not be nil")
	}
}

func TestDefaultGenesisEmptyAgents(t *testing.T) {
	gs := types.DefaultGenesis()
	if len(gs.Agents) != 0 {
		t.Errorf("default genesis should have 0 agents, got %d", len(gs.Agents))
	}
}

func TestDefaultGenesisValid(t *testing.T) {
	gs := types.DefaultGenesis()
	if err := gs.Validate(); err != nil {
		t.Errorf("default genesis should be valid: %v", err)
	}
}

func TestGenesisWithInvalidParams(t *testing.T) {
	gs := types.DefaultGenesis()
	gs.Params.MinRegisterStake = 0
	if err := gs.Validate(); err == nil {
		t.Error("genesis with invalid params should fail validation")
	}
}

func TestGenesisParamsMatchDefaults(t *testing.T) {
	gs := types.DefaultGenesis()
	dp := types.DefaultParams()
	if gs.Params.MinRegisterStake != dp.MinRegisterStake {
		t.Errorf("genesis MinRegisterStake = %d, want %d", gs.Params.MinRegisterStake, dp.MinRegisterStake)
	}
	if gs.Params.MaxReputation != dp.MaxReputation {
		t.Errorf("genesis MaxReputation = %d, want %d", gs.Params.MaxReputation, dp.MaxReputation)
	}
}

// ---------------------------------------------------------------------------
// Module constants
// ---------------------------------------------------------------------------

func TestModuleConstants(t *testing.T) {
	if types.ModuleName != "agent" {
		t.Errorf("ModuleName = %q, want %q", types.ModuleName, "agent")
	}
	if types.StoreKey != types.ModuleName {
		t.Errorf("StoreKey = %q, should equal ModuleName %q", types.StoreKey, types.ModuleName)
	}
	if types.RouterKey != types.ModuleName {
		t.Errorf("RouterKey = %q, should equal ModuleName %q", types.RouterKey, types.ModuleName)
	}
}

// ---------------------------------------------------------------------------
// Error sentinel values
// ---------------------------------------------------------------------------

func TestErrorsNotNil(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"ErrAgentAlreadyRegistered", types.ErrAgentAlreadyRegistered},
		{"ErrAgentNotFound", types.ErrAgentNotFound},
		{"ErrInsufficientStake", types.ErrInsufficientStake},
		{"ErrAgentSuspended", types.ErrAgentSuspended},
		{"ErrChallengeNotActive", types.ErrChallengeNotActive},
		{"ErrChallengeExpired", types.ErrChallengeExpired},
		{"ErrAlreadySubmitted", types.ErrAlreadySubmitted},
		{"ErrInvalidReveal", types.ErrInvalidReveal},
	}
	for _, e := range errors {
		if e.err == nil {
			t.Errorf("%s should not be nil", e.name)
		}
	}
}
