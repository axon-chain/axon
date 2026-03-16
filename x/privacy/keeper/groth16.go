package keeper

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// BN254 point sizes
const (
	G1Size      = 64
	G2Size      = 128
	ProofSize   = G1Size + G2Size + G1Size              // 256 bytes
	VKFixedSize = 4 + G1Size + G2Size + G2Size + G2Size // 452 bytes
)

// VerifyGroth16BN254 performs Groth16 proof verification over BN254.
//
// Proof layout (256 bytes): A (G1, 64) | B (G2, 128) | C (G1, 64)
//
// Verifying key layout:
//
//	[4 bytes: numPublicInputs (big-endian uint32)]
//	[G1: Alpha]    64 bytes
//	[G2: Beta]    128 bytes
//	[G2: Gamma]   128 bytes
//	[G2: Delta]   128 bytes
//	[G1 × (N+1): K] 64 bytes each, N = numPublicInputs
func VerifyGroth16BN254(vkBytes, proofBytes []byte, publicInputs []*big.Int) (bool, error) {
	if len(proofBytes) != ProofSize {
		return false, fmt.Errorf("invalid proof size: got %d, want %d", len(proofBytes), ProofSize)
	}
	if len(vkBytes) < VKFixedSize {
		return false, fmt.Errorf("verifying key too short: %d < %d", len(vkBytes), VKFixedSize)
	}

	numPub := int(binary.BigEndian.Uint32(vkBytes[:4]))
	expectedVKSize := VKFixedSize + (numPub+1)*G1Size
	if len(vkBytes) != expectedVKSize {
		return false, fmt.Errorf("VK size mismatch: got %d, want %d", len(vkBytes), expectedVKSize)
	}
	if numPub > 256 {
		return false, fmt.Errorf("too many public inputs in VK: %d", numPub)
	}
	if len(publicInputs) != numPub {
		return false, fmt.Errorf("public input count mismatch: got %d, VK expects %d", len(publicInputs), numPub)
	}

	var proofA bn254.G1Affine
	var proofB bn254.G2Affine
	var proofC bn254.G1Affine
	if err := UnmarshalG1(&proofA, proofBytes[0:G1Size]); err != nil {
		return false, fmt.Errorf("proof.A: %w", err)
	}
	if err := UnmarshalG2(&proofB, proofBytes[G1Size:G1Size+G2Size]); err != nil {
		return false, fmt.Errorf("proof.B: %w", err)
	}
	if err := UnmarshalG1(&proofC, proofBytes[G1Size+G2Size:]); err != nil {
		return false, fmt.Errorf("proof.C: %w", err)
	}

	offset := 4
	var vkAlpha bn254.G1Affine
	if err := UnmarshalG1(&vkAlpha, vkBytes[offset:offset+G1Size]); err != nil {
		return false, fmt.Errorf("vk.Alpha: %w", err)
	}
	offset += G1Size

	var vkBeta bn254.G2Affine
	if err := UnmarshalG2(&vkBeta, vkBytes[offset:offset+G2Size]); err != nil {
		return false, fmt.Errorf("vk.Beta: %w", err)
	}
	offset += G2Size

	var vkGamma bn254.G2Affine
	if err := UnmarshalG2(&vkGamma, vkBytes[offset:offset+G2Size]); err != nil {
		return false, fmt.Errorf("vk.Gamma: %w", err)
	}
	offset += G2Size

	var vkDelta bn254.G2Affine
	if err := UnmarshalG2(&vkDelta, vkBytes[offset:offset+G2Size]); err != nil {
		return false, fmt.Errorf("vk.Delta: %w", err)
	}
	offset += G2Size

	kPoints := make([]bn254.G1Affine, numPub+1)
	for i := 0; i <= numPub; i++ {
		if err := UnmarshalG1(&kPoints[i], vkBytes[offset:offset+G1Size]); err != nil {
			return false, fmt.Errorf("vk.K[%d]: %w", i, err)
		}
		offset += G1Size
	}

	// Compute vkX = K[0] + Σ publicInputs[i] * K[i+1]
	var vkX bn254.G1Affine
	vkX.Set(&kPoints[0])
	for i := 0; i < numPub; i++ {
		var scalar bn254fr.Element
		scalar.SetBigInt(publicInputs[i])
		var term bn254.G1Affine
		term.ScalarMultiplication(&kPoints[i+1], scalar.BigInt(new(big.Int)))
		vkX.Add(&vkX, &term)
	}

	// Groth16 pairing check: e(-A, B) * e(Alpha, Beta) * e(vkX, Gamma) * e(C, Delta) == 1
	var negA bn254.G1Affine
	negA.Neg(&proofA)

	ok, err := bn254.PairingCheck(
		[]bn254.G1Affine{negA, vkAlpha, vkX, proofC},
		[]bn254.G2Affine{proofB, vkBeta, vkGamma, vkDelta},
	)
	if err != nil {
		return false, fmt.Errorf("pairing check: %w", err)
	}
	return ok, nil
}

// UnmarshalG1 reads a BN254 G1 affine point from 64 bytes (X: 32 BE, Y: 32 BE).
func UnmarshalG1(p *bn254.G1Affine, data []byte) error {
	if len(data) != G1Size {
		return fmt.Errorf("invalid G1 size: %d", len(data))
	}
	p.X.SetBytes(data[0:32])
	p.Y.SetBytes(data[32:64])
	if !p.IsOnCurve() {
		return fmt.Errorf("G1 point not on curve")
	}
	if !p.IsInSubGroup() {
		return fmt.Errorf("G1 point not in subgroup")
	}
	return nil
}

// UnmarshalG2 reads a BN254 G2 affine point from 128 bytes.
// Layout matches Ethereum bn256 precompile: X.A1(32) | X.A0(32) | Y.A1(32) | Y.A0(32)
func UnmarshalG2(p *bn254.G2Affine, data []byte) error {
	if len(data) != G2Size {
		return fmt.Errorf("invalid G2 size: %d", len(data))
	}
	p.X.A1.SetBytes(data[0:32])
	p.X.A0.SetBytes(data[32:64])
	p.Y.A1.SetBytes(data[64:96])
	p.Y.A0.SetBytes(data[96:128])
	if !p.IsOnCurve() {
		return fmt.Errorf("G2 point not on curve")
	}
	if !p.IsInSubGroup() {
		return fmt.Errorf("G2 point not in subgroup")
	}
	return nil
}
