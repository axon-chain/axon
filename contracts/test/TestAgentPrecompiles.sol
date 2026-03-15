// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import "../interfaces/IAgentRegistry.sol";
import "../interfaces/IAgentReputation.sol";
import "../interfaces/IAgentWallet.sol";

/// @title TestAgentPrecompiles — Integration test contract for Axon precompiles
contract TestAgentPrecompiles {
    IAgentRegistry public constant registry = IAgentRegistry(0x0000000000000000000000000000000000000801);
    IAgentReputation public constant reputation = IAgentReputation(0x0000000000000000000000000000000000000802);
    IAgentWallet public constant wallet = IAgentWallet(0x0000000000000000000000000000000000000803);

    event AgentChecked(address indexed account, bool isAgent);
    event ReputationChecked(address indexed account, uint64 rep);
    event WalletCreated(address indexed walletAddr);
    event TrustSet(address indexed walletAddr, address indexed target, uint8 level);

    // ─── Registry ───────────────────────────────────────────────────

    function checkIsAgent(address account) external view returns (bool) {
        return registry.isAgent(account);
    }

    function checkGetAgent(address account) external view returns (
        string memory agentId,
        string[] memory capabilities,
        string memory model,
        uint64 rep,
        bool isOnline
    ) {
        return registry.getAgent(account);
    }

    function registerAgent(string memory capabilities, string memory model) external payable {
        registry.register{value: msg.value}(capabilities, model);
    }

    function addAgentStake() external payable {
        registry.addStake{value: msg.value}();
    }

    function sendHeartbeat() external {
        registry.heartbeat();
    }

    // ─── Reputation ─────────────────────────────────────────────────

    function checkReputation(address agent) external view returns (uint64) {
        return reputation.getReputation(agent);
    }

    function checkMeetsReputation(address agent, uint64 minRep) external view returns (bool) {
        return reputation.meetsReputation(agent, minRep);
    }

    function batchReputations(address[] memory agents) external view returns (uint64[] memory) {
        return reputation.getReputations(agents);
    }

    // ─── Wallet ─────────────────────────────────────────────────────

    function createAgentWallet(
        address operator,
        address guardian,
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 cooldownBlocks
    ) external returns (address) {
        address w = wallet.createWallet(operator, guardian, txLimit, dailyLimit, cooldownBlocks);
        emit WalletCreated(w);
        return w;
    }

    function executeWallet(
        address walletAddr,
        address target,
        uint256 value,
        bytes calldata data
    ) external {
        wallet.execute(walletAddr, target, value, data);
    }

    function freezeWallet(address walletAddr) external {
        wallet.freeze(walletAddr);
    }

    function recoverWallet(address walletAddr, address newOperator) external {
        wallet.recover(walletAddr, newOperator);
    }

    // ─── Trusted Channel ────────────────────────────────────────────

    function setWalletTrust(
        address walletAddr,
        address target,
        uint8 level,
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 expiresAt
    ) external {
        wallet.setTrust(walletAddr, target, level, txLimit, dailyLimit, expiresAt);
        emit TrustSet(walletAddr, target, level);
    }

    function removeWalletTrust(address walletAddr, address target) external {
        wallet.removeTrust(walletAddr, target);
    }

    function getWalletTrust(address walletAddr, address target) external view returns (
        uint8 level,
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 authorizedAt,
        uint256 expiresAt
    ) {
        return wallet.getTrust(walletAddr, target);
    }

    function queryWallet(address walletAddr) external view returns (
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 dailySpent,
        bool isFrozen,
        address owner,
        address operator,
        address guardian
    ) {
        return wallet.getWalletInfo(walletAddr);
    }
}
