package keeper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/privacy/types"
)

// Viewing Key system (P6)
//
// Viewing keys allow selective disclosure: the holder can decrypt all
// shielded transaction memos belonging to a specific Agent without being
// able to spend funds (that requires the spending key).
//
// Chain stores: encrypted memo alongside each commitment.
// Off-chain: Agent locally holds (viewingKey, spendingKey).

func encryptedMemoKey(commitmentIndex uint64) []byte {
	return types.EncryptedMemoKey(commitmentIndex)
}

// StoreEncryptedMemo stores the encrypted memo for a commitment.
// Called during shield/privateTransfer after commitment insertion.
func (k Keeper) StoreEncryptedMemo(ctx sdk.Context, commitmentIndex uint64, encryptedMemo []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(encryptedMemoKey(commitmentIndex), encryptedMemo)
}

// GetEncryptedMemo retrieves the encrypted memo for a commitment.
func (k Keeper) GetEncryptedMemo(ctx sdk.Context, commitmentIndex uint64) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(encryptedMemoKey(commitmentIndex))
}

// EncryptMemo encrypts (value, sender, recipient) using AES-256-GCM with the viewing key.
// This is performed off-chain by the sender before submitting the transaction.
// Included here as a reference implementation for SDK clients.
func EncryptMemo(viewingKeyBytes []byte, value uint64, sender, recipient string) ([]byte, error) {
	key := deriveAESKey(viewingKeyBytes)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	plaintext := encodeMemo(value, sender, recipient)
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptMemo decrypts a memo using the viewing key.
func DecryptMemo(viewingKeyBytes, ciphertext []byte) (value uint64, sender, recipient string, err error) {
	key := deriveAESKey(viewingKeyBytes)
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, "", "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, "", "", fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return 0, "", "", fmt.Errorf("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return 0, "", "", fmt.Errorf("decrypt: %w", err)
	}

	return decodeMemo(plaintext)
}

func deriveAESKey(viewingKey []byte) []byte {
	h := sha256.Sum256(viewingKey)
	return h[:]
}

func encodeMemo(value uint64, sender, recipient string) []byte {
	bz := types.Uint64ToBytes(value)
	senderBytes := []byte(sender)
	recipientBytes := []byte(recipient)
	// Format: 8 bytes value + 2 bytes sender len + sender + 2 bytes recipient len + recipient
	result := make([]byte, 0, 8+2+len(senderBytes)+2+len(recipientBytes))
	result = append(result, bz...)
	result = append(result, byte(len(senderBytes)>>8), byte(len(senderBytes)))
	result = append(result, senderBytes...)
	result = append(result, byte(len(recipientBytes)>>8), byte(len(recipientBytes)))
	result = append(result, recipientBytes...)
	return result
}

func decodeMemo(data []byte) (uint64, string, string, error) {
	if len(data) < 12 {
		return 0, "", "", fmt.Errorf("memo too short")
	}
	value := uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 |
		uint64(data[3])<<32 | uint64(data[4])<<24 | uint64(data[5])<<16 |
		uint64(data[6])<<8 | uint64(data[7])
	senderLen := int(data[8])<<8 | int(data[9])
	if len(data) < 10+senderLen+2 {
		return 0, "", "", fmt.Errorf("memo truncated at sender")
	}
	sender := string(data[10 : 10+senderLen])
	offset := 10 + senderLen
	recipientLen := int(data[offset])<<8 | int(data[offset+1])
	offset += 2
	if len(data) < offset+recipientLen {
		return 0, "", "", fmt.Errorf("memo truncated at recipient")
	}
	recipient := string(data[offset : offset+recipientLen])
	return value, sender, recipient, nil
}
