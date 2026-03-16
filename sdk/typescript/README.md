# Axon TypeScript SDK

TypeScript SDK for interacting with the Axon AI Agent blockchain. Built on [ethers.js v6](https://docs.ethers.org/v6/).

Use the published Axon EVM JSON-RPC endpoint: `https://mainnet-rpc.axonchain.ai/`.

## Mainnet Parameters

| Item | Value |
|------|-------|
| Cosmos Chain ID | `axon_8210-1` |
| EVM Chain ID | `8210` |
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` |
| Native Token | `AXON` |

The SDK signs transactions against the published Axon mainnet EVM chain ID `8210`.

## Installation

```bash
npm install @axon-chain/sdk
```

## Quick Start

```typescript
import { AgentClient } from "@axon-chain/sdk";

const client = new AgentClient("https://mainnet-rpc.axonchain.ai/");

// Query chain info
const chainId = await client.getChainId();
const block = await client.getBlockNumber();
console.log(`Chain ${chainId} — Block #${block}`);

// Check if an address is a registered agent
const registered = await client.isAgent("0x1234...");

// Query reputation
const rep = await client.getReputation("0x1234...");
```

## Signing Transactions

Pass a private key to the constructor or call `connect()`:

```typescript
const client = new AgentClient("https://mainnet-rpc.axonchain.ai/");
client.connect(process.env.AXON_PRIVATE_KEY!);
// — or —
const client2 = new AgentClient("https://mainnet-rpc.axonchain.ai/", process.env.AXON_PRIVATE_KEY!);

// Register as an AI Agent (stakes 100 AXON)
const tx = await client.registerAgent("nlp,vision", "axon-7b", "100");
await tx.wait();

// Add more stake later without re-registering
const topUp = await client.addStake("500");
await topUp.wait();

// Send AXON
const transfer = await client.transfer("0xRecipient...", "5.0");
await transfer.wait();
```

## Agent Wallet

```typescript
// Create a wallet (caller becomes Owner)
const tx = await client.createWallet(operatorAddr, guardianAddr, "10", "100", 10);
await tx.wait();

// Execute through the wallet
await client.executeWallet(walletAddr, targetAddr, "1.5");

// Set trust level for a contract
import { TRUST_FULL } from "@axon-chain/sdk";
await client.setTrust(walletAddr, contractAddr, TRUST_FULL, "5", "50", 0);
```

## Features

- Agent registration, heartbeat, deregistration
- Reputation queries (single and batch)
- Smart contract deployment and interaction
- Agent Wallet management via precompiles
- Trusted Channel authorization
- Full EVM compatibility
