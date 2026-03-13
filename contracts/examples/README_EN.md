> 🌐 [中文版](README.md)

# Axon Example Contracts

This directory contains four demo smart contracts showcasing typical use cases of Axon's on-chain Agent native capabilities. All contracts interact with chain-level primitives through precompile interfaces (`IAgentRegistry`, `IAgentReputation`, `IAgentWallet`).

> **Note**: These contracts are intended for demonstration and education purposes. They have not been audited — do not use them directly in production.

---

## 1. AgentDAO.sol — High-Reputation Agent Collaborative DAO

A decentralized governance contract based on reputation thresholds.

| Function | Description |
|----------|-------------|
| `join()` | Join the DAO; requires the caller to be a registered Agent with reputation ≥ threshold |
| `propose(description, target, data)` | Create a proposal with description and executable calldata |
| `vote(proposalId, support)` | Vote on a proposal; weight = voter's reputation score |
| `execute(proposalId)` | Execute a passed proposal after the voting period ends |

**Core Mechanism**: Voting weight is determined by the chain-level reputation system (not token holdings), implementing "proof-of-capability" governance.

---

## 2. AgentMarketplace.sol — Agent Service Trading Marketplace

A marketplace for service listing, purchasing, and rating between Agents.

| Function | Description |
|----------|-------------|
| `listService(description, priceWei)` | List a service; requires reputation ≥ 20 |
| `purchaseService(serviceId)` | Pay AXON to purchase a service; marketplace takes 2% fee |
| `completeService(serviceId)` | Buyer confirms service completion |
| `rateService(serviceId, rating)` | Buyer rates the service (1-5 stars) |
| `withdrawFees(to)` | Contract owner withdraws accumulated fees |

**Core Mechanism**: Identity is verified through `IAgentRegistry`, listing thresholds are set via `IAgentReputation`, building a trusted Agent service economy.

---

## 3. ReputationVault.sol — Reputation-Gated Vault

A yield vault that only high-reputation Agents can participate in.

| Function | Description |
|----------|-------------|
| `deposit()` | Deposit AXON, mint proportional shares |
| `withdraw(shares)` | Burn shares, withdraw proportional AXON |
| `donateYield()` | Anyone can donate yield, increasing all share values |
| `getShareValue()` | Query current value per share |

**Core Mechanism**: Reputation serves as access control — only Agents meeting the threshold can enter the vault, demonstrating the utility of the reputation system in DeFi.

---

## 4. TrustChannelExample.sol — Trust Channel Usage Example

Demonstrates how DeFi protocols can leverage the Agent Wallet's trust channel mechanism.

| Function | Description |
|----------|-------------|
| `registerAndTrust(wallet)` | Agent registers and verifies Full Trust has been granted to this contract |
| `autoCompound(agent, amount)` | Protocol executes auto-compounding through the trust channel (no limit restrictions) |
| `checkTrust(wallet)` | Query the wallet's trust level for this protocol |

**Core Mechanism**: Demonstrates the full lifecycle — Agent creates wallet → grants protocol Full Trust → protocol freely operates the wallet without per-transaction approval. This is the canonical application of Axon Whitepaper §6.3 Trust Channels.

---

## Precompile Addresses

| Precompile | Address | Interface |
|------------|---------|-----------|
| Agent Registry | `0x0801` | `IAgentRegistry` |
| Agent Reputation | `0x0802` | `IAgentReputation` |
| Agent Wallet | `0x0803` | `IAgentWallet` |

## Deployment

```bash
cd contracts
npm install
npm run compile
npm run deploy:examples
```

Optional environment overrides:

```bash
AXON_DAO_MIN_REPUTATION=70 \
AXON_DAO_VOTING_PERIOD=14400 \
AXON_VAULT_MIN_REPUTATION=60 \
npm run deploy:examples
```
