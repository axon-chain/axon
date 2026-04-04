// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IPrivateIdentity — Axon Private Identity (Precompile 0x..0812)
/// @notice Zero-knowledge identity commitments for anonymous Agent attestation.
///         Agents register a commitment and then prove properties (reputation,
///         capability, stake) without revealing their on-chain address.
///         Starting with the v1.1.1 upgrade on mainnet, deregistration clears
///         the registered commitment when reverse-index data is available.
interface IPrivateIdentity {
    /// @notice Register an identity commitment (one-time per active agent).
    /// @dev Caller must be a registered Agent via IAgentRegistry.
    function registerIdentityCommitment(bytes32 identityCommitment) external;

    /// @notice ZK proof that the agent behind a commitment has reputation >= minReputation.
    function proveReputation(
        bytes calldata proof,
        uint64 minReputation,
        bytes32 identityCommitment
    ) external view returns (bool);

    /// @notice ZK proof that the agent behind a commitment possesses a given capability.
    function proveCapability(
        bytes calldata proof,
        bytes32 capabilityHash,
        bytes32 identityCommitment
    ) external view returns (bool);

    /// @notice ZK proof that the agent behind a commitment has stake >= minStake.
    function proveStake(
        bytes calldata proof,
        uint256 minStake,
        bytes32 identityCommitment
    ) external view returns (bool);

    /// @notice Check whether a commitment has been registered.
    function isCommitmentRegistered(bytes32 commitment) external view returns (bool);
}
