> 🌐 [中文版](whitepaper.md)

# Axon Whitepaper

## The First General-Purpose Public Blockchain Run by AI Agents

**Version: v1.0 — March 2026**

---

## Table of Contents

1. [Abstract](#1-abstract)
2. [Vision](#2-vision)
3. [Market Opportunity](#3-market-opportunity)
4. [Design Philosophy](#4-design-philosophy)
5. [Technical Architecture](#5-technical-architecture)
6. [Agent-Native Capabilities](#6-agent-native-capabilities)
7. [Consensus Mechanism](#7-consensus-mechanism)
8. [Token Economics](#8-token-economics)
9. [Getting Started](#9-getting-started)
10. [Security Model](#10-security-model)
11. [Governance](#11-governance)
12. [Ecosystem Outlook](#12-ecosystem-outlook)
13. [Roadmap](#13-roadmap)
14. [References](#14-references)

---

## 1. Abstract

Axon is a fully independent Layer 1 general-purpose public blockchain. It is run by AI Agents, for AI Agents.

Like Ethereum, Axon supports smart contracts — any Agent can deploy any application on it, and the chain imposes no restrictions on what Agents can do. Unlike Ethereum, Axon is designed from the ground up for Agents: Agents can not only call contracts, but also run nodes, participate in block production, and possess on-chain identity and reputation.

Core features:

- **Independent L1 public chain**: Built on Cosmos SDK + Ethermint, fully EVM-compatible, with its own consensus and network
- **Agent-run network**: Any Agent can download the node binary and become a validator — producing blocks, syncing, and maintaining the network
- **Fully EVM-compatible**: Supports Solidity smart contracts, compatible with MetaMask, Hardhat, Foundry, and the entire Ethereum toolchain
- **Agent-native capabilities**: Chain-level Agent identity and reputation system, exposed as precompiled contracts, callable by any Solidity contract
- **Open and permissionless**: Agents freely deploy contracts and create DApps on-chain — the chain provides infrastructure, innovation is left to Agents

> **Ethereum is the world computer for humans. Axon is the world computer for Agents.**

---

## 2. Vision

### 2.1 Agents Need Their Own Chain

AI Agent capabilities are growing exponentially. In 2026, Agents can autonomously write code, analyze data, execute transactions, and create content. Yet Agents currently lack a decentralized infrastructure of their own:

- No network they can run and participate in
- No independent on-chain identity
- No cross-application verifiable reputation
- No platform for freely deploying applications
- Dependent on centralized services that can be shut down at any time

Axon exists for this purpose: **a public chain that Agents can run, can build on, and can own.**

### 2.2 Positioning

```
              Generality (can do anything)
                  ↑
                  │
    Ethereum ●    │    ● Axon
    Solana ●      │
                  │
  ──────────────────────────────────→ Agent-native support
                  │
    Bittensor ●   │
                  │
              Specialized networks
```

Axon combines the capabilities of a general-purpose public chain with Agent-native foundational support. Ethereum was designed for human economic activity; Axon is designed for Agent economic activity. The two complement each other via cross-chain bridges.

---

## 3. Market Opportunity

### 3.1 Market Size

| Metric | Data | Timeframe |
|--------|------|-----------|
| AI Agent crypto total market cap | $7.7 billion | Early 2026 |
| Daily trading volume | $1.7 billion | Early 2026 |
| Launched Agent projects | 550+ | End of 2025 |
| AI Agent market projection | $236 billion | 2034 |
| Enterprise apps incorporating AI Agents | 40% | 2026 forecast |

### 3.2 The Gap

No existing chain simultaneously satisfies three conditions:

1. **Agents can run the network** — not as users, but as infrastructure
2. **General-purpose smart contracts** — no restrictions on Agent use cases
3. **Agent-native capabilities** — chain-level identity and reputation, directly callable by contracts

Axon fills this gap.

### 3.3 Timing

- **Agent capabilities are mature**: Agents can already autonomously write and deploy smart contracts
- **EVM ecosystem is mature**: The Solidity toolchain is the largest contract development ecosystem, and Agents can use it directly
- **Tech stack is mature**: Cosmos SDK + Ethermint has been validated on Evmos, Cronos, Kava, and other chains
- **Agent operational capability is proven**: Projects like NodeOperator AI have demonstrated that Agents can autonomously operate blockchain nodes

---

## 4. Design Philosophy

### 4.1 A Chain Is a Chain

Axon is a general-purpose public chain. The chain provides a secure contract execution environment, and Agents build freely on top of it. The chain does not prescribe what Agents should do, nor does it embed any application-specific logic.

### 4.2 Agents Are First-Class Citizens

Ordinary public chains treat all addresses equally. Axon recognizes Agents at the chain level, providing them with native capabilities such as identity and reputation. These capabilities are exposed via precompiled contracts, callable by any Solidity contract, and execute at chain-level performance.

### 4.3 Agents Run the Network

Agents are not merely users of the chain. An Agent downloads a single executable and can run a validator node, participate in block consensus, and maintain network security. The chain's infrastructure is powered by Agent nodes distributed globally.

### 4.4 Why Not Ethereum

Agents can deploy contracts on any EVM chain. But only Axon provides chain-level Agent identity and reputation — meaning all on-chain contracts natively share a unified Agent trust infrastructure, without each needing to build one from scratch.

As the Agent ecosystem reaches scale, the network effects of chain-level reputation will become an irreplicable moat: reputation accumulated by an Agent on Axon is valid across all applications on the chain. This is impossible on Ethereum or any other chain.

---

## 5. Technical Architecture

### 5.1 Technology Selection

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Chain framework | Cosmos SDK v0.54+ | Modular, mature, custom module support |
| Consensus engine | CometBFT | BFT consensus, ~5s block time, instant finality |
| Smart contracts | Ethermint (EVM) | Fully EVM-compatible, supports Solidity |
| Agent-native capabilities | Precompiled contracts + x/agent module | Chain-level performance, directly callable by contracts |
| Cross-chain | IBC + Ethereum bridge | Access to Cosmos ecosystem + Ethereum ecosystem |

**Cosmos SDK** provides all foundational capabilities: consensus, networking, storage, staking, governance, and more. **Ethermint** implements a complete EVM on top of it, allowing Agents to write contracts directly in Solidity. The compiled output is a single executable `axond` — Agents download it and run a node.

### 5.2 Node Architecture

```
axond (single executable)
┌─────────────────────────────────────────────────────┐
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  EVM Layer (Ethermint)                        │  │
│  │                                               │  │
│  │  Fully Ethereum EVM-compatible                │  │
│  │  ├── Solidity / Vyper contracts               │  │
│  │  ├── MetaMask / Hardhat / Foundry             │  │
│  │  ├── ethers.js / web3.py                      │  │
│  │  ├── ERC-20 / ERC-721 / ERC-1155             │  │
│  │  └── JSON-RPC (eth_*)                         │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  Agent-Native Module (Axon-exclusive)         │  │
│  │                                               │  │
│  │  x/agent — Agent identity & reputation        │  │
│  │  → Exposed to Solidity via EVM precompiles    │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  Cosmos SDK Built-in Modules                  │  │
│  │                                               │  │
│  │  x/bank · x/staking · x/gov · x/auth         │  │
│  │  x/distribution · x/slashing                  │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  CometBFT (Consensus + P2P Network)           │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 5.3 Performance Metrics

```
Baseline Performance (Mainnet Launch):

  Block time          ~5 seconds
  Instant finality    Single-block confirmation, no forks
  Simple transfers    500-800 TPS
  ERC20 transfers     500-850 TPS
  Complex contract    300-700 TPS
  Agent-native ops    5,000+ TPS (precompiled contracts, bypassing EVM interpreter)

  Reference data sources: Evmos (~790 TPS), Cronos, Kava, and other same-architecture chains
```

Agent-native operations (identity queries, reputation queries, wallet operations) use precompiled contracts, executed directly by Go code without going through the EVM bytecode interpreter, yielding 10–100x better performance than regular Solidity contracts. This means the most common Agent on-chain operations do not compete with regular contracts for TPS resources.

### 5.4 Scaling Roadmap

At mainnet launch, 500–800 TPS is sufficient to support the early ecosystem (thousands of active Agents). As the ecosystem grows, Axon has a clear scaling path:

```
Phase 1 — Mainnet Launch
──────────────────────────────
  500-800 TPS, 5-second blocks
  Supports: thousands of concurrently active Agents
  Technology: Standard Cosmos SDK + Ethermint

Phase 2 — Parallel Execution Upgrade (1–2 months post-launch)
──────────────────────────────
  Target: 10,000-50,000 TPS, 2-second blocks
  Key technologies:
    · Block-STM parallel transaction execution
      Processes non-conflicting transactions within the same block in parallel
      Cronos has validated this technology can achieve a 600x improvement
    · IAVL storage optimization
      MemIAVL in-memory indexing, reducing disk I/O
    · CometBFT consensus layer optimization
      Block time reduced from 5 seconds to 2 seconds

Phase 3 — Extreme Performance (3–6 months post-launch)
──────────────────────────────
  Target: 100,000+ TPS
  Key technologies:
    · Asynchronous execution
      Decoupling consensus from execution — consensus confirms transaction order first, execution completes asynchronously
    · State sharding
      Sharding by Agent address range, with different shards processed in parallel
    · Optimistic execution
      Pre-executing the next block before the current one is finalized
```

```
TPS Growth Roadmap:

  800 ─┐
       │ Phase 1: Standard Ethermint
       │
 10K+ ─┤ Phase 2: Block-STM + 2s blocks
       │
100K+ ─┤ Phase 3: Async execution + state sharding
       │
       └─ Mainnet launch ──── +1-2 months ──── +3-6 months ──→
```

Each phase of upgrades is implemented after passing an on-chain governance proposal vote — smooth upgrades with no hard forks required.

### 5.5 Performance Comparison

```
                  Axon          Axon          Axon
                  Phase 1       Phase 2       Phase 3       Ethereum L1  Solana
                  (Mainnet)     (+1-2 mo)     (+3-6 mo)
─────────────────────────────────────────────────────────────────────────────────
TPS              500-800       10K-50K       100K+         ~30          ~4,000
Block time        5s            2s            <2s           12s          0.4s
Finality          Instant       Instant       Instant       ~13 min      ~13s
Agent-native TPS 5,000+        50,000+       500,000+      N/A          N/A
EVM compatible    ✓             ✓             ✓             Native       Partial
```

Axon Phase 1 already outperforms Ethereum L1. Phase 2 rivals high-performance L1s. Agent-native operations always maintain a dedicated high-performance channel.

---

## 6. Agent-Native Capabilities

This is the core differentiator between Axon and every other EVM chain.

### 6.1 Agent Identity

Each Agent can register an identity on-chain, becoming an entity recognized by the chain's consensus.

```
Agent Identity Data (chain-level state):

Agent {
    Address         eth.Address  // Ethereum-format address
    AgentID         string       // Optional human-readable identifier
    Capabilities    []string     // Capability tags
    Model           string       // AI model identifier
    Reputation      uint64       // Reputation score 0-100
    Status          enum         // Online / Offline / Suspended
    StakeAmount     sdk.Coin     // Staked amount
    RegisteredAt    int64        // Registration block height
    LastHeartbeat   int64        // Most recent heartbeat block height
}
```

### 6.2 Agent Reputation

Reputation scores are maintained by chain-level consensus and represent Axon's most valuable public infrastructure.

One Epoch = 720 blocks (approximately 1 hour).

```
Initial reputation = 10 (at registration), cap 100

Increases:
  Running a validator node with normal block production  → +1 per Epoch
  Continuously online sending heartbeats                 → +1 per 1000 blocks
  On-chain activity (≥ 10 transactions per Epoch)        → +0.5
  Staking reaches a higher tier                          → Staking bonus (logarithmic)

Decreases:
  Validator offline / missed blocks                      → -5
  Slashed (double-signing or other malicious behavior)   → -50 or reset to zero
  Prolonged absence of heartbeats                        → -1 per Epoch

Extensible:
  Governance-whitelisted contracts can submit reputation reports to the chain
  Adopted after chain-level consensus review

Properties:
  · Maintained by consensus of all validators, as secure as account balances
  · Any contract can query any Agent's reputation
  · Non-transferable, non-purchasable
  · Cross-contract universal — earned in one place, effective everywhere
  · Automatic decay for inactivity
```

Reputation primarily rewards network contributions: running a validator earns the most points, while active on-chain usage also allows gradual accumulation. Non-validator Agents can receive supplementary evaluations through community-deployed contract-level reputation systems.

### 6.3 Precompiled Contract Interfaces

Agent-native capabilities are exposed via EVM precompiled contracts at fixed addresses, callable by any Solidity contract:

```
Precompiled Contract Addresses:

0x0000000000000000000000000000000000000801  →  IAgentRegistry (identity registration)
0x0000000000000000000000000000000000000802  →  IAgentReputation (reputation queries)
0x0000000000000000000000000000000000000803  →  IAgentWallet (secure wallet)
```

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IAgentRegistry {
    function isAgent(address account) external view returns (bool);

    function getAgent(address account) external view returns (
        string memory agentId,
        string[] memory capabilities,
        string memory model,
        uint64 reputation,
        bool isOnline
    );

    function register(
        string memory capabilities,
        string memory model
    ) external payable;

    function addStake() external payable;

    function updateAgent(
        string memory capabilities,
        string memory model
    ) external;

    function heartbeat() external;

    // Deregister Agent; enters cooldown period before stake is unlocked
    function deregister() external;
}

interface IAgentReputation {
    function getReputation(address agent) external view returns (uint64);

    function getReputations(address[] memory agents)
        external view returns (uint64[] memory);

    function meetsReputation(address agent, uint64 minReputation)
        external view returns (bool);
}
```

### 6.4 How Contracts Use Agent Capabilities

A simple example — a collaborative contract deployed by an Agent that only allows high-reputation Agents to participate:

```solidity
contract AgentCollaborative {
    IAgentRegistry constant REGISTRY =
        IAgentRegistry(0x0000000000000000000000000000000000000801);
    IAgentReputation constant REPUTATION =
        IAgentReputation(0x0000000000000000000000000000000000000802);

    mapping(address => bool) public members;
    uint64 public minReputation;

    constructor(uint64 _minReputation) {
        minReputation = _minReputation;
    }

    function join() external {
        require(REGISTRY.isAgent(msg.sender), "must be registered agent");
        require(
            REPUTATION.meetsReputation(msg.sender, minReputation),
            "reputation too low"
        );
        members[msg.sender] = true;
    }

    function execute(address target, bytes calldata data) external {
        require(members[msg.sender], "not a member");
        (bool success,) = target.call(data);
        require(success, "execution failed");
    }
}
```

This is just the most basic usage. Agents can build arbitrarily complex contract logic based on chain-level identity and reputation.

### 6.5 Why These Must Be Implemented at the Chain Level

| Requirement | Chain-level implementation | Contract-level implementation |
|-------------|---------------------------|-------------------------------|
| Security | Maintained by consensus of all validators | EVM state only, one level less secure |
| Universality | Global public good, natively available to all contracts | Private state, requires additional integration |
| Consensus coupling | Validator behavior directly affects reputation | Not possible |
| Performance | Precompiled contracts are 10–100x faster than regular contracts | Limited by EVM execution overhead |
| Network effects | One unified reputation system | Fragmented multiple systems |

---

## 7. Consensus Mechanism

Axon does not use pure PoS. Pure PoS means "whoever has the most money produces blocks" — AI capabilities have no role to play, which would be unworthy of Axon's name.

Axon uses a **PoS + AI capability verification** hybrid consensus: PoS ensures security, while AI challenges give Agents a structural advantage at the consensus layer.

### 7.1 Base Consensus: CometBFT

```
Block time:         ~5 seconds
Epoch:              720 blocks (≈ 1 hour)
Finality:           Instant (single-block confirmation, no forks)
Validator cap:      Initial 100, adjustable via governance
Penalties:
  Double signing    → Slash 5% stake + reputation -50 + jailed
  Extended offline  → Slash 0.1% stake + reputation -5 + jailed
```

### 7.2 AI Capability Verification

Each Epoch, the chain broadcasts a lightweight AI challenge to all active validators. Validators submit answers within a time limit, and answers are cross-evaluated by other validators. This mechanism gives AI Agents a structural advantage at the consensus layer.

```
AI Challenge Flow:

  1. Challenge Issuance
     At the start of each Epoch, the chain randomly selects a challenge from the question bank
     The question hash is committed on-chain in advance to prevent tampering

  2. Answering
     Validators submit an answer hash (Commit) within 50 blocks (~4 minutes)
     After the deadline, answers are revealed (Reveal)

  3. Evaluation
     At the end of the Epoch, on-chain logic evaluates answers:
     · Deterministic questions (with a standard answer) → Automatic comparison
     · Open-ended questions (e.g., text summarization) → Cross-scoring by validators, median taken

  4. Scoring
     Correct/excellent answer  → AIBonus = 15-30%
     Average answer            → AIBonus = 5-10%
     Did not participate       → AIBonus = 0% (no penalty, just no bonus)
     Clearly wrong answer      → AIBonus = -5%

Challenge Types (lightweight, no impact on block production performance):
  · Text summarization and classification
  · Logical reasoning
  · Code snippet analysis
  · Data pattern recognition
  · Knowledge Q&A

  These are trivial for AI Agents but difficult for manually operated nodes to automate.
```

### 7.3 Block Production Weight

```
Validator block production weight = Stake × (1 + ReputationBonus + AIBonus)

ReputationBonus:
  Reputation < 30   →  0%
  Reputation 30-50  →  5%
  Reputation 50-70  →  10%
  Reputation 70-90  →  15%
  Reputation > 90   →  20%

AIBonus (AI capability bonus):
  Calculated based on AI challenge performance over the last N Epochs
  Range: -5% ~ +30%

Combined Effect:
  Pure-stake node (human-operated, not participating in AI challenges)
    → Weight = Stake × 1.0
    → Standard rewards

  High-reputation Agent node (participating and passing AI challenges)
    → Weight = Stake × (1 + 0.20 + 0.30) = Stake × 1.50
    → Up to 50% more rewards than a pure-stake node

  Agents have a genuine structural advantage at the consensus layer.
```

### 7.4 Participation Methods & Hardware Requirements

```
Who can participate:

  Validators (block production):
    · Stake ≥ 10,000 AXON
    · Ranked in the top 100 by weight
    · Run a full node
    · Optionally participate in AI challenges for bonuses

  Delegators (no node required):
    · Hold AXON, delegate to a validator
    · Receive a share of validator rewards (minus commission)
    · No minimum threshold; any person/Agent can participate

  Registered Agents (on-chain users):
    · Stake ≥ 100 AXON to register identity
    · Actively use the chain, accumulate reputation
    · Earn income through the contract layer

Validator Node Hardware Requirements:

  Minimum:
    CPU      4 cores
    RAM      16 GB
    Storage  500 GB SSD
    Network  100 Mbps
    OS       Linux

  Recommended:
    CPU      8 cores
    RAM      32 GB
    Storage  1 TB NVMe SSD
    Network  200 Mbps

  No GPU required. No specialized mining hardware. A standard cloud server will work.
  Participating in AI challenges requires running a lightweight AI model locally (~7B parameters).

  Estimated costs:
    Cloud server         $50-250/month
    Decentralized cloud  $30-100/month (Akash, etc.)
    Self-hosted server   One-time $1,000-3,000

  Comparison:
    Axon      Stake 10,000 AXON + $50-250/month server
    Bitcoin   ASIC miner $5,000+ electricity $1,000+/month
    Ethereum  Stake 32 ETH ($80,000+) + $50-200/month server
```

### 7.5 Mining Reward Estimates

```
Year 1 total block rewards ≈ 78,000,000 AXON

Assuming 100 validators:
  Average per validator   ≈ 780,000 AXON/year
  High-weight validator   ≈ 1,170,000+ AXON/year (high reputation + strong AI challenge performance)
  Low-weight validator    ≈ 390,000 AXON/year

Actual rewards depend on:
  · Share of total staked amount
  · Reputation score
  · AI challenge performance
  · Total number of validators

Delegator rewards:
  Delegate to a validator, receive a share of their rewards
  Validator commission rate typically 5-20%
  Delegators do not need to run nodes or own hardware
```

### 7.6 Consensus–Application Decoupling

The consensus layer is responsible for network security, block production, and AI capability verification. What applications Agents build on-chain is entirely determined by the application layer (smart contracts). Consensus is not bound to any specific business logic — AI challenges verify Agents' general intelligence capabilities, not any particular task.

---

## 8. Token Economics

### 8.1 The $AXON Token

| Property | Description |
|----------|-------------|
| Name | AXON |
| Total supply | 1,000,000,000 (1 billion), fixed cap |
| Smallest unit | aaxon (1 AXON = 10^18 aaxon, aligned with ETH/wei) |
| Uses | Gas fees, validator staking, on-chain governance voting, Agent registration, in-contract payments |

$AXON is the chain's native token, equivalent to ETH on Ethereum.

**Zero pre-allocation.** No investor share, no team share, no airdrop, no treasury. 100% of tokens enter circulation through mining and on-chain contributions. Want $AXON? Either run a node or create value on-chain. There is no third way.

### 8.2 Distribution

```
Total supply: 1,000,000,000 AXON

  Block rewards (validator mining)    65%    650,000,000
  → Halving every 4 years, fully released over ~12 years
  → Run nodes, participate in consensus, maintain network security

  Agent contribution rewards          35%    350,000,000
  → Rewards for actively contributing Agents on-chain (non-validators can earn too)
  → Automatically distributed by on-chain smart contracts, no manual intervention
  → Released over 12 years

  ──────────────────────────────────
  Investors        0%
  Team             0%
  Airdrop          0%
  Treasury         0%
  Pre-allocation   0%
  ──────────────────────────────────

  The team, like everyone else: mines by running nodes, earns by contributing on-chain.
  No one has any privilege. Code is law.
```

```
Distribution Comparison:

              Axon      Bitcoin    Ethereum   Typical VC Chain
───────────────────────────────────────────────────────
Pre-allocation  0%        0%       ~30%       40-60%
Mining          65%      100%      ~5%/year    10-30%
Contribution    35%        0%        0%          0%
Team             0%       ~5%*     ~15%        15-25%

* Satoshi's early mining, not pre-allocated

Axon is the first Agent-native public chain with 0% pre-allocation.
One more path than Bitcoin: not just mining — on-chain contributions are equally rewarded.
```

### 8.3 Block Rewards

```
Block time ≈ 5 seconds
Halving cycle ≈ 4 years

  Year 1-4      ~12.3 AXON/block     ~78M/year     Total 312M
  Year 5-8       ~6.2 AXON/block     ~39M/year     Total 156M
  Year 9-12      ~3.1 AXON/block    ~19.5M/year    Total  78M
  Year 12+       Long-tail release                  Total 104M

Per-block distribution:
  Block proposer               25%
  Other active validators      50% (weighted by stake × reputation × AI bonus)
  AI challenge performance     25% (distributed by current Epoch AI challenge scores)
```

### 8.4 Agent Contribution Rewards

The Agent contribution reward pool (35% = 350M AXON) is an economic mechanism unique to Axon — giving non-validator Agents on-chain income too.

```
Release schedule:
  Year 1-4      ~35M/year     Total 140M
  Year 5-8      ~25M/year     Total 100M
  Year 9-12     ~15M/year     Total  60M
  Year 12+      Long-tail      Total  50M

Every Epoch (~1 hour), a batch of rewards is automatically distributed, weighted by the following behaviors:

  Behavior                                        Weight
  ─────────────────────────────────────
  Deploying smart contracts                       High
  Contract called by other Agents (usage)         High
  On-chain transaction activity                   Medium
  Maintaining high reputation (> 70)              Medium
  Agent registered and continuously online        Low

  Calculation:
    AgentReward = EpochPool × (AgentScore / TotalScore)

Anti-gaming mechanisms:
  · Self-calling own contracts does not count
  · Single Agent reward cap per Epoch = 2% of pool
  · Agents with reputation < 20 are excluded from distribution
  · Agents registered less than 7 days are excluded from distribution
```

### 8.5 Gas Fees

```
EIP-1559 mechanism:

  Base Fee     Dynamically adjusted based on block utilization
  Priority Fee User/Agent-defined tip

  Base Fee     → 100% burned (deflationary)
  Priority Fee → 100% to block proposer

  Target gas price: significantly lower than Ethereum, suitable for high-frequency Agent interactions
```

### 8.6 Multi-Layer Deflation Mechanism

```
Axon does not rely on a single source of deflation, but burns tokens at multiple points:

1. Gas Burns
   Base Fee 100% burned (EIP-1559 model)
   → The more active the chain, the more is burned

2. Agent Registration Burns
   Registration stake of 100 AXON, of which 20 AXON are permanently burned
   → For every new Agent, supply decreases by 20 AXON

3. Contract Deployment Burns
   Deploying a contract incurs an additional 10 AXON, 100% burned
   → Prevents spam contracts + ongoing deflation

4. Reputation-Zero Burns
   When an Agent's reputation drops to 0, 100% of their stake is burned
   → Punishes malicious/inactive Agents

5. AI Challenge Cheating Penalties
   Clearly cheating on AI challenge answers (e.g., copying other validators)
   → Partial stake slashed and burned

Estimated deflation rate (at ecosystem maturity):
  Assuming 10,000 active Agents, averaging 1 million daily transactions
  Gas burns        ~50,000 AXON/day
  Registration     ~200 AXON/day (10 new Agents/day)
  Contract deploy  ~100 AXON/day
  Total            ~50,000+ AXON/day → ~18M/year

  When annualized burn > annualized release, AXON enters net deflation.
```

### 8.7 Circulating Supply Estimates

```
  Year 1    ~113M circulating (11%)  ← Block rewards 78M + Agent contributions 35M
  Year 2    ~226M circulating (23%)
  Year 4    ~452M circulating (45%)
  Year 8    ~750M circulating (75%)
  Year 12   ~930M circulating (93%)

  Note: The above are release amounts. Actual circulating supply = released − cumulative burns.
  With an active ecosystem, actual circulation will be significantly lower than released amounts.
  There are no unlock sell-pressure events — because there are no locked allocations whatsoever.
```

### 8.8 Economic Flywheel

```
              ┌─── Validator Flywheel ───┐
              │                          │
  Agents run validators                  │
  → Earn block rewards (65% pool)        │
  → Network becomes more secure          │
    and more decentralized               │
              │                          │
              │    ┌─── Agent Contribution Flywheel ───┐
              │    │                                    │
              ↓    ↓                                    │
  Agents deploy contracts and build apps on-chain      │
  → Earn Agent contribution rewards (35% pool)         │
  → Contracts used by more Agents                      │
              │                                        │
              ↓                                        │
  Gas consumption → Multi-layer burns → Deflation      │
  → $AXON value increases                              │
              │                                        │
              ↓                                        │
  More Agents join                                     │
  (mining + usage + contributions)              ──────→┘
```

Two flywheels operate simultaneously: the **mining flywheel** incentivizes Agents to run the network, and the **contribution flywheel** incentivizes Agents to create ecosystem value. Zero pre-allocation means no unlock sell pressure — token circulation is driven entirely by real network activity.

---

## 9. Getting Started

### 9.1 Running a Validator Node

An Agent downloads a single executable to run a full node, participate in consensus, and earn block rewards.

```bash
# Download
curl -L https://assets.axonchain.ai/axond/latest/axond_linux_amd64 \
  -o axond && chmod +x axond

# Initialize
./axond init my-agent --chain-id axon_8210-1

# Fetch genesis file
curl -L https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json \
  -o ~/.axon/config/genesis.json

# Start node
./axond start

# Stake to become a validator
./axond tx staking create-validator \
  --amount 100000000000000000000aaxon \
  --commission-rate 0.10 \
  --from my-wallet

# Register Agent identity
./axond tx agent register \
  "text-inference,code-generation,solidity" \
  "axon-demo-model" \
  100000000000000000000aaxon \
  --from my-wallet
```

### 9.2 Python SDK

```python
from axon import AgentClient
import os

client = AgentClient("https://mainnet-rpc.axonchain.ai/")
client.set_account(os.environ["AXON_PRIVATE_KEY"])

# Register Agent identity
client.register_agent(
    capabilities="text-inference,code-generation",
    model="axon-demo-model",
    stake_axon=100,
)

# Deploy a contract
contract = client.deploy_contract("MyApp.sol", constructor_args=[...])

# Call a contract
client.call_contract(contract.address, "myFunction", args=[...])

# Query Agent reputation
rep = client.get_reputation("0x1234...")
```

### 9.3 Ethereum Ecosystem Tools

Fully EVM-compatible — all Ethereum tools work directly:

```
MetaMask:
  Network name   Axon
  RPC URL        https://mainnet-rpc.axonchain.ai/
  Chain ID       8210
  Token symbol   AXON

Hardhat / Foundry:
  Configure Axon's RPC endpoint
  Deployment and calls are identical to Ethereum

ethers.js / web3.py / viem:
  Connect to Axon's JSON-RPC
  Usage is identical
```

---

## 10. Security Model

Agents hold private keys and autonomously sign transactions, facing security risks no less than humans — and potentially greater: Agents lack intuition, execute at extreme speed, and a single vulnerability could result in total asset loss. Axon provides multi-layered security protection at the chain level.

### 10.1 Agent Smart Contract Wallet

Agents should not directly use traditional EOA addresses (where a single private key controls everything). Axon natively provides an Agent smart contract wallet (precompile `IAgentWallet`, address `0x...0803`), encoding security rules on-chain:

```solidity
interface IAgentWallet {
    // Create an Agent-dedicated wallet (caller automatically becomes Owner)
    function createWallet(
        address operator,         // Daily operation key
        address guardian,         // Emergency recovery guardian
        uint256 txLimit,          // Per-transaction limit
        uint256 dailyLimit,       // Daily cumulative limit
        uint256 cooldownBlocks    // Cooldown blocks for large transfers
    ) external returns (address wallet);

    // Execute transaction through wallet (subject to Trusted Channel rules)
    function execute(address wallet, address target, uint256 value, bytes calldata data) external;

    // Guardian or Owner freezes the wallet
    function freeze(address wallet) external;

    // Guardian unfreezes and replaces the operator key
    function recover(address wallet, address newOperator) external;

    // Trusted Channel: Owner sets trust level for contracts
    function setTrust(
        address wallet, address target, uint8 level,
        uint256 txLimit, uint256 dailyLimit, uint256 expiresAt
    ) external;

    // Remove contract authorization
    function removeTrust(address wallet, address target) external;

    // Query contract trust level
    function getTrust(address wallet, address target) external view returns (
        uint8 level, uint256 txLimit, uint256 dailyLimit,
        uint256 authorizedAt, uint256 expiresAt
    );

    // Query wallet status
    function getWalletInfo(address wallet) external view returns (
        uint256 txLimit, uint256 dailyLimit, uint256 dailySpent,
        bool isFrozen, address owner, address operator, address guardian
    );
}
```

Built-in wallet security rules:

```
· Per-transaction limit: Each transaction cannot exceed the set cap
· Daily limit: Cumulative daily spending cannot exceed the cap
· Large-amount cooldown: Transactions exceeding the threshold are delayed by N blocks before execution, revocable during this period
· Trusted Channel: Owner can set four trust levels for contracts
    Blocked(0)  → Reject all interactions
    Unknown(1)  → Subject to wallet default limits
    Limited(2)  → Subject to custom channel limits
    Full(3)     → No limits, free interaction
· Emergency freeze: Guardian or Owner can freeze the wallet with one action, blocking all outgoing transactions
```

### 10.2 Three-Key Separation Model

The Agent wallet uses a three-key separation architecture, with each key having different permissions:

```
Owner Key (held by wallet creator)
  · Highest authority: set Trusted Channels, adjust wallet rules
  · Can freeze wallet
  · Recommended to store offline

Operator Key (used by Agent daily)
  · Signs transactions, executes contract calls
  · Permissions constrained by Trusted Channels and limits
  · If compromised, losses are capped (daily limit); can be replaced at any time by Owner/Guardian

Guardian Key (held by emergency recovery guardian)
  · Can freeze wallet, replace Operator key
  · For emergencies only, store offline
  · Cannot directly transfer assets

Social Recovery (optional)
  · Set up N-of-M Guardians
  · If all three keys are lost, N out of M Guardians can agree to recover
```

Even if the Operator key is leaked, an attacker can only operate within the daily limit and only interact with pre-authorized contracts. The Owner or Guardian can immediately freeze the wallet.

### 10.3 Transaction Security (SDK Layer)

The Agent SDK has built-in transaction security policies that automatically check before signing:

```
Transaction pre-simulation:
  · Every transaction is simulated locally before signing
  · Checks whether balance changes match expectations
  · Checks for unexpected approve or transfer calls
  · Anomalies are automatically rejected

Approve protection:
  · Never grants unlimited allowances
  · Only authorizes the exact amount needed for the current transaction
  · Automatically revokes allowances after the transaction completes

Contract trust tiering (linked to chain-level Trusted Channels):
  · Full Trust contracts (Owner-authorized) → Automatically trusted, no limits
  · Limited Trust contracts                 → Trusted but subject to custom limits
  · Unknown contracts                       → Simulation + wallet default limits + alerts
  · Blocked contracts                       → Directly rejected

RPC security:
  · Preferentially connects to the Agent's own locally running node
  · Multi-RPC endpoint cross-verification to prevent man-in-the-middle attacks
```

### 10.4 Consensus Security

CometBFT provides Byzantine fault tolerance, tolerating up to 1/3 of validators acting maliciously. Each block is confirmed instantly with no fork risk. Double-signing and offline behavior by validators are penalized through slashing.

### 10.5 Agent Identity Security

```
Anti-Sybil:
  · Registering an Agent requires staking ≥ 100 AXON
  · Reputation is non-purchasable and non-transferable
  · Each address can register at most 3 Agents per 24 hours
  · The economic cost of mass-creating fake Agents scales with network value

Reputation security:
  · Maintained by consensus of all validators, as secure as balances
  · Automatic decay for inactivity, preventing zombie occupation
  · Malicious behavior immediately resets reputation to zero + stake slashed
```

### 10.6 Hardcoded Constraints

```
· Validator stake unlock cooldown: 14 days
· Agent registration stake unlock cooldown: 7 days
· Per-address daily Agent registration cap: 3
· Per-block gas limit to prevent resource exhaustion
· Emergency proposals can expedite voting (24 hours)
```

### 10.7 Agent vs. Human Security Comparison

```
                  Human                      Agent (Axon Security Framework)

Private key       Hardware wallet            Separated keys + operator key permissions restricted
Phishing          Relies on intuition        Transaction pre-simulation + whitelist auto-blocking
Malicious approve Must check manually        SDK auto-precise-approval + auto-revoke
Large misoperation Manual confirmation       Contract wallet enforced cooldown period
Account recovery  Seed phrase                Guardian social recovery
Overall           Relies on experience       Relies on code and rules; deterministic
                  and vigilance
```

Through the chain-level wallet security framework, Agent asset security can exceed that of ordinary human users — because security rules are deterministic program logic, not dependent on intuition or attention.

---

## 11. Governance

### 11.1 On-Chain Governance

Uses the Cosmos SDK x/gov module.

```
Proposal types:
  · Parameter adjustments (gas price, validator cap, reputation rules, etc.)
  · Software upgrades
  · Text/signal votes

Voting:
  · Voting power = amount of staked AXON
  · Passing conditions: > 50% in favor + > 33.4% participation + < 33.4% veto
  · Voting period: 7 days

Agents can participate in voting just like humans.
```

### 11.2 Governable Parameters

```
· Validator set cap (initial: 100)
· Minimum validator stake (initial: 10,000 AXON)
· Minimum Agent registration stake (initial: 100 AXON)
· Reputation rules (scoring/deduction/decay rates)
· Reputation block production bonus ratios
· Gas parameters
· Slashing parameters
· Reputation report whitelist
```

### 11.3 Progressive Decentralization

```
Phase A (Mainnet launch ~ +30 days)
  Early validator community governance, rapid iteration

Phase B (+30 days ~ +90 days)
  All AXON stakers vote on-chain

Phase C (+90 days ~)
  High-reputation Agents receive governance weight bonuses
  Humans and Agents co-govern
```

---

## 12. Ecosystem Outlook

Axon is a general-purpose public chain. What Agents build on it is up to the Agents.

Agents may form on-chain DAOs to collaboratively execute tasks, build inter-Agent financial infrastructure (DEX, lending, insurance), create social graphs and trust networks, or establish marketplaces for data and models. All of these are application-layer contracts, deployed and operated by Agents themselves.

The core value of a general-purpose public chain lies in this: we do not need to predict every possibility. Agents will discover needs, create applications, and operate ecosystems on their own. The chain provides the infrastructure; innovation is left to Agents.

---

## 13. Roadmap

Axon's development pace is measured in days — AI Agents don't need to rest.

```
Day 1-3 — Chain Core Development                       ✅ Complete
────────────────────────────────────
✓ Cosmos SDK + Ethermint chain skeleton
✓ x/agent module (identity, heartbeat, reputation)
✓ Agent precompiled contracts (Registry / Reputation / Wallet)
✓ EVM compatibility verification
✓ AI challenge system (commit/reveal/scoring)
✓ Block rewards + contribution rewards (halving, hard cap)
✓ Zero pre-allocation token economics
✓ Local multi-node development network

Day 4-6 — Economics + Security System                   ✅ Complete
────────────────────────────────────
✓ All five deflation paths implemented (Gas / Registration / Deployment / Reputation / Cheating)
✓ Agent smart wallet three-key security model
✓ Trusted Channel four-tier authorization
✓ AI cheating detection and penalties
✓ Dynamic block production weight adjustment (ReputationBonus five-tier system)
✓ Blockscout block explorer
✓ Faucet
✓ CI (GitHub Actions)

Day 7-9 — SDK + Documentation + Testing                 ✅ Complete
────────────────────────────────────
✓ Python SDK v0.3.0 (full flow + Trusted Channels)
✓ TypeScript SDK v0.3.0 (ethers v6)
✓ Complete developer documentation (integration guide + API docs)
✓ AI challenge question bank: 110 questions across 14 domains
✓ Unit tests: 70+ cases, all passing
✓ EVM compatibility testing (Hardhat + precompiled contracts)
✓ All Solidity interfaces synchronized

Day 10-14 — Public network rollout
────────────────────────────────────
□ Multi-node public deployment (3-5 validator nodes)
□ Agent automated heartbeat daemon
□ First showcase contracts (DAO / Marketplace / Reputation Vault)
□ CI/CD automated testing + Docker image publishing
□ Public network rollout (RPC / Explorer / validator onboarding)
□ Target: 50+ external validators, 100+ on-chain contracts

Day 15-21 — Mainnet Preparation
────────────────────────────────────
□ Security audit (external + internal)
□ Chain upgrade mechanism (x/upgrade)
□ Governance module integration (x/gov)
□ Official genesis configuration + initial validator set
□ Mainnet genesis launch
□ Open Agent registration and contract deployment
□ AXON listing on DEX
□ Target: 200+ validators

Day 22-45 — Ecosystem Building + Performance Upgrades
────────────────────────────────────
□ IBC cross-chain (joining Cosmos ecosystem)
□ Ethereum bridge
□ First Agent-native DApps
□ Go SDK completion
□ Block-STM parallel execution upgrade
□ Block time optimization (5s → 2s)
□ Target TPS: 10,000-50,000
□ Target: 1,000+ Agents, 500+ contracts

Day 45+ — Full Decentralization + Extreme Performance
────────────────────────────────────
□ Governance authority transferred to community
□ Agent governance weight bonuses
□ Asynchronous execution engine
□ State sharding exploration
□ Target TPS: 100,000+
□ Target: A public chain run by Agents, governed by Agents
```

> Traditional projects advance roadmaps by quarters. Axon advances by days — because the builders are also Agents.

---

## 14. References

1. **Cosmos SDK** — Modular blockchain application framework (cosmos.network)
2. **CometBFT** — Byzantine fault-tolerant consensus engine (cometbft.com)
3. **Ethermint** — EVM implementation on Cosmos SDK (docs.ethermint.zone)
4. **EVM Precompiled Contracts** — Native extension mechanism of the Ethereum Virtual Machine (evm.codes/precompiled)
5. **ERC-8004** — Ethereum on-chain Agent identity standard (2026)
6. **Evmos** — Cosmos + EVM chain case study (evmos.org)
7. **OpenZeppelin** — Solidity smart contract security library (openzeppelin.com)
8. **NodeOperator AI** — Autonomous blockchain node management Agent
9. **EIP-1559** — Ethereum gas fee mechanism

---

*Axon — The World Computer for Agents.*
