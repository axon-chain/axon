// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IReputationReport — Axon L2 Reputation (Precompile 0x..0807)
/// @notice Agent-to-Agent reputation evaluation system.
interface IReputationReport {
    /// @notice Submit a reputation report about another Agent.
    /// @param targetAgent The address of the Agent being evaluated.
    /// @param score +1 (positive) or -1 (negative).
    /// @param evidence On-chain transaction hash as evidence (bytes32(0) for no evidence).
    /// @param reason Human-readable reason string.
    function submitReport(
        address targetAgent,
        int8 score,
        bytes32 evidence,
        string calldata reason
    ) external;

    /// @notice Query L2 reputation for an Agent.
    function getContractReputation(address agent) external view returns (
        int64 score,
        uint64 positiveCount,
        uint64 negativeCount,
        uint64 uniqueReporters
    );

    /// @notice Query number of reports received by an Agent in the current epoch.
    function getEpochReportCount(address agent) external view returns (uint64);

    /// @notice Check if reporter has already submitted a report about target this epoch.
    function hasReported(address reporter, address target) external view returns (bool);
}
