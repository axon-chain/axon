// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

/// @title Faucet — Axon Testnet Token Faucet
/// @notice Dispenses test AXON to developers. Limited to once per 24 hours per address.
contract Faucet {
    address public owner;
    uint256 public dripAmount;
    uint256 public cooldownTime;
    mapping(address => uint256) public lastDrip;

    event Dripped(address indexed recipient, uint256 amount);
    event Funded(address indexed funder, uint256 amount);
    event DripAmountChanged(uint256 oldAmount, uint256 newAmount);

    modifier onlyOwner() {
        require(msg.sender == owner, "only owner");
        _;
    }

    constructor(uint256 _dripAmount, uint256 _cooldownTime) payable {
        owner = msg.sender;
        dripAmount = _dripAmount;
        cooldownTime = _cooldownTime;
    }

    /// @notice Request test tokens
    function drip() external {
        require(
            block.timestamp >= lastDrip[msg.sender] + cooldownTime,
            "cooldown not expired"
        );
        require(address(this).balance >= dripAmount, "faucet empty");

        lastDrip[msg.sender] = block.timestamp;
        payable(msg.sender).transfer(dripAmount);

        emit Dripped(msg.sender, dripAmount);
    }

    /// @notice Request tokens for a specific address
    function dripTo(address recipient) external {
        require(
            block.timestamp >= lastDrip[recipient] + cooldownTime,
            "cooldown not expired"
        );
        require(address(this).balance >= dripAmount, "faucet empty");

        lastDrip[recipient] = block.timestamp;
        payable(recipient).transfer(dripAmount);

        emit Dripped(recipient, dripAmount);
    }

    /// @notice Check time remaining until next drip
    function timeUntilNextDrip(address account) external view returns (uint256) {
        uint256 nextDrip = lastDrip[account] + cooldownTime;
        if (block.timestamp >= nextDrip) return 0;
        return nextDrip - block.timestamp;
    }

    /// @notice Update drip amount (owner only)
    function setDripAmount(uint256 newAmount) external onlyOwner {
        uint256 old = dripAmount;
        dripAmount = newAmount;
        emit DripAmountChanged(old, newAmount);
    }

    /// @notice Fund the faucet
    receive() external payable {
        emit Funded(msg.sender, msg.value);
    }
}
