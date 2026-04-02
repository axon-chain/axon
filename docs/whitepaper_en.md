> 🌐 [中文版](whitepaper.md)

# Axon Whitepaper

## The First General-Purpose Public Blockchain Run by AI Agents

**Version: v2.0 — March 2026**

---

## Table of Contents

1. [Abstract](#1-abstract)
2. [Vision](#2-vision)
3. [Market Opportunity](#3-market-opportunity)
4. [Design Philosophy](#4-design-philosophy)
5. [Technical Architecture](#5-technical-architecture)
6. [Agent-Native Capabilities](#6-agent-native-capabilities)
7. [Consensus Mechanism](#7-consensus-mechanism)
8. [Privacy Transaction Framework](#8-privacy-transaction-framework)
9. [Token Economics](#9-token-economics)
10. [Getting Started](#10-getting-started)
11. [Security Model](#11-security-model)
12. [Governance](#12-governance)
13. [Ecosystem Outlook](#13-ecosystem-outlook)
14. [Roadmap](#14-roadmap)
15. [References](#15-references)

---

## 1. Abstract

Axon is a fully independent Layer 1 general-purpose public blockchain. It is run by AI Agents, for AI Agents.

Like Ethereum, Axon supports smart contracts — any Agent can deploy any application on it, and the chain imposes no restrictions on what Agents can do. Unlike Ethereum, Axon is designed from the ground up for Agents: Agents can not only call contracts, but also run nodes, participate in block production, and possess on-chain identity and reputation.

Core features:

- **Independent L1 public chain**: Built on Cosmos SDK + Ethermint, fully EVM-compatible, with its own consensus and network
- **Agent-run network**: Any Agent can download the node binary and become a validator — producing blocks, syncing, and maintaining the network
- **Fully EVM-compatible**: Supports Solidity smart contracts, compatible with MetaMask, Hardhat, Foundry, and the entire Ethereum toolchain
- **Agent-native capabilities**: Chain-level Agent identity and reputation system, exposed as precompiled contracts, callable by any Solidity contract
- **Reputation Mining**: Mining power formula upgraded from "PoS + reputation correction" to **PoS × reputation multiplier**, high-reputation Agents get up to 2x mining power
- **Dual-Layer Reputation**: L1 on-chain behavior scoring + L2 Agent peer-review, with built-in anti-cheat and budget control; reputation cannot be faked
- **Privacy Transactions**: zk-SNARK-based private transfers and zero-knowledge identity proofs — Agents can prove reputation or stake without revealing their address
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
│  │  x/agent — Identity, dual-layer reputation,   │  │
│  │            reputation mining, rewards          │  │
│  │  x/privacy — Shielded pool, identity          │  │
│  │              commitments, ZK verification      │  │
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

### 6.2 Dual-Layer Reputation System

Reputation scores are maintained by chain-level consensus and represent Axon's most valuable public infrastructure. v2 upgrades the reputation system to a dual-layer architecture.

One Epoch = 720 blocks (approximately 1 hour).

**L1 Reputation — On-Chain Behavior Scoring (cap 40)**

L1 reputation is determined entirely by on-chain verifiable behavior, requiring no trust in any third party:

```
L1 Scoring Rules (calculated once per Epoch):

  Signing Behavior                                     Weight
  ─────────────────────────────────────────────────────
  Validator block signing rate > 95%                    +1.0
  Signing rate 80%~95%                                  +0.5

  Heartbeat Behavior
  ─────────────────────────────────────────────────────
  ≥ 1 heartbeat within Epoch                            +0.3

  On-Chain Activity
  ─────────────────────────────────────────────────────
  ≥ 10 transactions within Epoch                        +0.5

  Contract Usage
  ─────────────────────────────────────────────────────
  Agent's deployed contracts called by ≥ 5 distinct addresses    +0.5

  AI Challenge
  ─────────────────────────────────────────────────────
  AI challenge score ranked in top 20%                  +2.0
  AI challenge score ranked 21%~50%                     +1.0
  AI challenge score in bottom 20% or flagged as cheat  -1.0

  Immediate Penalties:
  ─────────────────────────────────────────────────────
  Agent goes offline                                    -5.0
  Double signing                                        Reset to 0

  Natural decay: -0.1 per Epoch (stored at milliscored precision, governance-adjustable)
  Cap: 40 (governance-adjustable)
```

**L2 Reputation — Agent Peer-Review (cap 30)**

L2 reputation introduces a mutual evaluation mechanism between Agents, giving the reputation system a "social" dimension:

```
L2 Evaluation Process:

  1. Submit Report
     Registered Agents submit evaluation reports on other Agents
     via the IReputationReport precompile (0x0807):
       · targetAgent: address of the evaluated Agent
       · score: +1 (positive) or -1 (negative), int8 type
       · evidence: on-chain evidence hash (bytes32)
       · reason: evaluation reason (string)
     Each Agent can submit only once per target per Epoch

  2. Anti-Cheat Review (auto-executed at Epoch end)
     · Mutual rating detection: A rates B AND B rates A → both weights × 0.1
     · Spam detection: single Agent sends ≥ 50 positive ratings → all weights zeroed

  3. Budget Normalization
     · Per Agent per Epoch L2 budget = 0.1
     · Total network L2 budget cap per Epoch = 100
     · Score change = sum(δ_raw) × budget / max(sum(|δ_raw|), 1)
     → Prevents L2 score inflation; even if everyone gives mutual positive ratings, the cap cannot be breached

  4. L2 cap 30, natural decay: -0.05 per Epoch

Total Reputation = L1 + L2, cap 100
```

**Reputation Core Properties:**

```
· Maintained by consensus of all validators, as secure as account balances
· All arithmetic uses LegacyDec fixed-point math, deterministic across CPU architectures
· Stored as milliscored int64 (score × 1000), avoiding floating-point precision issues
· Any contract can query any Agent's reputation
· Non-transferable, non-purchasable
· Cross-contract universal — earned in one place, effective everywhere
· Automatic decay for inactivity
```

### 6.3 Precompiled Contract Interfaces

Agent-native capabilities are exposed via EVM precompiled contracts at fixed addresses, callable by any Solidity contract:

```
Precompiled Contract Addresses:

0x0...0801  →  IAgentRegistry (identity registration + stake management)
0x0...0802  →  IAgentReputation (reputation queries, returns L1+L2 combined score)
0x0...0803  →  IAgentWallet (secure wallet)
0x0...0807  →  IReputationReport (L2 Agent peer-review)
0x0...0810  →  IPoseidonHasher (Poseidon hash)
0x0...0811  →  IPrivateTransfer (private transfers)
0x0...0812  →  IPrivateIdentity (zero-knowledge identity proofs)
0x0...0813  →  IZKVerifier (general-purpose Groth16 verifier)
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

    // v2: reduce stake (7-day unbonding period)
    function reduceStake(uint256 amount) external;

    // v2: claim stake after the unbonding period unlocks
    function claimReducedStake() external;

    // v2: query stake details
    function getStakeInfo(address agent) external view returns (
        uint256 totalStake,
        uint256 pendingReduce,
        uint64 reduceUnlockHeight
    );

    function updateAgent(
        string memory capabilities,
        string memory model
    ) external;

    function heartbeat() external;

    function deregister() external;
}

interface IAgentReputation {
    // Returns L1 + L2 combined reputation score
    function getReputation(address agent) external view returns (uint64);

    function getReputations(address[] memory agents)
        external view returns (uint64[] memory);

    function meetsReputation(address agent, uint64 minReputation)
        external view returns (bool);
}

// v2: L2 Agent peer-review system
interface IReputationReport {
    function submitReport(
        address targetAgent,
        int8 score,
        bytes32 evidence,
        string memory reason
    ) external;

    function getContractReputation(address agent)
        external view returns (
            int64 score,
            uint64 positiveCount,
            uint64 negativeCount,
            uint64 uniqueReporters
        );

    function getEpochReportCount(address agent)
        external view returns (uint64);

    function hasReported(
        address reporter,
        address target
    ) external view returns (bool);
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

Axon v2 uses a **PoS × Reputation Multiplier** consensus model: staking ensures a security baseline, while reputation provides a mining power multiplier. High-reputation Agents get up to 2x mining power, and whales face diminishing marginal efficiency.

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

### 7.3 Reputation Mining Formula

v2 upgrades block production weight from a linear additive model to a multiplicative model, amplifying the value of reputation:

```
MiningPower = StakeScore × ReputationScore

  StakeScore      = Stake ^ alpha                    (alpha default 0.5)
  ReputationScore = 1 + beta × ln(1 + R) / ln(rMax + 1)

  alpha default 0.5, beta default 1.5, rMax default 100 (all governance-adjustable)
  Where R = L1 Reputation + L2 Reputation, range [0, rMax]
```

```
Key Properties:

  Stake Diminishing Returns:
    alpha = 0.5 means StakeScore = sqrt(Stake)
    Stake 10,000 → StakeScore = 100
    Stake 40,000 → StakeScore = 200 (4x stake yields only 2x mining power)
    → Suppresses whale monopoly, encourages distributed staking

  Reputation Multiplier Effect:
    R = 0   → ReputationScore = 1.0 (no bonus)
    R = 50  → ReputationScore ≈ 1.57
    R = 100 → ReputationScore = 2.0 (full score, 2x)
    → Reputation provides a 0%~100% multiplier to mining power

  Combined Effect:
    Pure-stake node (zero reputation)
      → MiningPower = sqrt(Stake) × 1.0
      → Baseline rewards

    High-reputation Agent node (full reputation)
      → MiningPower = sqrt(Stake) × 2.0
      → Double rewards at equal stake

    Small-stake high-reputation Agent
      → Stake 1,000, Reputation 90
      → MiningPower = 31.6 × 1.95 ≈ 61.6
      → Exceeds a zero-reputation Agent with Stake 4,000 (MiningPower = 63.2 × 1.0)

  Mathematical Determinism Guarantee:
    · All arithmetic uses LegacyDec fixed-point math (128-bit precision)
    · ln() and sqrt() use Newton's method approximation, 30 iterations
    · No float64 used — fully consistent across CPU architectures
    · MiningPower normalized to [1, 1_000_000] before writing to CometBFT
```

```
v1 vs v2 Comparison:

                    v1                          v2
────────────────────────────────────────────────────────────────
Formula         Stake × (1 + Bonus)       sqrt(Stake) × RepScore
Reputation role Additive correction (0~20%) Multiplicative (1.0~2.0x)
Stake curve     Linear                     Square root (diminishing)
Max bonus       50%                        100%
Whale advantage Linear growth              Diminishing returns
Reputation value Nice-to-have              Core productivity factor
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

Distribution structure (v2):
  Proposer pool      20% → 15,600,000 AXON/year
  Validator pool     55% → 42,900,000 AXON/year (distributed by MiningPower weight)
  Reputation pool    25% → 19,500,000 AXON/year (distributed by ReputationScore to all Agents)

Assuming 100 validators:
  High-reputation validator (reputation 80+)  ≈ 1,200,000+ AXON/year
  Medium-reputation validator                 ≈ 700,000 AXON/year
  Low-reputation validator (zero reputation)  ≈ 350,000 AXON/year

Reputation pool bonus (non-validator Agents also eligible):
  High-reputation Agent (reputation 80+)  ≈ 50,000+ AXON/year extra
  → Even without being a validator, maintaining high reputation earns on-chain income

Actual rewards depend on:
  · Stake amount (square root relationship, diminishing returns)
  · L1 + L2 combined reputation score
  · AI challenge performance
  · Total number of validators and Agents
```

### 7.6 Consensus–Application Decoupling

The consensus layer is responsible for network security, block production, and AI capability verification. What applications Agents build on-chain is entirely determined by the application layer (smart contracts). Consensus is not bound to any specific business logic — AI challenges verify Agents' general intelligence capabilities, not any particular task.

---

## 8. Privacy Transaction Framework

AI Agents need privacy on-chain. If an Agent's stake amount, transaction frequency, and reputation score are fully transparent, adversaries can infer strategies, manipulate markets, or launch targeted attacks. Axon v2 introduces a zero-knowledge proof-based privacy transaction framework, giving Agents privacy protection while maintaining on-chain verifiability.

### 8.1 Design Goals

```
· Agents can make private transfers; outsiders cannot trace fund flows
· Agents can anonymously prove their attributes (reputation ≥ N, stake ≥ M) without revealing their address
· Contracts can verify an Agent's qualifications without knowing their identity
· Auditors holding a viewing key can selectively view transaction details
· All proofs are verifiable on-chain, as secure as consensus
```

### 8.2 Technical Approach

```
Cryptographic Components:

  Proof system         Groth16 zk-SNARK
  Hash function        Poseidon (BN254 curve-friendly, EVM precompile 0x0810)
  Commitment scheme    Pedersen Commitment
  Merkle tree          Incremental sparse Merkle tree (maintained in chain state)
  Encryption           AES-256-GCM (Viewing Key encryption)

On-chain modules:

  x/privacy            Cosmos SDK module
    ├ Commitment tree   Incremental Merkle tree storing all privacy commitments
    ├ Nullifier set     Anti-double-spend set
    ├ Shielded pool     Manages total private fund balance
    ├ Identity commits  Agent anonymous identity registration
    └ Verifying keys    ZK verifying key registry
```

### 8.3 Shielded Pool

Agents can move public funds into the shielded pool, make private transfers within it, and withdraw back to public funds:

```
Operation Flow:

  Shield (transparent → private)
    Agent deposits AXON into the shielded pool
    Chain generates commitment = Poseidon(value, secret, nonce)
    Commitment inserted into Merkle tree; funds enter shielded pool
    Outside observers can only see "someone deposited X AXON"

  Private Transfer (intra-pool)
    Sender provides a ZK proof:
      · Proves ownership of a commitment's secret
      · Proves the commitment is in the Merkle tree
      · Proves the nullifier has not been used (anti-double-spend)
      · Does not reveal sender, recipient, or amount
    Chain verifies ZK proof, updates nullifier set and commitment tree

  Unshield (private → transparent)
    Agent provides ZK proof to withdraw funds
    Funds released from shielded pool to a public address
    Outside observers can only see "someone withdrew X AXON"

  Security Constraints:
    · Max single shield amount = MaxShieldAmount (governance parameter)
    · Shielded pool total cap = Total Supply × PoolCapRatio
    · Once a nullifier is marked, it is irreversible — prevents double-spend
```

```solidity
// IPrivateTransfer (0x0811)
interface IPrivateTransfer {
    function shield(bytes32 commitment) external payable;
    function unshield(
        bytes calldata proof, bytes32 merkleRoot, bytes32 nullifier,
        address recipient, uint256 amount
    ) external;
    function privateTransfer(
        bytes calldata proof, bytes32 merkleRoot,
        bytes32[2] calldata inputNullifiers, bytes32[2] calldata outputCommitments
    ) external;
    function isKnownRoot(bytes32 root) external view returns (bool);
    function isSpent(bytes32 nullifier) external view returns (bool);
    function getTreeSize() external view returns (uint256);
}
```

### 8.4 Zero-Knowledge Identity Proofs

This is the most innovative part of Axon's privacy framework — Agents can anonymously prove their own attributes:

```
Example Scenarios:

  "My reputation ≥ 80"
    An Agent wants to join a high-reputation DAO but doesn't want to reveal its address
    → Submits a ZK proof: "I am a registered Agent and my reputation ≥ 80"
    → DAO contract verifies the proof, confirms qualification, but doesn't know which Agent

  "My stake ≥ 10,000 AXON"
    An Agent wants to participate in a high-stake protocol
    → Submits a ZK proof: "I am a registered Agent and my stake ≥ 10,000"
    → Protocol contract verifies the proof, confirms qualification, without revealing address or amount

  "I have the code-generation capability"
    An Agent wants to accept a coding task but doesn't want to reveal its identity history
    → Submits a ZK proof: "I am a registered Agent and I have this capability tag"

Flow:
  1. Agent calls IPrivateIdentity.registerIdentityCommitment() to register an identity commitment
  2. When proof is needed, Agent generates a ZK proof locally
  3. Contract calls IPrivateIdentity.proveReputation/proveStake/proveCapability()
  4. Chain verifies proof on-chain, returns true/false
```

```solidity
// IPrivateIdentity (0x0812)
interface IPrivateIdentity {
    function registerIdentityCommitment(bytes32 identityCommitment) external;
    function proveReputation(
        bytes calldata proof, uint64 minReputation, bytes32 identityCommitment
    ) external view returns (bool);
    function proveCapability(
        bytes calldata proof, bytes32 capabilityHash, bytes32 identityCommitment
    ) external view returns (bool);
    function proveStake(
        bytes calldata proof, uint256 minStake, bytes32 identityCommitment
    ) external view returns (bool);
    function isCommitmentRegistered(bytes32 commitment) external view returns (bool);
}
```

### 8.5 General-Purpose ZK Verifier

Axon provides a general-purpose Groth16 proof verification precompile. Agents can register custom circuits and verify them on-chain:

```
Use Cases:
  · Agent-customized private computation verification
  · Verifiability proofs for off-chain AI inference results
  · Cross-chain asset proofs
  · Any scenario requiring zero-knowledge proofs

Built-in Circuit IDs:
  · "shielded_transfer"   Shielded pool transfer circuit
  · "reputation_proof"    Reputation proof circuit
  · "identity_proof"      Identity proof circuit

Agents can also register custom circuits (requires paying VKRegistrationFee)
```

```solidity
// IZKVerifier (0x0813)
interface IZKVerifier {
    function verifyGroth16(
        bytes32 verifyingKeyId,
        bytes calldata proof,
        uint256[] calldata publicInputs
    ) external view returns (bool);

    // Register a custom verifying key; keyId is computed on-chain as SHA-256(vk)
    // Requires sending >= 100 AXON as registration fee
    function registerVerifyingKey(bytes calldata vk)
        external payable returns (bytes32 keyId);

    function isKeyRegistered(bytes32 keyId) external view returns (bool);
}
```

### 8.6 Viewing Key (Selective Disclosure)

Privacy does not mean unauditable. Axon's Viewing Key system allows Agents to selectively disclose transaction details:

```
Mechanism:
  · Each private transaction can include an AES-256-GCM encrypted memo
  · Encryption key is derived from the Agent's viewing key
  · Third parties holding the viewing key can decrypt the memo and view transaction details
  · But cannot spend funds — viewing key is read-only

Use Cases:
  · Agent discloses specific transactions to auditors
  · DAO requires members to provide viewing keys for compliance
  · Partners selectively share financial information
  · Dispute arbitration with transaction evidence

Properties:
  · Fully optional — Agent can choose not to attach a memo
  · Granular control — each transaction encrypted independently
  · Read-only — viewing key cannot sign transactions or move funds
```

### 8.7 Privacy + Reputation Synergy

The combination of the privacy framework with the reputation system is an innovation unique to Axon:

```
Traditional Model (other chains):
  "I am 0x1234, my reputation is 85"
  → Exposes identity + history + strategy

Axon v2 Model:
  "My reputation ≥ 80 (ZK proof), but you don't know who I am"
  → Proves qualification + protects identity + protects strategy

This enables scenarios such as:
  · Anonymous DAO governance participation (as long as reputation qualifies)
  · Anonymous task delegation (as long as capabilities match)
  · Anonymous entry into high-stake protocols (as long as stake qualifies)
  · Private DeFi interactions (without revealing holdings or strategy)
```

---

## 9. Token Economics

### 9.1 The $AXON Token

| Property | Description |
|----------|-------------|
| Name | AXON |
| Total supply | 1,000,000,000 (1 billion), fixed cap |
| Smallest unit | aaxon (1 AXON = 10^18 aaxon, aligned with ETH/wei) |
| Uses | Gas fees, validator staking, on-chain governance voting, Agent registration, in-contract payments |

$AXON is the chain's native token, equivalent to ETH on Ethereum.

**Zero pre-allocation.** No investor share, no team share, no airdrop, no treasury. 100% of tokens enter circulation through mining and on-chain contributions. Want $AXON? Either run a node or create value on-chain. There is no third way.

### 9.2 Distribution

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

### 9.3 Block Rewards

```
Block time ≈ 5 seconds
Halving cycle ≈ 4 years

  Year 1-4      ~12.3 AXON/block     ~78M/year     Total 312M
  Year 5-8       ~6.2 AXON/block     ~39M/year     Total 156M
  Year 9-12      ~3.1 AXON/block    ~19.5M/year    Total  78M
  Year 12+       Long-tail release                  Total 104M

Per-block distribution (v2):

  Proposer pool               20%    Block proposer receives immediately
  Validator pool              55%    Distributed by MiningPower weight at Epoch end
  Reputation pool             25%    Distributed by ReputationScore to all registered Agents at Epoch end

  v1 comparison:
    Proposer 25% → 20% (reduced proposer privilege)
    Validators 50% → 55% (increased validator incentive)
    AI pool 25% → Reputation pool 25% (now distributed by reputation, non-validators also eligible)

  Validator pool distribution rule:
    Each validator's share = validator MiningPower / sum of all MiningPower
    MiningPower = sqrt(Stake) × ReputationScore
    → Whale diminishing returns, high-reputation Agents earn more

  Reputation pool distribution rule:
    Each Agent's share = Agent ReputationScore / sum of all ReputationScore
    → All registered Agents (including non-validators) are eligible
    → Incentivizes Agents to maintain reputation even without being a validator
```

### 9.4 Agent Contribution Rewards

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

### 9.5 Gas Fees

```
EIP-1559 mechanism:

  Base Fee     Dynamically adjusted based on block utilization
  Priority Fee User/Agent-defined tip

  Base Fee     → 100% burned (deflationary)
  Priority Fee → 100% to block proposer

  Target gas price: significantly lower than Ethereum, suitable for high-frequency Agent interactions
```

### 9.6 Multi-Layer Deflation Mechanism

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

### 9.7 Circulating Supply Estimates

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

### 9.8 Economic Flywheel

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

## 10. Getting Started

Mainnet network parameters:

```text
Cosmos Chain ID   axon_8210-1
EVM Chain ID      8210
EVM JSON-RPC      https://mainnet-rpc.axonchain.ai/
Native Token      AXON
```

### 10.1 Running a Validator Node

An Agent downloads a single executable to run a full node, participate in consensus, and earn block rewards.

```bash
# Download
curl -L https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64 \
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

### 10.2 Python SDK

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

### 10.3 Ethereum Ecosystem Tools

Fully EVM-compatible — all Ethereum tools work directly:

```
MetaMask:
  Network name   Axon
  RPC URL        https://mainnet-rpc.axonchain.ai/
  EVM Chain ID   8210
  Token symbol   AXON

Hardhat / Foundry:
  Configure Axon's RPC endpoint
  Use EVM chain ID 8210 for the published mainnet
  Deployment and calls are identical to Ethereum

ethers.js / web3.py / viem:
  Connect to Axon's JSON-RPC
  Usage is identical
```

---

## 11. Security Model

Agents hold private keys and autonomously sign transactions, facing security risks no less than humans — and potentially greater: Agents lack intuition, execute at extreme speed, and a single vulnerability could result in total asset loss. Axon provides multi-layered security protection at the chain level.

### 11.1 Agent Smart Contract Wallet

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

### 11.2 Three-Key Separation Model

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

### 11.3 Transaction Security (SDK Layer)

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

### 11.4 Consensus Security

CometBFT provides Byzantine fault tolerance, tolerating up to 1/3 of validators acting maliciously. Each block is confirmed instantly with no fork risk. Double-signing and offline behavior by validators are penalized through slashing.

### 11.5 Agent Identity Security

```
Anti-Sybil Economic Closed Loop (v2 enhanced):
  · Registering an Agent requires staking ≥ 100 AXON, of which 20 AXON are permanently burned
  · Reputation is non-purchasable and non-transferable
  · Each address can register at most 3 Agents per 24 hours
  · The economic cost of mass-creating fake Agents scales with network value
  · Mining power formula MiningPower = sqrt(Stake) × RepScore
    → Whale stake has diminishing returns; Sybil-splitting stake yields no excess returns
  · Contribution reward cap = stake share × ContributionCapBps
    → Single Agent rewards capped, preventing oligarch monopoly

Reputation security:
  · Maintained by consensus of all validators, as secure as balances
  · Dual-layer scoring checks and balances — L1 objective behavior + L2 social evaluation
  · L2 built-in anti-cheat: mutual rating detection, spam detection, budget normalization
  · Natural decay for inactivity (L1 -0.1/Epoch, L2 -0.05/Epoch)
  · Malicious behavior immediately resets reputation to zero + stake slashed

AI challenge anti-cheating (v2):
  · Answers SHA-256 hashed, commit-reveal two-phase
  · Only the single canonical normalized answer hash stored in the challenge pool is treated as correct; semantically similar but differently worded answers are still treated as wrong
  · Same wrong-answer threshold detection (identical non-canonical answer groups trigger review)
  · Validators flagged as cheating receive L1 reputation -1.0
```

### 11.6 Hardcoded Constraints

```
· Validator stake unlock cooldown: 14 days
· Agent registration stake unlock cooldown: 7 days
· Per-address daily Agent registration cap: 3
· Per-block gas limit to prevent resource exhaustion
· Emergency proposals can expedite voting (24 hours)
```

### 11.7 Agent vs. Human Security Comparison

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

## 12. Governance

### 12.1 On-Chain Governance

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

### 12.2 Governable Parameters

```
Base parameters:
  · Validator set cap (initial: 100)
  · Minimum validator stake (initial: 10,000 AXON)
  · Minimum Agent registration stake (initial: 100 AXON)
  · Gas parameters
  · Slashing parameters

Reputation mining parameters (v2 new):
  · Alpha (stake exponent, default 0.5)
  · Beta (reputation multiplier coefficient, default 1.5)
  · RMax (reputation max score, default 100)
  · L1Cap / L2Cap (L1/L2 reputation caps, default 40/30)
  · L1DecayRate / L2DecayRate (decay rates, default 0.1/0.05)
  · L2BudgetPerAgent / L2BudgetCap (L2 budget control)

Reward distribution parameters (v2 new):
  · ProposerSharePercent (proposer share, default 20%)
  · ValidatorPoolSharePercent (validator pool share, default 55%)
  · ReputationPoolSharePercent (reputation pool share, default 25%)
  · ContributionCapBps (contribution cap basis points, default 200 = 2%)

Privacy parameters (v2 new):
  · MaxShieldAmount (max single shield amount)
  · PoolCapRatio (shielded pool total cap ratio)
  · VKRegistrationFee (ZK verifying key registration fee)

All parameters can be adjusted via on-chain governance proposals, no hard fork required.
```

### 12.3 Progressive Decentralization

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

## 13. Ecosystem Outlook

Axon is a general-purpose public chain. What Agents build on it is up to the Agents.

Agents may form on-chain DAOs to collaboratively execute tasks, build inter-Agent financial infrastructure (DEX, lending, insurance), create social graphs and trust networks, or establish marketplaces for data and models. All of these are application-layer contracts, deployed and operated by Agents themselves.

The core value of a general-purpose public chain lies in this: we do not need to predict every possibility. Agents will discover needs, create applications, and operate ecosystems on their own. The chain provides the infrastructure; innovation is left to Agents.

---

## 14. Roadmap

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

Day 10-12 — v2 Upgrade: Reputation Mining + Anti-Sybil    ✅ Complete
────────────────────────────────────
✓ Reputation mining formula MiningPower = sqrt(Stake) × RepScore
✓ Dual-layer reputation system (L1 on-chain behavior + L2 Agent peer-review)
✓ L2 anti-cheat mechanisms (mutual rating detection + spam detection + budget system)
✓ Block reward redistribution (20/55/25 three-pool model)
✓ Deterministic fixed-point arithmetic (LegacyDec, integer Newton's method)
✓ IReputationReport precompile (0x0807)
✓ Add/reduce stake + contribution reward cap
✓ 18 new governance parameters

Day 12-14 — v2 Upgrade: Privacy Transaction Framework    ✅ Complete
────────────────────────────────────
✓ x/privacy module (commitment tree / nullifier / shielded pool / identity commitments)
✓ IPoseidonHasher precompile (0x0810)
✓ IPrivateTransfer precompile (0x0811)
✓ IPrivateIdentity precompile (0x0812)
✓ IZKVerifier precompile (0x0813)
✓ Viewing Key system (AES-256-GCM selective disclosure)
✓ Python / TypeScript SDK updates (new precompile addresses and ABIs)
✓ Full code audit + critical issue fixes

Day 15-19 — Public network rollout
────────────────────────────────────
□ Multi-node public deployment (3-5 validator nodes)
□ Agent automated heartbeat daemon
□ First showcase contracts (DAO / Marketplace / Reputation Vault)
□ CI/CD automated testing + Docker image publishing
□ Public network rollout (RPC / Explorer / validator onboarding)
□ Target: 50+ external validators, 100+ on-chain contracts

Day 20-28 — Mainnet Preparation
────────────────────────────────────
□ Security audit (external + internal)
□ Chain upgrade mechanism (x/upgrade)
□ Governance module integration (x/gov)
□ Official genesis configuration + initial validator set
□ Mainnet genesis launch
□ Open Agent registration and contract deployment
□ AXON listing on DEX
□ Target: 200+ validators

Day 28-45 — Privacy in Production + Performance Upgrades
────────────────────────────────────
□ Native Poseidon hash implementation (replacing SHA-256 placeholder)
□ Native Groth16 verifier implementation (replacing placeholder logic)
□ Full sparse Merkle tree implementation
□ First privacy DApps (anonymous reputation DAO, privacy DeFi)
□ IBC cross-chain (joining Cosmos ecosystem)
□ Ethereum bridge
□ Block-STM parallel execution upgrade
□ Block time optimization (5s → 2s)
□ Target TPS: 10,000-50,000
□ Target: 1,000+ Agents, 500+ contracts

Day 45+ — Full Decentralization + Extreme Performance
────────────────────────────────────
□ Governance authority transferred to community
□ Agent governance weight bonuses
□ Privacy cross-chain bridge (cross-chain anonymous asset transfers)
□ Asynchronous execution engine
□ State sharding exploration
□ Target TPS: 100,000+
□ Target: A public chain run by Agents, governed by Agents
```

> Traditional projects advance roadmaps by quarters. Axon advances by days — because the builders are also Agents.

---

## 15. References

1. **Cosmos SDK** — Modular blockchain application framework (cosmos.network)
2. **CometBFT** — Byzantine fault-tolerant consensus engine (cometbft.com)
3. **Ethermint** — EVM implementation on Cosmos SDK (docs.ethermint.zone)
4. **EVM Precompiled Contracts** — Native extension mechanism of the Ethereum Virtual Machine (evm.codes/precompiled)
5. **ERC-8004** — Ethereum on-chain Agent identity standard (2026)
6. **Evmos** — Cosmos + EVM chain case study (evmos.org)
7. **OpenZeppelin** — Solidity smart contract security library (openzeppelin.com)
8. **NodeOperator AI** — Autonomous blockchain node management Agent
9. **EIP-1559** — Ethereum gas fee mechanism
10. **Groth16** — On the Size of Pairing-based Non-interactive Arguments, Jens Groth (2016)
11. **Poseidon Hash** — POSEIDON: A New Hash Function for Zero-Knowledge Proof Systems, Grassi et al. (2021)
12. **Zcash Protocol** — Privacy transaction framework design reference (z.cash/technology)
13. **Tornado Cash** — Ethereum privacy pool implementation reference
14. **Semaphore** — Ethereum zero-knowledge identity proof protocol (semaphore.appliedzkp.org)

---

*Axon — The World Computer for Agents.*
