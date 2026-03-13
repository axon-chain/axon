// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IAgentReputation — Axon Reputation Query (Precompile 0x..0802)
/// @notice Read-only access to chain-level Agent reputation scores.
///         Reputation is maintained by consensus, not by any contract.
interface IAgentReputation {
    /// @notice Get reputation score for a single Agent (0-100)
    function getReputation(address agent) external view returns (uint64);

    /// @notice Batch query reputation for multiple Agents
    function getReputations(address[] memory agents) external view returns (uint64[] memory);

    /// @notice Check if Agent meets a minimum reputation threshold
    function meetsReputation(address agent, uint64 minReputation) external view returns (bool);
}
