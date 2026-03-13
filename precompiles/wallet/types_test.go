package wallet

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestWalletEncodeDecodeRoundTrip(t *testing.T) {
	original := WalletInfo{
		Owner:          common.HexToAddress("0x1111111111111111111111111111111111111111"),
		Operator:       common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Guardian:       common.HexToAddress("0x3333333333333333333333333333333333333333"),
		TxLimit:        big.NewInt(1e18),
		DailyLimit:     new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		CooldownBlocks: big.NewInt(100),
		DailySpent:     big.NewInt(5e17),
		LastResetBlock: 42000,
		IsFrozen:       false,
	}

	bz := encodeWallet(original)
	if len(bz) != walletInfoSize {
		t.Fatalf("expected %d bytes, got %d", walletInfoSize, len(bz))
	}

	decoded := decodeWallet(bz)
	assertAddressEq(t, "Owner", original.Owner, decoded.Owner)
	assertAddressEq(t, "Operator", original.Operator, decoded.Operator)
	assertAddressEq(t, "Guardian", original.Guardian, decoded.Guardian)
	assertBigEq(t, "TxLimit", original.TxLimit, decoded.TxLimit)
	assertBigEq(t, "DailyLimit", original.DailyLimit, decoded.DailyLimit)
	assertBigEq(t, "CooldownBlocks", original.CooldownBlocks, decoded.CooldownBlocks)
	assertBigEq(t, "DailySpent", original.DailySpent, decoded.DailySpent)
	if decoded.LastResetBlock != original.LastResetBlock {
		t.Errorf("LastResetBlock: want %d, got %d", original.LastResetBlock, decoded.LastResetBlock)
	}
	if decoded.IsFrozen != original.IsFrozen {
		t.Errorf("IsFrozen: want %v, got %v", original.IsFrozen, decoded.IsFrozen)
	}
}

func TestWalletFrozenFlag(t *testing.T) {
	w := WalletInfo{
		Owner:          common.HexToAddress("0xaaaa"),
		Operator:       common.HexToAddress("0xbbbb"),
		Guardian:       common.HexToAddress("0xcccc"),
		TxLimit:        big.NewInt(0),
		DailyLimit:     big.NewInt(0),
		CooldownBlocks: big.NewInt(0),
		DailySpent:     big.NewInt(0),
		LastResetBlock: 0,
		IsFrozen:       true,
	}
	decoded := decodeWallet(encodeWallet(w))
	if !decoded.IsFrozen {
		t.Error("expected IsFrozen = true")
	}
}

func TestLegacyWalletDecode(t *testing.T) {
	op := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	guardian := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	txLimit := big.NewInt(999)
	dailyLimit := big.NewInt(9999)
	cooldown := big.NewInt(10)
	spent := big.NewInt(100)
	resetBlock := int64(500)

	// Build a 177-byte legacy blob (no Owner field)
	buf := make([]byte, 0, legacyWalletInfoSize)
	buf = append(buf, op.Bytes()...)
	buf = append(buf, guardian.Bytes()...)
	buf = append(buf, common.LeftPadBytes(txLimit.Bytes(), 32)...)
	buf = append(buf, common.LeftPadBytes(dailyLimit.Bytes(), 32)...)
	buf = append(buf, common.LeftPadBytes(cooldown.Bytes(), 32)...)
	buf = append(buf, common.LeftPadBytes(spent.Bytes(), 32)...)
	rb := make([]byte, 8)
	rb[0] = byte(resetBlock >> 56)
	rb[1] = byte(resetBlock >> 48)
	rb[2] = byte(resetBlock >> 40)
	rb[3] = byte(resetBlock >> 32)
	rb[4] = byte(resetBlock >> 24)
	rb[5] = byte(resetBlock >> 16)
	rb[6] = byte(resetBlock >> 8)
	rb[7] = byte(resetBlock)
	buf = append(buf, rb...)
	buf = append(buf, 0) // not frozen

	decoded := decodeWallet(buf)
	assertAddressEq(t, "Owner (legacy)", op, decoded.Owner)
	assertAddressEq(t, "Operator (legacy)", op, decoded.Operator)
	assertAddressEq(t, "Guardian (legacy)", guardian, decoded.Guardian)
	assertBigEq(t, "TxLimit (legacy)", txLimit, decoded.TxLimit)
}

func TestTrustChannelEncodeDecodeRoundTrip(t *testing.T) {
	original := TrustedChannel{
		Level:        TrustLimited,
		TxLimit:      big.NewInt(5e18),
		DailyLimit:   new(big.Int).Mul(big.NewInt(50), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		AuthorizedAt: 12345,
		ExpiresAt:    99999,
	}

	bz := encodeTrustChannel(original)
	if len(bz) != trustChannelSize {
		t.Fatalf("expected %d bytes, got %d", trustChannelSize, len(bz))
	}

	decoded := decodeTrustChannel(bz)
	if decoded.Level != original.Level {
		t.Errorf("Level: want %d, got %d", original.Level, decoded.Level)
	}
	assertBigEq(t, "TxLimit", original.TxLimit, decoded.TxLimit)
	assertBigEq(t, "DailyLimit", original.DailyLimit, decoded.DailyLimit)
	if decoded.AuthorizedAt != original.AuthorizedAt {
		t.Errorf("AuthorizedAt: want %d, got %d", original.AuthorizedAt, decoded.AuthorizedAt)
	}
	if decoded.ExpiresAt != original.ExpiresAt {
		t.Errorf("ExpiresAt: want %d, got %d", original.ExpiresAt, decoded.ExpiresAt)
	}
}

func TestTrustChannelFullLevel(t *testing.T) {
	tc := TrustedChannel{
		Level:        TrustFull,
		TxLimit:      big.NewInt(0),
		DailyLimit:   big.NewInt(0),
		AuthorizedAt: 100,
		ExpiresAt:    0,
	}
	decoded := decodeTrustChannel(encodeTrustChannel(tc))
	if decoded.Level != TrustFull {
		t.Errorf("expected TrustFull(3), got %d", decoded.Level)
	}
	if decoded.ExpiresAt != 0 {
		t.Errorf("expected ExpiresAt=0, got %d", decoded.ExpiresAt)
	}
}

func TestTrustChannelBlockedLevel(t *testing.T) {
	tc := TrustedChannel{
		Level:        TrustBlocked,
		TxLimit:      big.NewInt(0),
		DailyLimit:   big.NewInt(0),
		AuthorizedAt: 200,
		ExpiresAt:    300,
	}
	decoded := decodeTrustChannel(encodeTrustChannel(tc))
	if decoded.Level != TrustBlocked {
		t.Errorf("expected TrustBlocked(0), got %d", decoded.Level)
	}
}

func TestDailySpentEncodeDecodeRoundTrip(t *testing.T) {
	original := DailySpentEntry{
		Amount:     big.NewInt(123456789),
		ResetBlock: 77777,
	}

	bz := encodeDailySpent(original)
	if len(bz) != dailySpentEntrySize {
		t.Fatalf("expected %d bytes, got %d", dailySpentEntrySize, len(bz))
	}

	decoded := decodeDailySpent(bz)
	assertBigEq(t, "Amount", original.Amount, decoded.Amount)
	if decoded.ResetBlock != original.ResetBlock {
		t.Errorf("ResetBlock: want %d, got %d", original.ResetBlock, decoded.ResetBlock)
	}
}

func TestDailySpentDecodeNil(t *testing.T) {
	decoded := decodeDailySpent(nil)
	if decoded.Amount.Sign() != 0 {
		t.Errorf("expected zero amount for nil input, got %s", decoded.Amount)
	}
}

func TestStorageKeys(t *testing.T) {
	w := common.HexToAddress("0x1234")
	c := common.HexToAddress("0x5678")

	wk := walletKey(w)
	if string(wk[:len(walletPrefix)]) != walletPrefix {
		t.Errorf("walletKey missing prefix")
	}

	tk := trustKey(w, c)
	if len(tk) == 0 {
		t.Error("trustKey returned empty")
	}

	sk := spentToKey(w, c)
	if len(sk) == 0 {
		t.Error("spentToKey returned empty")
	}
}

func TestSafeBigBytesNil(t *testing.T) {
	result := safeBigBytes(nil)
	if len(result) != 1 || result[0] != 0 {
		t.Errorf("safeBigBytes(nil) should return [0], got %v", result)
	}
}

func TestGenerateWalletAddress(t *testing.T) {
	caller := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	addr1 := generateWalletAddress(caller, 100)
	addr2 := generateWalletAddress(caller, 101)
	if addr1 == addr2 {
		t.Error("different block heights should produce different wallet addresses")
	}
}

// helpers

func assertAddressEq(t *testing.T, name string, want, got common.Address) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %s, got %s", name, want.Hex(), got.Hex())
	}
}

func assertBigEq(t *testing.T, name string, want, got *big.Int) {
	t.Helper()
	if want.Cmp(got) != 0 {
		t.Errorf("%s: want %s, got %s", name, want, got)
	}
}
