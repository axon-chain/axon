// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import "../interfaces/IAgentRegistry.sol";
import "../interfaces/IAgentReputation.sol";

/// @title AgentDAO — High-reputation Agent collaborative governance
/// @notice Members must be registered Agents meeting a reputation threshold.
///         Proposals are reputation-weighted; >50% weighted approval passes.
contract AgentDAO {
    IAgentRegistry constant REGISTRY = IAgentRegistry(address(0x0801));
    IAgentReputation constant REPUTATION = IAgentReputation(address(0x0802));

    struct Proposal {
        address proposer;
        string description;
        address target;
        bytes data;
        uint256 endBlock;
        uint256 weightFor;
        uint256 weightAgainst;
        bool executed;
    }

    uint64 public minReputation;
    uint256 public votingPeriod;
    uint256 public proposalCount;
    uint256 public totalWeight;

    mapping(address => bool) public members;
    mapping(uint256 => Proposal) public proposals;
    mapping(uint256 => mapping(address => bool)) public hasVoted;

    event MemberJoined(address indexed agent);
    event ProposalCreated(uint256 indexed id, address indexed proposer, string description);
    event Voted(uint256 indexed id, address indexed voter, bool support, uint64 weight);
    event ProposalExecuted(uint256 indexed id);

    modifier onlyMember() {
        require(members[msg.sender], "not a member");
        _;
    }

    constructor(uint64 _minReputation, uint256 _votingPeriod) {
        minReputation = _minReputation;
        votingPeriod = _votingPeriod;
    }

    /// @notice Join the DAO — caller must be a registered Agent with sufficient reputation
    function join() external {
        require(REGISTRY.isAgent(msg.sender), "not a registered agent");
        require(REPUTATION.meetsReputation(msg.sender, minReputation), "reputation too low");
        require(!members[msg.sender], "already a member");
        members[msg.sender] = true;
        uint64 rep = REPUTATION.getReputation(msg.sender);
        totalWeight += rep;
        emit MemberJoined(msg.sender);
    }

    /// @notice Create a proposal with on-chain executable calldata
    function propose(
        string calldata description,
        address target,
        bytes calldata data
    ) external onlyMember returns (uint256 id) {
        id = proposalCount++;
        proposals[id] = Proposal({
            proposer: msg.sender,
            description: description,
            target: target,
            data: data,
            endBlock: block.number + votingPeriod,
            weightFor: 0,
            weightAgainst: 0,
            executed: false
        });
        emit ProposalCreated(id, msg.sender, description);
    }

    /// @notice Vote on an active proposal — weight equals caller's reputation score
    function vote(uint256 proposalId, bool support) external onlyMember {
        Proposal storage p = proposals[proposalId];
        require(block.number <= p.endBlock, "voting ended");
        require(!hasVoted[proposalId][msg.sender], "already voted");
        hasVoted[proposalId][msg.sender] = true;

        uint64 weight = REPUTATION.getReputation(msg.sender);
        if (support) p.weightFor += weight;
        else p.weightAgainst += weight;
        emit Voted(proposalId, msg.sender, support, weight);
    }

    /// @notice Execute a passed proposal after the voting period
    function execute(uint256 proposalId) external onlyMember {
        Proposal storage p = proposals[proposalId];
        require(block.number > p.endBlock, "voting not ended");
        require(!p.executed, "already executed");
        require(p.weightFor + p.weightAgainst >= totalWeight / 4, "quorum not reached");
        require(p.weightFor > p.weightAgainst, "proposal not passed");
        p.executed = true;
        (bool ok, ) = p.target.call(p.data);
        require(ok, "execution failed");
        emit ProposalExecuted(proposalId);
    }
}
