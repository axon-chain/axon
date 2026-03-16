package types

const (
	ModuleName = "privacy"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	// Commitment Tree
	CommitmentKeyPrefix  = 0x20
	CommitmentRootKey    = 0x21
	CommitmentSizeKey    = 0x22
	HistoricalRootPrefix = 0x23

	// Nullifier Set
	NullifierKeyPrefix = 0x24

	// Shielded Pool
	ShieldedBalanceKey = 0x25

	// Identity Commitments
	IdentityCommitmentPrefix = 0x26

	// Agent Identity index (agent address → registered flag)
	AgentIdentityPrefix = 0x27

	// Verifying Keys
	VerifyingKeyPrefix = 0x28

	// Encrypted Memos
	EncryptedMemoPrefix = 0x29

	// Historical root counter (for FIFO eviction)
	HistoricalRootCounterKey  = 0x2A
	HistoricalRootIndexPrefix = 0x2B
)

func CommitmentKey(index uint64) []byte {
	key := []byte{CommitmentKeyPrefix}
	key = append(key, Uint64ToBytes(index)...)
	return key
}

func HistoricalRootKey(rootHash []byte) []byte {
	key := []byte{HistoricalRootPrefix}
	key = append(key, rootHash...)
	return key
}

func NullifierKey(nullifier []byte) []byte {
	key := []byte{NullifierKeyPrefix}
	key = append(key, nullifier...)
	return key
}

func IdentityKey(commitment []byte) []byte {
	key := []byte{IdentityCommitmentPrefix}
	key = append(key, commitment...)
	return key
}

func AgentIdentityKey(agentAddr string) []byte {
	key := []byte{AgentIdentityPrefix}
	key = append(key, []byte(agentAddr)...)
	return key
}

func HistoricalRootIndexKey(counter uint64) []byte {
	key := []byte{HistoricalRootIndexPrefix}
	key = append(key, Uint64ToBytes(counter)...)
	return key
}

func EncryptedMemoKey(commitmentIndex uint64) []byte {
	key := []byte{EncryptedMemoPrefix}
	key = append(key, Uint64ToBytes(commitmentIndex)...)
	return key
}

func VerifyingKeyKey(keyId []byte) []byte {
	key := []byte{VerifyingKeyPrefix}
	key = append(key, keyId...)
	return key
}

func Uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
	return b
}
