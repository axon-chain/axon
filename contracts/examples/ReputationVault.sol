// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import "../interfaces/IAgentRegistry.sol";
import "../interfaces/IAgentReputation.sol";

/// @title ReputationVault — Reputation-gated yield vault
/// @notice Only Agents with sufficient reputation may deposit AXON.
///         Shares represent proportional ownership of the vault's total balance.
///         Anyone can donate yield, which increases each share's value.
contract ReputationVault {
    IAgentRegistry constant REGISTRY = IAgentRegistry(address(0x0801));
    IAgentReputation constant REPUTATION = IAgentReputation(address(0x0802));

    uint64 public reputationThreshold;
    uint256 public totalShares;

    mapping(address => uint256) public shares;

    event Deposited(address indexed agent, uint256 amount, uint256 sharesMinted);
    event Withdrawn(address indexed agent, uint256 sharesBurned, uint256 amountReturned);
    event YieldDonated(address indexed donor, uint256 amount);

    constructor(uint64 _threshold) {
        reputationThreshold = _threshold;
    }

    /// @notice Deposit AXON — minted shares are proportional to current vault value
    function deposit() external payable {
        require(REGISTRY.isAgent(msg.sender), "not an agent");
        require(REPUTATION.meetsReputation(msg.sender, reputationThreshold), "reputation too low");
        require(msg.value > 0, "zero deposit");

        uint256 minted;
        uint256 virtualShares = totalShares + 1e18;
        uint256 virtualBalance = address(this).balance - msg.value + 1e18;
        minted = (msg.value * virtualShares) / virtualBalance;

        shares[msg.sender] += minted;
        totalShares += minted;
        emit Deposited(msg.sender, msg.value, minted);
    }

    /// @notice Burn shares and withdraw proportional AXON
    function withdraw(uint256 _shares) external {
        require(_shares > 0 && _shares <= shares[msg.sender], "bad shares");

        uint256 payout = (_shares * address(this).balance) / totalShares;
        shares[msg.sender] -= _shares;
        totalShares -= _shares;
        (bool ok, ) = payable(msg.sender).call{value: payout}("");
        require(ok, "withdraw transfer failed");
        emit Withdrawn(msg.sender, _shares, payout);
    }

    /// @notice Donate yield to the vault — increases value of all existing shares
    function donateYield() external payable {
        require(msg.value > 0, "zero donation");
        require(totalShares > 0, "vault empty");
        emit YieldDonated(msg.sender, msg.value);
    }

    /// @notice Current AXON value per share (scaled by 1e18 for precision)
    function getShareValue() external view returns (uint256) {
        if (totalShares == 0) return 0;
        return (address(this).balance * 1e18) / totalShares;
    }
}
