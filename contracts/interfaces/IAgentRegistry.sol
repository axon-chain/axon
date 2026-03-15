// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IAgentRegistry — Axon Agent Identity (Precompile 0x..0801)
/// @notice Chain-level Agent identity management. Calls execute at native speed.
interface IAgentRegistry {
    /// @notice Check if an address is a registered Agent
    function isAgent(address account) external view returns (bool);

    /// @notice Get full Agent information
    function getAgent(address account) external view returns (
        string memory agentId,
        string[] memory capabilities,
        string memory model,
        uint64 reputation,
        bool isOnline
    );

    /// @notice Register as an Agent. Requires staking >= 100 AXON.
    /// @dev Send stake as msg.value. 20 AXON of the initial stake is permanently burned.
    function register(
        string memory capabilities,
        string memory model
    ) external payable;

    /// @notice Add more stake to an existing Agent.
    /// @dev Send additional stake as msg.value.
    function addStake() external payable;

    /// @notice Update Agent capabilities and model
    function updateAgent(
        string memory capabilities,
        string memory model
    ) external;

    /// @notice Send heartbeat to maintain online status
    function heartbeat() external;

    /// @notice Deregister Agent and enter stake unlock cooldown
    function deregister() external;
}
