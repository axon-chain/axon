package keeper_test

import (
	"bytes"
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axon-chain/axon/x/agent/keeper"
	"github.com/axon-chain/axon/x/agent/types"
)

// mockBankKeeper implements types.BankKeeper for testing
type mockBankKeeper struct {
	balances map[string]sdk.Coins
	module   map[string]sdk.Coins
	burned   sdk.Coins
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		balances: make(map[string]sdk.Coins),
		module:   make(map[string]sdk.Coins),
	}
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	addr := senderAddr.String()
	m.balances[addr] = m.balances[addr].Sub(amt...)
	m.module[recipientModule] = m.module[recipientModule].Add(amt...)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	addr := recipientAddr.String()
	m.module[senderModule] = m.module[senderModule].Sub(amt...)
	m.balances[addr] = m.balances[addr].Add(amt...)
	return nil
}

func (m *mockBankKeeper) MintCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	m.module[moduleName] = m.module[moduleName].Add(amt...)
	return nil
}

func (m *mockBankKeeper) BurnCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	m.module[moduleName] = m.module[moduleName].Sub(amt...)
	m.burned = m.burned.Add(amt...)
	return nil
}

func (m *mockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.balances[addr.String()]
	return sdk.NewCoin(denom, coins.AmountOf(denom))
}

type mockStakingKeeper struct {
	validators map[string]stakingtypes.Validator
}

func newMockStakingKeeper() *mockStakingKeeper {
	return &mockStakingKeeper{validators: make(map[string]stakingtypes.Validator)}
}

func (m *mockStakingKeeper) GetValidator(_ context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error) {
	validator, ok := m.validators[addr.String()]
	if !ok {
		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
	}
	return validator, nil
}

func (m *mockStakingKeeper) GetValidatorByConsAddr(_ context.Context, _ sdk.ConsAddress) (stakingtypes.Validator, error) {
	return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
}

func (m *mockStakingKeeper) GetBondedValidatorsByPower(_ context.Context) ([]stakingtypes.Validator, error) {
	validators := make([]stakingtypes.Validator, 0, len(m.validators))
	for _, validator := range m.validators {
		if validator.IsBonded() && !validator.IsJailed() {
			validators = append(validators, validator)
		}
	}
	return validators, nil
}

func setupTestKeeper(t *testing.T) (keeper.Keeper, sdk.Context, *mockBankKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := runtime.NewKVStoreService(storeKey)
	_ = db

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	bankKeeper := newMockBankKeeper()

	k := keeper.NewKeeper(cdc, storeKey, bankKeeper, nil)

	// NOTE: In Cosmos SDK v0.54, setting up a proper test context with an
	// in-memory KV store is complex. For real testing, use the testutil framework.
	// This file provides the structure; actual execution requires a proper test harness.
	_ = k

	return k, sdk.Context{}, bankKeeper
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	if params.MinRegisterStake != 100 {
		t.Errorf("expected MinRegisterStake=100, got %d", params.MinRegisterStake)
	}
	if params.RegisterBurnAmount != 20 {
		t.Errorf("expected RegisterBurnAmount=20, got %d", params.RegisterBurnAmount)
	}
	if params.MaxReputation != 100 {
		t.Errorf("expected MaxReputation=100, got %d", params.MaxReputation)
	}
	if params.EpochLength != 720 {
		t.Errorf("expected EpochLength=720, got %d", params.EpochLength)
	}
	if params.HeartbeatTimeout != 720 {
		t.Errorf("expected HeartbeatTimeout=720, got %d", params.HeartbeatTimeout)
	}
}

func TestParamsValidation(t *testing.T) {
	p := types.DefaultParams()
	if err := p.Validate(); err != nil {
		t.Errorf("default params should be valid: %v", err)
	}

	p.MinRegisterStake = 0
	if err := p.Validate(); err == nil {
		t.Error("params with zero MinRegisterStake should be invalid")
	}
}

func TestDefaultGenesis(t *testing.T) {
	gs := types.DefaultGenesis()
	if gs == nil {
		t.Fatal("DefaultGenesis should not be nil")
	}
	if len(gs.Agents) != 0 {
		t.Errorf("expected 0 agents in default genesis, got %d", len(gs.Agents))
	}
	if err := gs.Validate(); err != nil {
		t.Errorf("default genesis should be valid: %v", err)
	}
}

func TestChallengePoolNotEmpty(t *testing.T) {
	// Verify challenge pool has entries for AI challenge generation
	if len(keeper.GetChallengePoolSize()) == 0 {
		t.Error("challenge pool should not be empty")
	}
}

func TestIsActiveValidatorAddress(t *testing.T) {
	stakingKeeper := newMockStakingKeeper()
	accAddr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
	valAddr := sdk.ValAddress(accAddr)
	stakingKeeper.validators[valAddr.String()] = stakingtypes.Validator{
		OperatorAddress: valAddr.String(),
		Status:          stakingtypes.Bonded,
	}

	k := keeper.NewKeeper(nil, nil, nil, stakingKeeper)
	if !keeper.IsActiveValidatorAddressForTest(k, sdk.Context{}, accAddr.String()) {
		t.Fatal("expected bonded validator account to be eligible")
	}

	stakingKeeper.validators[valAddr.String()] = stakingtypes.Validator{
		OperatorAddress: valAddr.String(),
		Status:          stakingtypes.Unbonded,
	}
	if keeper.IsActiveValidatorAddressForTest(k, sdk.Context{}, accAddr.String()) {
		t.Fatal("expected unbonded validator account to be ineligible")
	}
}

func TestNormalizeAnswer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"O(log n)", "o(logn)"},
		{"  CALL  ", "call"},
		{"insertion sort", "insertionsort"},
		{"less than 1/3", "lessthan1/3"},
	}

	for _, tt := range tests {
		result := keeper.NormalizeAnswerForTest(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeAnswer(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestScoreResponse(t *testing.T) {
	tests := []struct {
		reveal   string
		answer   string
		minScore int
		maxScore int
	}{
		{"O(log n)", "O(log n)", 100, 100},
		{"o(log n)", "O(log n)", 100, 100},
		{"CALL", "CALL", 100, 100},
		{"the answer is CALL", "CALL", 50, 50},
		{"completely wrong", "CALL", 10, 10},
		{"", "CALL", 0, 0},
	}

	for _, tt := range tests {
		score := keeper.ScoreResponseForTest(tt.reveal, tt.answer)
		if score < tt.minScore || score > tt.maxScore {
			t.Errorf("scoreResponse(reveal=%q, answer=%q) = %d, want [%d,%d]",
				tt.reveal, tt.answer, score, tt.minScore, tt.maxScore)
		}
	}
}

func TestCalculateAIBonus(t *testing.T) {
	tests := []struct {
		score int
		bonus int64
	}{
		{100, 30},
		{90, 30},
		{80, 20},
		{50, 10},
		{30, 5},
		{10, 0},
		{0, 0},
	}

	for _, tt := range tests {
		bonus := keeper.CalculateAIBonusForTest(tt.score)
		if bonus != tt.bonus {
			t.Errorf("calculateAIBonus(%d) = %d, want %d", tt.score, bonus, tt.bonus)
		}
	}
}

func TestKeyFunctions(t *testing.T) {
	addr := "axon1abc123"

	agentKey := types.KeyAgent(addr)
	if string(agentKey) != types.AgentKeyPrefix+addr {
		t.Errorf("KeyAgent mismatch: %s", agentKey)
	}

	deregKey := types.KeyDeregisterQueue(addr)
	if string(deregKey) != types.DeregisterQueueKeyPrefix+addr {
		t.Errorf("KeyDeregisterQueue mismatch: %s", deregKey)
	}

	epoch := uint64(42)
	challengeKey := types.KeyChallenge(epoch)
	if len(challengeKey) != len(types.ChallengeKeyPrefix)+8 {
		t.Errorf("KeyChallenge unexpected length: %d", len(challengeKey))
	}

	respKey := types.KeyAIResponse(epoch, addr)
	if len(respKey) == 0 {
		t.Error("KeyAIResponse should not be empty")
	}
}

func TestUint64ByteConversion(t *testing.T) {
	values := []uint64{0, 1, 255, 65535, 1<<32 - 1, 1<<64 - 1}
	for _, v := range values {
		bz := types.Uint64ToBytes(v)
		got := types.BytesToUint64(bz)
		if got != v {
			t.Errorf("Uint64 roundtrip failed: input=%d, got=%d", v, got)
		}
	}
}
