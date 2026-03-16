// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IPrivateTransfer — Axon Shielded Pool (Precompile 0x..0811)
/// @notice Privacy-preserving transfers backed by a Merkle commitment tree
///         and ZK nullifier proofs maintained by consensus.
interface IPrivateTransfer {
    /// @notice Deposit msg.value into the shielded pool.
    /// @param commitment Pedersen commitment to (value, blinding, recipient).
    function shield(bytes32 commitment) external payable;

    /// @notice Withdraw from the shielded pool with a ZK proof.
    /// @param proof       Serialised ZK proof (gnark/groth16).
    /// @param merkleRoot  A recent valid Merkle root.
    /// @param nullifier   Unique nullifier derived from the spent note.
    /// @param recipient   Destination address for the withdrawn funds.
    /// @param amount      Amount to withdraw (must match the note value).
    function unshield(
        bytes calldata proof,
        bytes32 merkleRoot,
        bytes32 nullifier,
        address recipient,
        uint256 amount
    ) external;

    /// @notice Transfer value privately within the shielded pool.
    /// @param proof             Serialised ZK proof (gnark/groth16).
    /// @param merkleRoot        A recent valid Merkle root.
    /// @param inputNullifiers   Nullifiers of the two consumed notes.
    /// @param outputCommitments Commitments of the two newly created notes.
    function privateTransfer(
        bytes calldata proof,
        bytes32 merkleRoot,
        bytes32[2] calldata inputNullifiers,
        bytes32[2] calldata outputCommitments
    ) external;

    /// @notice Check whether a Merkle root is known to the tree.
    function isKnownRoot(bytes32 root) external view returns (bool);

    /// @notice Check whether a nullifier has already been spent.
    function isSpent(bytes32 nullifier) external view returns (bool);

    /// @notice Return the current number of leaves in the commitment tree.
    function getTreeSize() external view returns (uint256);
}
