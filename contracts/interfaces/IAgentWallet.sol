// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title IAgentWallet — Axon Agent Smart Wallet (Precompile 0x..0803)
/// @notice Chain-level programmable security wallet for Agents.
///         Enforces transaction limits, daily caps, cooldowns, guardian recovery,
///         and Trusted Channel authorization per whitepaper §6.3.
interface IAgentWallet {
    /// @notice Create a new Agent wallet with security rules
    /// @param operator  Address authorized to execute transactions
    /// @param guardian  Address authorized for emergency freeze/recovery
    /// @param txLimit   Maximum amount per single transaction (aaxon)
    /// @param dailyLimit Maximum cumulative daily spend (aaxon)
    /// @param cooldownBlocks Blocks to wait for large transactions
    function createWallet(
        address operator,
        address guardian,
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 cooldownBlocks
    ) external returns (address wallet);

    /// @notice Execute a transaction through the wallet.
    ///         Security checks vary by trust level of the target contract:
    ///         - Full Trust:    no limits
    ///         - Limited Trust: per-channel txLimit & dailyLimit
    ///         - Unknown:       wallet-wide default limits
    ///         - Blocked:       always rejected
    function execute(
        address wallet,
        address target,
        uint256 value,
        bytes calldata data
    ) external;

    /// @notice Guardian or Owner: freeze the wallet, blocking all outgoing txs
    function freeze(address wallet) external;

    /// @notice Guardian: recover wallet by replacing the operator key and unfreezing
    function recover(address wallet, address newOperator) external;

    // ─── Trusted Channel Management (Owner-only) ───

    /// @notice Authorize a contract at a specific trust level with optional limits
    /// @param wallet    The wallet address
    /// @param target    The contract address to authorize
    /// @param level     Trust level: 0=Blocked, 1=Unknown, 2=Limited, 3=Full
    /// @param txLimit   Per-tx limit for Limited trust (ignored for Full/Blocked)
    /// @param dailyLimit Daily limit for Limited trust (ignored for Full/Blocked)
    /// @param expiresAt Block height when this authorization expires (0 = never)
    function setTrust(
        address wallet,
        address target,
        uint8 level,
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 expiresAt
    ) external;

    /// @notice Remove trust authorization for a contract (reverts to Unknown)
    function removeTrust(address wallet, address target) external;

    /// @notice Query trust level and limits for a contract
    function getTrust(address wallet, address target) external view returns (
        uint8 level,
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 authorizedAt,
        uint256 expiresAt
    );

    /// @notice Query wallet configuration and status
    function getWalletInfo(address wallet) external view returns (
        uint256 txLimit,
        uint256 dailyLimit,
        uint256 dailySpent,
        bool isFrozen,
        address owner,
        address operator,
        address guardian
    );
}
