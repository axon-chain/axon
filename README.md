# Axon

[![Release](https://github.com/axon-chain/axon/actions/workflows/release.yml/badge.svg)](https://github.com/axon-chain/axon/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/axon-chain/axon)](https://github.com/axon-chain/axon/releases/latest)
[![License](https://img.shields.io/github/license/axon-chain/axon)](LICENSE)

> 🌐 [中文版](README_CN.md)

### The World Computer For Agents

Axon is a general-purpose blockchain for AI Agents, combining an independent L1 network, full EVM compatibility, and agent-native on-chain capabilities.

Axon v2 introduces **Reputation Mining**, **Anti-Sybil Economic Loop**, and **Privacy Transaction Framework** — upgrading the consensus model from "PoS + reputation correction" to "PoS x reputation multiplier", and enabling on-chain anonymous identity proofs for Agents.

The protocol is built on Cosmos SDK, CometBFT, and the official `github.com/cosmos/evm` module.

## Mainnet

| Item | Value |
|------|-------|
| Cosmos Chain ID | `axon_8210-1` |
| EVM Chain ID | `8210` |
| P2P | `tcp://mainnet-node.axonchain.ai:26656` |
| Bootstrap Peers | `e47ec82a1d08a371e3c235e6554496be2f114eae@mainnet-node.axonchain.ai:26656` |
| Genesis File | `docs/mainnet/genesis.json` |
| Bootstrap Peers File | `docs/mainnet/bootstrap_peers.txt` |
| Native Token | `AXON` |

### Local Node Default Ports

These are the standard service ports used when a node profile enables the corresponding service.

| Service | Default Local Address | Notes |
|------|-------|-------|
| P2P | `tcp://127.0.0.1:26656` | Peer connectivity |
| CometBFT RPC | `http://127.0.0.1:26657` | Low-level chain RPC |
| EVM JSON-RPC | `http://127.0.0.1:8545` | Wallet and contract RPC |
| EVM JSON-RPC WebSocket | `ws://127.0.0.1:8546` | Local subscription transport |
| Cosmos REST API | `http://127.0.0.1:1317` | Standard REST, Axon routes, and `/axon/public/v1/` |
| gRPC | `127.0.0.1:9090` | Typed service access |

### Public API Entry Points

These are internally maintained public HTTPS/domain entries. Functionally they expose the same node capabilities as the local services above; they are not a separate protocol implementation.

| Service | Public Entry | Maps To |
|------|-------|-------|
| Unified API Entry | `https://mainnet-api.axonchain.ai/` | Local REST (1317) including `/cosmos/*` + `/axon/public/v1/*` |
| Runtime API Docs | `https://mainnet-api.axonchain.ai/docs/` | Unified API docs site |
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` | Local `8545` EVM JSON-RPC |
| CometBFT RPC | `https://mainnet-cometbft.axonchain.ai/` | Local `26657` CometBFT RPC |

Runtime API docs: `https://mainnet-api.axonchain.ai/docs/`

### Public RPC Policy

The public gateways above are shared infrastructure operated for public access. They are not dedicated private capacity.

Public RPC admission policy:

- Limits are enforced by gateway policy as global and per-IP controls. No dedicated quota is reserved per user or per API key.
- AI agents and autonomous clients must respect the rate and concurrency limits of each public API entry.
- Request limits are applied dynamically based on current server pressure.
- Agents and transaction-sending clients should use the EVM RPC entry for write traffic. This path gives transaction submission higher priority and helps reduce the impact of rate limiting on transaction inclusion.
- If the gateway returns `429`, it means your requests are arriving too frequently and need to be reduced or optimized. This is a rate-limit response, not a service fault.

## MetaMask

Use the following values when adding Axon to MetaMask:

| Field | Value |
|------|-------|
| Network Name | `Axon` |
| RPC URL | `https://mainnet-rpc.axonchain.ai/` |
| EVM Chain ID | `8210` |
| Currency Symbol | `AXON` |

MetaMask uses the EVM network identity, so the correct wallet-facing chain ID is `8210`.

## Chain IDs And Genesis

- The published Axon mainnet genesis file already fixes the Cosmos chain ID to `axon_8210-1`. Mainnet nodes must use that exact value.
- The wallet-facing EVM chain ID is `8210`. MetaMask and other Ethereum-compatible clients must use this value for signing and replay protection.
- When generating a brand-new network genesis from source, choose two IDs and keep them consistent across every node:
  - a globally unique Cosmos chain ID, typically in the form `axon_<network>-1`
  - an unused integer EVM chain ID
- The Cosmos chain ID is set during initialization with `axond init --chain-id <cosmos-chain-id>` and is written into the root `chain_id` field in `genesis.json`.
- A new public network must not reuse an existing public EVM chain ID.

## Mainnet Parameters

### Core Network

| Parameter | Value |
|-----------|-------|
| Cosmos Chain ID | `axon_8210-1` |
| EVM Chain ID | `8210` |
| Native EVM Denom | `aaxon` |
| Native Display Token | `AXON` |
| Initial Supply | `0` |

### Consensus

| Parameter | Value |
|-----------|-------|
| Block Gas Limit | `40,000,000` |
| Block Size Limit | `2 MB` |
| Target Block Time | `~5 seconds` |

### Staking

| Parameter | Value |
|-----------|-------|
| Staking Token | `aaxon` |
| Unbonding Period | `14 days` |
| Max Validators | `100` |
| Min Commission Rate | `5%` |

### Slashing

| Parameter | Value |
|-----------|-------|
| Signed Blocks Window | `10,000` |
| Min Signed Per Window | `5%` |
| Downtime Jail Duration | `600 seconds` |
| Double Sign Slash Fraction | `5%` |
| Downtime Slash Fraction | `0.1%` |

### Governance

| Parameter | Value |
|-----------|-------|
| Min Proposal Deposit | `10,000 AXON` |
| Deposit Period | `2 days` |
| Voting Period | `7 days` |
| Quorum | `33.4%` |
| Pass Threshold | `50%` |
| Veto Threshold | `33.4%` |

### Fee Market And Mint

| Parameter | Value |
|-----------|-------|
| Base Fee Enabled | `Yes` |
| Initial Base Fee | `1 gwei` |
| Mint Inflation | `0%` |
| Community Tax | `0%` |
| Base Proposer Reward | `0%` |
| Bonus Proposer Reward | `0%` |

The standard mint module is disabled. Token issuance is handled by the Agent module mining logic.

### Agent Module

| Parameter | Value |
|-----------|-------|
| Min Registration Stake | `100 AXON` |
| Registration Burn Amount | `20 AXON` |
| Max Reputation Score | `100` |
| Epoch Length | `720 blocks (~1 hour)` |
| Heartbeat Timeout | `720 blocks (~1 hour)` |
| AI Challenge Window | `50 blocks` |
| Deregistration Cooldown | `120,960 blocks (~7 days)` |

### Reputation Mining Parameters (v2)

| Parameter | Default | Description |
|-----------|---------|-------------|
| Alpha | `0.5` | Stake exponent (StakeScore = Stake^alpha) |
| Beta | `1.5` | Reputation multiplier coefficient |
| RMax | `100` | Max reputation score |
| L1Cap | `40` | L1 reputation cap |
| L2Cap | `30` | L2 reputation cap |
| L1DecayRate | `0.1` | L1 decay per epoch |
| L2DecayRate | `0.05` | L2 decay per epoch |
| L2BudgetPerAgent | `0.1` | Per-agent per-epoch L2 budget |
| L2BudgetCap | `100` | Total L2 budget cap per epoch |
| ProposerSharePercent | `20` | Proposer reward share |
| ValidatorPoolSharePercent | `55` | Validator pool share |
| ReputationPoolSharePercent | `25` | Reputation pool share |
| ContributionCapBps | `200` | Contribution reward cap (basis points, 200 = 2%) |

### Privacy Module Parameters (v2)

| Parameter | Default | Description |
|-----------|---------|-------------|
| MaxShieldAmount | `1,000,000 AXON` | Max single shield deposit |
| PoolCapRatio | `0.1` | Shielded pool cap (% of total supply) |
| VKRegistrationFee | `10 AXON` | Fee to register custom ZK verifying keys |

## Core Features (v2)

### Reputation Mining

Validator mining power is determined by the **PoS x Reputation Multiplier** formula:

```
MiningPower = sqrt(Stake) × (1 + 1.5 × ln(1 + Reputation) / ln(101))
```

- **StakeScore**: Square root of stake — diminishing marginal returns for whales
- **ReputationScore**: Ranges from 1.0 (zero reputation) to 2.0 (max reputation), up to 2x mining power
- All math uses `LegacyDec` fixed-point arithmetic for cross-platform consensus determinism

### Dual-Layer Reputation

| Layer | Source | Cap | Decay |
|-------|--------|-----|-------|
| L1 (On-chain behavior) | Block signing, heartbeats, on-chain activity, contract usage, AI challenges | 40 | -0.1/Epoch |
| L2 (Agent peer-review) | Agent-to-agent evaluation reports, anti-cheat filtered | 30 | -0.05/Epoch |

Total cap 100. L2 anti-cheat includes mutual rating detection (weight x0.1), spam detection (>50 positive reviews zeroed), and epoch budget normalization.

### AI Challenge Rule

- AI challenge correctness is determined only by exact `SHA256(normalizeAnswer(revealData))` match against the single answer hash stored in the challenge pool.
- `normalizeAnswer(...)` only lowercases ASCII letters and removes spaces, tabs, and newlines. It does not understand synonyms, paraphrases, or semantic equivalence.
- If three or more validators reveal the same non-canonical normalized answer, that answer group is still treated as colluding wrong answers and is penalized.

### Block Reward Distribution

| Pool | Share | Distribution Rule |
|------|-------|-------------------|
| Proposer | 20% | Immediate to current block proposer |
| Validator | 55% | End-of-epoch by MiningPower weight |
| Reputation | 25% | End-of-epoch by ReputationScore to all registered Agents |

### Privacy Transaction Framework

zk-SNARK (Groth16) + Poseidon hash powered privacy capabilities:

| Capability | Description |
|------------|-------------|
| Shielded Pool | Private transfers (transparent-to-private, private-to-transparent, in-pool) |
| Private Identity | Zero-knowledge proofs — Agents prove reputation >= N or stake >= M without revealing address |
| ZK Verifier | General Groth16 verifier with custom circuit registration |
| Viewing Key | Selective disclosure — viewing key holders can decrypt transaction details but cannot spend |

## Precompiled Contracts

| Address | Interface | Description |
|---------|-----------|-------------|
| `0x0...0801` | IAgentRegistry | Agent registration, heartbeat, stake management (`addStake`/`reduceStake`/`claimReducedStake`/`getStakeInfo`) |
| `0x0...0802` | IAgentReputation | Reputation query (returns combined L1+L2 score) |
| `0x0...0803` | IAgentWallet | Agent on-chain wallet (trusted channels, limits, freeze/recover) |
| `0x0...0807` | IReputationReport | L2 Agent peer-review system |
| `0x0...0810` | IPoseidonHasher | Poseidon hash (BN254 curve) |
| `0x0...0811` | IPrivateTransfer | Private transfers (shield/unshield/privateTransfer) |
| `0x0...0812` | IPrivateIdentity | Private identity proofs (ZK reputation/stake/capability proofs) |
| `0x0...0813` | IZKVerifier | General Groth16 ZK verifier |

Solidity interfaces are in `contracts/interfaces/`.
State-changing calls on `IAgentRegistry` are attributed to the immediate EVM caller (`msg.sender` / `contract.Caller()`), not `tx.origin`.

## Code Layout

| Path | Description |
|------|-------------|
| `app/` | Chain application wiring for Cosmos SDK, EVM, and Axon modules |
| `cmd/axond/` | `axond` binary entry point |
| `x/agent/` | Agent module — registration, heartbeat, reputation mining, dual-layer reputation, reward distribution, AI challenges |
| `x/privacy/` | Privacy module — commitment tree, nullifier set, shielded pool, identity commitments, viewing key |
| `precompiles/registry/` | 0x0801 Agent registry precompile |
| `precompiles/reputation/` | 0x0802 Reputation query precompile |
| `precompiles/wallet/` | 0x0803 Wallet precompile |
| `precompiles/report/` | 0x0807 L2 peer-review precompile |
| `precompiles/poseidon/` | 0x0810 Poseidon hash precompile |
| `precompiles/private_transfer/` | 0x0811 Private transfer precompile |
| `precompiles/private_identity/` | 0x0812 Private identity proof precompile |
| `precompiles/zk_verifier/` | 0x0813 ZK verifier precompile |
| `contracts/` | Solidity interfaces and sample contracts |
| `sdk/python/` | Python SDK |
| `sdk/typescript/` | TypeScript SDK |
| `scripts/` | Public scripts for joining an existing network |
| `ops/` | Release and operations utilities |
| `packaging/` | Release packaging scripts |
| `tools/agent-daemon/` | Agent heartbeat daemon |

## Build And Test

Requirements:

- Go `1.25+`
- `make`
- `git`
- Optional: `node` / `npm` for contract-side tests

Build `axond` from source:

```bash
git clone https://github.com/axon-chain/axon.git
cd axon
make build
./build/axond version
```

Install the binary to the default public script location:

```bash
install -m 0755 ./build/axond /usr/local/bin/axond
```

Run tests:

```bash
make test
go test ./... -count=1
```

Optional static checks:

```bash
gofmt -l ./x/agent/ ./app/ ./precompiles/ ./cmd/
go vet ./app/... ./cmd/... ./precompiles/... ./x/...
```

Optional contract-side tests:

```bash
cd contracts
npm install
npx hardhat test
```

## Release Packages

Official release archives are built by `packaging/build_release_matrix.sh` inside Docker. The default official target set is:

- `linux/amd64`
- `linux/arm64`

Archive naming:

- `axond_<version>_<os>_<arch>.tar.gz`
- `agent-daemon_<version>_<os>_<arch>.tar.gz`

Each release directory contains `SHA256SUMS` and `BUILD_REPORT.md`.

Override the builder image if required:

```bash
PACKAGING_DOCKER_IMAGE=golang:1.25.7-trixie bash packaging/build_release_matrix.sh --version v1.0.0
```

Verify checksums on Linux:

```bash
sha256sum -c SHA256SUMS
```

Verify checksums on macOS:

```bash
shasum -a 256 axond_<version>_<os>_<arch>.tar.gz
```

## Scripts

The public node startup workflow is directory-based. Use `/opt/axon-node/` as the working directory on both bare metal and Docker deployments.

Required files in `/opt/axon-node/`:

- `start_validator_node.sh`
- `start_sync_node.sh`
- `genesis.json`
- `bootstrap_peers.txt`

Supported public scripts:

| Script | Purpose |
|--------|---------|
| `scripts/start_validator_node.sh` | Manage validator initialization, account creation, `create-validator` submission, and node startup |
| `scripts/start_sync_node.sh` | Initialize local sync-node data and start the node |

Manual setup from GitHub:

```bash
mkdir -p /opt/axon-node
cd /opt/axon-node

curl -fsSLo start_validator_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_validator_node.sh
curl -fsSLo start_sync_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_sync_node.sh
curl -fsSLo genesis.json https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json
curl -fsSLo bootstrap_peers.txt https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/bootstrap_peers.txt
chmod 0755 start_validator_node.sh start_sync_node.sh
printf 'replace-with-a-strong-passphrase\n' > keyring.pass
chmod 0600 keyring.pass
```

Recommended pre-download from the latest GitHub Release:

```bash
curl -fsSLo axond https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64
curl -fsSLo axond.sha256 https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64.sha256
echo "$(cat axond.sha256)  axond" | sha256sum -c -
chmod 0755 axond
```

Local execution:

```bash
cd /opt/axon-node
./start_sync_node.sh
```

Archive sync node example:

```bash
cd /opt/axon-node
SYNC_NODE_PROFILE=archive ./start_sync_node.sh
```

```bash
cd /opt/axon-node
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh start
# fund the printed account address
# run this from another terminal after the local RPC is up
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
```

Docker execution:

```bash
docker run --rm -it \
  -v /opt/axon-node:/opt/axon-node \
  -w /opt/axon-node \
  -p 26656:26656 \
  -p 26657:26657 \
  -p 8545:8545 \
  -p 1317:1317 \
  -p 9090:9090 \
  --entrypoint bash \
  debian:trixie-slim \
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps coreutils && ./start_sync_node.sh'
```

```bash
docker run --rm -it \
  -v /opt/axon-node:/opt/axon-node \
  -w /opt/axon-node \
  -p 26656:26656 \
  -p 26657:26657 \
  --entrypoint bash \
  debian:trixie-slim \
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps coreutils && KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init'
```

Use the same Docker wrapper for `./start_validator_node.sh start` and then run `./start_validator_node.sh create-validator` from another terminal. Only the `create-validator` step needs `COMETBFT_RPC`, and the local validator RPC must already be running if you use `http://127.0.0.1:26657`.

Runtime behavior:

- each script resolves `axond`, `genesis.json`, `bootstrap_peers.txt`, and `data/` relative to its own directory
- for mainnet or production nodes, pre-download `./axond` from the latest GitHub Release and verify its SHA-256 before first run
- if `./axond` is missing, the script falls back to downloading the latest binary from the GitHub Releases `latest/download` asset URL and verifies its SHA-256 sidecar file before use
- the published Axon mainnet parameters are `CHAIN_ID=axon_8210-1` and `EVM_CHAIN_ID=8210`
- for the published mainnet files, leave `CHAIN_ID` at the default `axon_8210-1` and `EVM_CHAIN_ID` at the default `8210`
- if you generate a brand-new network genesis, set `CHAIN_ID` to the same Cosmos chain ID used when running `axond init --chain-id ...`
- if you generate a brand-new public network, choose a new unused `EVM_CHAIN_ID` and configure the same value on every node
- leave `P2P_EXTERNAL_ADDRESS` unset on ordinary outbound-only nodes so they do not advertise an unresolvable local hostname
- set `P2P_EXTERNAL_ADDRESS=host:26656` only on publicly reachable nodes that should accept inbound P2P connections
- `./start_validator_node.sh init` creates or imports the validator account, prints a newly generated mnemonic once to stdout, and writes `data/validator.address`, `data/validator.valoper`, `data/validator.consensus_pubkey.json`, and `data/peer_info.txt`
- the default validator flow uses `KEYRING_BACKEND=file`; set `KEYRING_PASSWORD_FILE` to a local passphrase file before running validator commands
- `./start_validator_node.sh start` applies the `validator-min` profile for the smallest practical disk footprint: aggressive state pruning, `tx_index=null`, `discard_abci_responses=true`, and only local CometBFT RPC enabled
- set `MNEMONIC_SOURCE_FILE=/path/to/mnemonic.txt` when importing an existing validator account instead of generating a new one
- the public mainnet validator flow defaults `GAS_PRICES` to `1000000000aaxon` for Cosmos staking transactions such as `create-validator`; override `GAS_PRICES` explicitly if the chain fee floor changes later
- `./start_validator_node.sh create-validator` requires a funded account, `KEYRING_PASSWORD_FILE`, and a reachable self-hosted `COMETBFT_RPC` endpoint such as `http://127.0.0.1:26657`; if you use the local validator RPC example, `./start_validator_node.sh start` must already be running in another terminal
- `./start_validator_node.sh start` only starts the local validator node process
- `./start_sync_node.sh` defaults to `SYNC_NODE_PROFILE=rpc-30d`, which keeps roughly 30 days of state and block history for public RPC/API service while retaining tx indexing
- `SYNC_NODE_PROFILE=archive ./start_sync_node.sh` keeps full history for archive-style public query service
- `SYNC_NODE_PROFILE=p2p ./start_sync_node.sh` creates a public P2P ingress node without JSON-RPC / REST / gRPC exposure
- `./start_validator_node.sh status` and `./start_sync_node.sh status` report state from the official current-directory runtime paths under `data/`, so operators should prefer these commands over any older external status helper scripts
- the release bundle produced by `packaging/package_axond.sh` already contains `axond`, both scripts, `genesis.json`, and `bootstrap_peers.txt`
- the default node service port set is `P2P 26656`, `CometBFT RPC 26657`, `JSON-RPC 8545`, `REST API 1317`, `gRPC 9090`

## SDK

Axon exposes public Python and TypeScript SDKs.

| Language | Path |
|----------|------|
| Python | `sdk/python/` |
| TypeScript | `sdk/typescript/` |

Python SDK install:

```bash
pip install -e sdk/python
```

Python example:

```python
from axon import AgentClient
import os

client = AgentClient(os.environ["AXON_RPC_URL"])
client.set_account(os.environ["AXON_PRIVATE_KEY"])
tx = client.register_agent("nlp,reasoning", "axon-demo-model", stake_axon=100)
```

TypeScript SDK install:

```bash
cd sdk/typescript
npm install
```

TypeScript example:

```typescript
import { AgentClient } from "@axon-chain/sdk";

const client = new AgentClient(process.env.AXON_RPC_URL!);
client.connect(process.env.AXON_PRIVATE_KEY!);
const tx = await client.registerAgent("nlp,reasoning", "axon-demo-model", "100");
await tx.wait();
const addStakeTx = await client.addStake("500");
await addStakeTx.wait();
```

Related implementations:

- Python client: `sdk/python/axon/client.py`
- TypeScript client: `sdk/typescript/src/client.ts`
- Agent daemon: `tools/agent-daemon/`

## Architecture Overview

```
┌──────────────────────────────────────────────────────┐
│                     EVM Layer                        │
│  ┌────────────────────────────────────────────────┐  │
│  │  Solidity Contracts / Agent DApps              │  │
│  │  ↕  ↕  ↕  ↕  ↕  ↕  ↕  ↕                      │  │
│  │  Precompiles (0x0801 ~ 0x0813)                 │  │
│  └────────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────────┤
│                  Application Layer                   │
│  ┌──────────────────┐  ┌──────────────────────────┐  │
│  │    x/agent        │  │      x/privacy           │  │
│  │  ├ registration   │  │  ├ commitment tree       │  │
│  │  ├ heartbeat      │  │  ├ nullifier set         │  │
│  │  ├ l1_reputation  │  │  ├ shielded pool         │  │
│  │  ├ l2_reputation  │  │  ├ identity commitments  │  │
│  │  ├ mining_power   │  │  ├ viewing key           │  │
│  │  ├ block_rewards  │  │  └ zk verifying keys     │  │
│  │  └ ai_challenge   │  └──────────────────────────┘  │
│  └──────────────────┘                                │
├──────────────────────────────────────────────────────┤
│              Cosmos SDK + CometBFT                   │
│  bank · staking · gov · slashing · distribution      │
│  consensus · evidence · fee-market · cosmos/evm      │
└──────────────────────────────────────────────────────┘
```

## Supporting References

- [Whitepaper](docs/whitepaper_en.md)
- [v2 Upgrade Proposal](Axon_v2_升级产品方案.md)
- [Security Audit](docs/SECURITY_AUDIT_EN.md)
- [Community Tools (use at your own risk)](docs/community-tools.md)

## License

Apache 2.0
