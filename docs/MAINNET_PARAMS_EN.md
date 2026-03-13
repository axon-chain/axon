> 🌐 [中文版](MAINNET_PARAMS.md)

# Axon Mainnet Parameter Configuration

This document records the core parameters and genesis configuration for the Axon mainnet and only includes values that are already published for public use.

---

## Chain Base Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Chain ID | `axon_8210-1` | Mainnet chain identifier |
| Native Token | `aaxon` | Smallest unit (1 AXON = 10¹⁸ aaxon) |
| Initial Supply | 0 | All tokens are produced through mining |

## Mainnet Connectivity

| Parameter | Value | Description |
|-----------|-------|-------------|
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` | Public wallet and contract RPC endpoint |
| P2P Address | `tcp://mainnet-node.axonchain.ai:26656` | Public P2P address |
| Bootstrap Peer | `65c18fb46f34cb0bd8430423491e5a36dea15aa2@mainnet-node.axonchain.ai:26656` | Current bootstrap node |
| Genesis File | `docs/mainnet/genesis.json` | Published genesis file in this repository |
| Bootstrap Peers File | `docs/mainnet/bootstrap_peers.txt` | Published bootstrap peers file in this repository |

Notes:

- This repository currently publishes the EVM JSON-RPC endpoint and the P2P bootstrap information
- CometBFT RPC is for node operations only and is not presented as wallet-facing public access information

## Consensus Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Block Gas Limit | 40,000,000 | Maximum gas consumption per block |
| Block Size Limit | 2 MB | Maximum bytes per block |
| Block Production Time | ~5 seconds | Target block interval |

## Staking

| Parameter | Value | Description |
|-----------|-------|-------------|
| Staking Token | `aaxon` | Token used for staking |
| Unbonding Period | 14 days | Freeze period after unstaking |
| Max Validators | 100 | Active validator cap |
| Min Commission Rate | 5% | Minimum validator commission rate |

## Slashing

| Parameter | Value | Description |
|-----------|-------|-------------|
| Signed Blocks Window | 10,000 blocks | Liveness detection window |
| Min Signed Per Window | 5% | Minimum signature rate within window |
| Downtime Jail Duration | 600 seconds | Jail time after downtime penalty |
| Double Sign Slash Fraction | 5% | Staking slash ratio for double signing |
| Downtime Slash Fraction | 0.1% | Staking slash ratio for downtime |

## Governance

| Parameter | Value | Description |
|-----------|-------|-------------|
| Min Proposal Deposit | 10,000 AXON | Deposit required to submit a proposal |
| Deposit Period | 2 days | Deposit collection deadline |
| Voting Period | 7 days | Proposal voting duration |
| Quorum | 33.4% | Minimum participation rate for a vote to pass |
| Pass Threshold | 50% | Required proportion of Yes votes |
| Veto Threshold | 33.4% | Threshold for strong veto votes |

## Mint

| Parameter | Value | Description |
|-----------|-------|-------------|
| Mint Token | `aaxon` | Minted token type |
| Inflation Rate Change | 0% | Standard inflation disabled |
| Max Inflation Rate | 0% | Standard inflation disabled |
| Min Inflation Rate | 0% | Standard inflation disabled |

> Note: The standard mint module is disabled. All tokens are produced through the custom mining mechanism in the Agent module.

## Distribution

| Parameter | Value | Description |
|-----------|-------|-------------|
| Community Tax | 0% | Deflation is achieved through the burn mechanism |
| Base Proposer Reward | 0% | Disabled |
| Bonus Proposer Reward | 0% | Disabled |

## Agent Module

| Parameter | Value | Description |
|-----------|-------|-------------|
| Min Registration Stake | 100 AXON | Staking amount required for Agent registration |
| Registration Burn Amount | 20 AXON | Amount permanently burned at registration |
| Max Reputation Score | 100 | Reputation score cap |
| Epoch Length | 720 blocks (~1 hour) | Reward cycle |
| Heartbeat Timeout | 720 blocks (~1 hour) | Heartbeat detection timeout threshold |
| AI Challenge Window | 50 blocks | AI verification challenge response time |
| Deregistration Cooldown | 120,960 blocks (~7 days) | Protocol constant (not a genesis-configurable parameter) |

## Fee Market (EIP-1559)

| Parameter | Value | Description |
|-----------|-------|-------------|
| Enable Base Fee | Yes | EIP-1559 mechanism enabled |
| Initial Base Fee | 1 gwei | Starting base gas price |

## EVM

| Parameter | Value | Description |
|-----------|-------|-------------|
| EVM Token | `aaxon` | Native token used by the EVM layer |

---

## Join The Existing Mainnet

```bash
mkdir -p /opt/axon-node
cd /opt/axon-node

curl -fsSLo start_sync_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_sync_node.sh
curl -fsSLo start_validator_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_validator_node.sh
curl -fsSLo genesis.json https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json
curl -fsSLo bootstrap_peers.txt https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/bootstrap_peers.txt
chmod 0755 start_sync_node.sh start_validator_node.sh

./start_sync_node.sh

# validator flow
./start_validator_node.sh init
# fund the printed account address
COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
./start_validator_node.sh start
```
