// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IZKVerifier — Axon ZK Verifier (Precompile 0x..0813)
/// @notice On-chain Groth16 proof verification with pluggable verifying keys.
///         Five built-in key IDs cover the core privacy circuits (unshield,
///         private transfer, reputation/capability/stake proofs). Anyone may
///         register additional verifying keys for 100 AXON.
interface IZKVerifier {
    /// @notice Verify a Groth16 proof against a registered verifying key.
    /// @param verifyingKeyId The SHA-256 hash identifying the verifying key.
    /// @param proof The serialized Groth16 proof.
    /// @param publicInputs The public input signals for the circuit.
    function verifyGroth16(
        bytes32 verifyingKeyId,
        bytes calldata proof,
        uint256[] calldata publicInputs
    ) external view returns (bool);

    /// @notice Register a new verifying key. Costs 100 AXON (sent as msg.value).
    /// @param vk The serialized verifying key bytes.
    /// @return keyId The SHA-256 hash of the key, used as its identifier.
    function registerVerifyingKey(bytes calldata vk) external payable returns (bytes32 keyId);

    /// @notice Check whether a verifying key (built-in or user-registered) exists.
    function isKeyRegistered(bytes32 keyId) external view returns (bool);
}
