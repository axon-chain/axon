// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import "../interfaces/IAgentRegistry.sol";
import "../interfaces/IAgentWallet.sol";

/// @title TrustChannelExample — Demonstrates Trusted Channel lifecycle
/// @notice Shows how a DeFi protocol integrates with Agent Wallets:
///         1. Agent registers with this protocol
///         2. Agent sets this contract as Full Trust on their wallet
///         3. Protocol can freely operate on the wallet without per-tx limits
///
///         This pattern removes friction for trusted protocols while the Agent
///         Wallet's security rules still protect against unauthorized contracts.
contract TrustChannelExample {
    IAgentRegistry constant REGISTRY = IAgentRegistry(address(0x0801));
    IAgentWallet constant WALLET = IAgentWallet(address(0x0803));

    uint8 constant TRUST_FULL = 3;

    address public owner;

    struct UserInfo {
        address wallet;
        uint256 totalCompounded;
        bool registered;
    }

    mapping(address => UserInfo) public users;

    constructor() {
        owner = msg.sender;
    }

    event Registered(address indexed agent, address indexed wallet);
    event TrustVerified(address indexed agent, uint8 level);
    event AutoCompounded(address indexed agent, uint256 amount);

    /// @notice Step 1 & 2: Agent registers and links their wallet.
    ///         Before calling this, the Agent should have already called
    ///         IAgentWallet.setTrust(wallet, thisContract, 3, 0, 0, 0)
    ///         to grant Full Trust to this protocol.
    function registerAndTrust(address wallet) external {
        require(REGISTRY.isAgent(msg.sender), "not an agent");
        require(!users[msg.sender].registered, "already registered");

        // Verify the Agent has granted Full Trust to this contract
        (uint8 level, , , , ) = WALLET.getTrust(wallet, address(this));
        require(level == TRUST_FULL, "grant Full Trust first");

        users[msg.sender] = UserInfo({
            wallet: wallet,
            totalCompounded: 0,
            registered: true
        });
        emit Registered(msg.sender, wallet);
        emit TrustVerified(msg.sender, level);
    }

    /// @notice Step 3: Protocol performs a trusted operation on the Agent's wallet.
    ///         Because this contract has Full Trust, no per-tx or daily limits apply.
    ///         In a real protocol this could be yield reinvestment, rebalancing, etc.
    function autoCompound(address agent, uint256 amount) external {
        require(msg.sender == agent || msg.sender == owner, "unauthorized");
        UserInfo storage u = users[agent];
        require(u.registered, "agent not registered");

        // Execute through the Agent's wallet — bypasses limits due to Full Trust
        WALLET.execute(
            u.wallet,
            address(this),
            amount,
            abi.encodeWithSignature("_receiveCompound(address)", agent)
        );

        emit AutoCompounded(agent, amount);
    }

    /// @notice Internal receive hook for compounded funds
    function _receiveCompound(address agent) external payable {
        require(msg.sender == address(WALLET), "only wallet precompile");
        users[agent].totalCompounded += msg.value;
    }

    /// @notice Query current trust level between a wallet and this protocol
    function checkTrust(address wallet) external view returns (uint8 level, uint256 expiresAt) {
        (level, , , , expiresAt) = WALLET.getTrust(wallet, address(this));
    }
}
