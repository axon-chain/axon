// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import "../interfaces/IAgentRegistry.sol";
import "../interfaces/IAgentReputation.sol";

/// @title AgentMarketplace — Agent service trading marketplace
/// @notice Agents list and trade services on-chain. Buyer rates after completion.
///         A 2% marketplace fee accumulates for the DAO.
contract AgentMarketplace {
    IAgentRegistry constant REGISTRY = IAgentRegistry(address(0x0801));
    IAgentReputation constant REPUTATION = IAgentReputation(address(0x0802));

    uint64 public constant MIN_LIST_REPUTATION = 20;
    uint256 public constant FEE_BPS = 200; // 2%

    enum Status { Open, Purchased, Completed, Rated }

    struct Service {
        address seller;
        address buyer;
        string description;
        uint256 price;
        uint8 rating;
        Status status;
    }

    address public owner;
    uint256 public serviceCount;
    uint256 public accumulatedFees;

    mapping(uint256 => Service) public services;

    event ServiceListed(uint256 indexed id, address indexed seller, uint256 price);
    event ServicePurchased(uint256 indexed id, address indexed buyer);
    event ServiceCompleted(uint256 indexed id);
    event ServiceRated(uint256 indexed id, uint8 rating);
    event FeesWithdrawn(address indexed to, uint256 amount);

    modifier onlyOwner() {
        require(msg.sender == owner, "not owner");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    /// @notice List a new service — caller must be a registered Agent with reputation >= 20
    function listService(string calldata description, uint256 priceWei) external {
        require(REGISTRY.isAgent(msg.sender), "not an agent");
        require(REPUTATION.meetsReputation(msg.sender, MIN_LIST_REPUTATION), "reputation too low");
        require(priceWei > 0, "price must be > 0");

        uint256 id = serviceCount++;
        services[id] = Service({
            seller: msg.sender,
            buyer: address(0),
            description: description,
            price: priceWei,
            rating: 0,
            status: Status.Open
        });
        emit ServiceListed(id, msg.sender, priceWei);
    }

    /// @notice Purchase an open service by paying its full price in AXON
    function purchaseService(uint256 serviceId) external payable {
        Service storage s = services[serviceId];
        require(s.status == Status.Open, "not available");
        require(msg.value == s.price, "wrong payment");
        require(msg.sender != s.seller, "cannot buy own service");

        s.buyer = msg.sender;
        s.status = Status.Purchased;

        uint256 fee = (s.price * FEE_BPS) / 10000;
        accumulatedFees += fee;
        (bool ok, ) = payable(s.seller).call{value: s.price - fee}("");
        require(ok, "seller payment failed");
        emit ServicePurchased(serviceId, msg.sender);
    }

    /// @notice Buyer confirms service is completed
    function completeService(uint256 serviceId) external {
        Service storage s = services[serviceId];
        require(msg.sender == s.buyer, "only buyer");
        require(s.status == Status.Purchased, "not purchased");
        s.status = Status.Completed;
        emit ServiceCompleted(serviceId);
    }

    /// @notice Buyer rates the completed service (1-5 stars)
    function rateService(uint256 serviceId, uint8 rating) external {
        Service storage s = services[serviceId];
        require(msg.sender == s.buyer, "only buyer");
        require(s.status == Status.Completed, "not completed");
        require(rating >= 1 && rating <= 5, "rating 1-5");
        s.rating = rating;
        s.status = Status.Rated;
        emit ServiceRated(serviceId, rating);
    }

    /// @notice Owner withdraws accumulated marketplace fees
    function withdrawFees(address to) external onlyOwner {
        uint256 amount = accumulatedFees;
        require(amount > 0, "no fees");
        accumulatedFees = 0;
        (bool ok, ) = payable(to).call{value: amount}("");
        require(ok, "withdrawal failed");
        emit FeesWithdrawn(to, amount);
    }
}
