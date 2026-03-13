package wallet

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axon-chain/axon/x/agent/types"
)

// TrustLevel defines the trust tier for a target contract.
type TrustLevel uint8

const (
	TrustBlocked TrustLevel = 0 // blacklisted, all interactions forbidden
	TrustUnknown TrustLevel = 1 // default: strict limits + cooldown
	TrustLimited TrustLevel = 2 // owner-set per-channel limits, no cooldown
	TrustFull    TrustLevel = 3 // no restrictions
)

// WalletInfo is the on-chain state for a smart wallet.
type WalletInfo struct {
	Owner          common.Address // human controller, can set trust channels
	Operator       common.Address // agent key, can execute within policy
	Guardian       common.Address // emergency freeze / recovery
	TxLimit        *big.Int       // default per-tx cap for unknown targets
	DailyLimit     *big.Int       // default daily cap for unknown targets
	CooldownBlocks *big.Int       // reserved for future cooldown queue
	DailySpent     *big.Int       // aggregated daily spend on unknown targets
	LastResetBlock int64
	IsFrozen       bool
}

// TrustedChannel stores an owner-authorized trust policy for a specific target.
type TrustedChannel struct {
	Level        TrustLevel
	TxLimit      *big.Int // per-tx limit (TrustLimited only, ignored for Full)
	DailyLimit   *big.Int // daily limit (TrustLimited only, ignored for Full)
	AuthorizedAt int64    // block height of authorization
	ExpiresAt    int64    // 0 = never expires
}

// DailySpentEntry tracks per-target daily spend (for TrustLimited channels).
type DailySpentEntry struct {
	Amount     *big.Int
	ResetBlock int64
}

// --- Storage key helpers ---

const (
	walletPrefix = "AgentWallet/"
	trustInfix   = "/trust/"
	spentToInfix = "/spentTo/"
)

func walletKey(addr common.Address) []byte {
	return []byte(walletPrefix + addr.Hex())
}

func trustKey(wallet, contract common.Address) []byte {
	return []byte(walletPrefix + wallet.Hex() + trustInfix + contract.Hex())
}

func spentToKey(wallet, target common.Address) []byte {
	return []byte(walletPrefix + wallet.Hex() + spentToInfix + target.Hex())
}

func generateWalletAddress(caller common.Address, blockHeight int64) common.Address {
	data := append(caller.Bytes(), types.Uint64ToBytes(uint64(blockHeight))...)
	return common.BytesToAddress(crypto.Keccak256(data))
}

// --- Binary encoding / decoding ---

// WalletInfo layout (197 bytes):
//
//	owner(20) + operator(20) + guardian(20) + txLimit(32) + dailyLimit(32) +
//	cooldownBlocks(32) + dailySpent(32) + lastResetBlock(8) + isFrozen(1)
const walletInfoSize = 197
const legacyWalletInfoSize = 177

func encodeWallet(w WalletInfo) []byte {
	buf := make([]byte, 0, walletInfoSize)
	buf = append(buf, w.Owner.Bytes()...)
	buf = append(buf, w.Operator.Bytes()...)
	buf = append(buf, w.Guardian.Bytes()...)
	buf = append(buf, common.LeftPadBytes(safeBigBytes(w.TxLimit), 32)...)
	buf = append(buf, common.LeftPadBytes(safeBigBytes(w.DailyLimit), 32)...)
	buf = append(buf, common.LeftPadBytes(safeBigBytes(w.CooldownBlocks), 32)...)
	buf = append(buf, common.LeftPadBytes(safeBigBytes(w.DailySpent), 32)...)
	buf = append(buf, types.Uint64ToBytes(uint64(w.LastResetBlock))...)
	if w.IsFrozen {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	return buf
}

func decodeWallet(bz []byte) WalletInfo {
	if len(bz) >= walletInfoSize {
		return decodeWalletV2(bz)
	}
	if len(bz) >= legacyWalletInfoSize {
		return decodeLegacyWallet(bz)
	}
	return WalletInfo{}
}

func decodeWalletV2(bz []byte) WalletInfo {
	return WalletInfo{
		Owner:          common.BytesToAddress(bz[0:20]),
		Operator:       common.BytesToAddress(bz[20:40]),
		Guardian:       common.BytesToAddress(bz[40:60]),
		TxLimit:        new(big.Int).SetBytes(bz[60:92]),
		DailyLimit:     new(big.Int).SetBytes(bz[92:124]),
		CooldownBlocks: new(big.Int).SetBytes(bz[124:156]),
		DailySpent:     new(big.Int).SetBytes(bz[156:188]),
		LastResetBlock: int64(types.BytesToUint64(bz[188:196])),
		IsFrozen:       bz[196] == 1,
	}
}

// Legacy format (177 bytes, no Owner field — Owner defaults to Operator).
func decodeLegacyWallet(bz []byte) WalletInfo {
	op := common.BytesToAddress(bz[0:20])
	return WalletInfo{
		Owner:          op,
		Operator:       op,
		Guardian:       common.BytesToAddress(bz[20:40]),
		TxLimit:        new(big.Int).SetBytes(bz[40:72]),
		DailyLimit:     new(big.Int).SetBytes(bz[72:104]),
		CooldownBlocks: new(big.Int).SetBytes(bz[104:136]),
		DailySpent:     new(big.Int).SetBytes(bz[136:168]),
		LastResetBlock: int64(types.BytesToUint64(bz[168:176])),
		IsFrozen:       bz[176] == 1,
	}
}

// TrustedChannel layout (81 bytes):
//
//	level(1) + txLimit(32) + dailyLimit(32) + authorizedAt(8) + expiresAt(8)
const trustChannelSize = 81

func encodeTrustChannel(tc TrustedChannel) []byte {
	buf := make([]byte, 0, trustChannelSize)
	buf = append(buf, byte(tc.Level))
	buf = append(buf, common.LeftPadBytes(safeBigBytes(tc.TxLimit), 32)...)
	buf = append(buf, common.LeftPadBytes(safeBigBytes(tc.DailyLimit), 32)...)
	buf = append(buf, types.Uint64ToBytes(uint64(tc.AuthorizedAt))...)
	buf = append(buf, types.Uint64ToBytes(uint64(tc.ExpiresAt))...)
	return buf
}

func decodeTrustChannel(bz []byte) TrustedChannel {
	if len(bz) < trustChannelSize {
		return TrustedChannel{Level: TrustUnknown}
	}
	return TrustedChannel{
		Level:        TrustLevel(bz[0]),
		TxLimit:      new(big.Int).SetBytes(bz[1:33]),
		DailyLimit:   new(big.Int).SetBytes(bz[33:65]),
		AuthorizedAt: int64(types.BytesToUint64(bz[65:73])),
		ExpiresAt:    int64(types.BytesToUint64(bz[73:81])),
	}
}

// DailySpentEntry layout (40 bytes):
//
//	amount(32) + resetBlock(8)
const dailySpentEntrySize = 40

func encodeDailySpent(e DailySpentEntry) []byte {
	buf := make([]byte, 0, dailySpentEntrySize)
	buf = append(buf, common.LeftPadBytes(safeBigBytes(e.Amount), 32)...)
	buf = append(buf, types.Uint64ToBytes(uint64(e.ResetBlock))...)
	return buf
}

func decodeDailySpent(bz []byte) DailySpentEntry {
	if len(bz) < dailySpentEntrySize {
		return DailySpentEntry{Amount: big.NewInt(0)}
	}
	return DailySpentEntry{
		Amount:     new(big.Int).SetBytes(bz[0:32]),
		ResetBlock: int64(types.BytesToUint64(bz[32:40])),
	}
}

func safeBigBytes(v *big.Int) []byte {
	if v == nil {
		return []byte{0}
	}
	return v.Bytes()
}
