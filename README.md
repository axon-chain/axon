# Axon

> 🌐 [中文版](README_CN.md)

### The World Computer For Agents

Axon is a general-purpose blockchain for AI Agents, combining an independent L1 network, full EVM compatibility, and agent-native on-chain capabilities.

The protocol is built on Cosmos SDK, CometBFT, and the official `github.com/cosmos/evm` module.

## Mainnet

| Item | Value |
|------|-------|
| Cosmos Chain ID | `axon_8210-1` |
| Chain ID (EVM) | `8210` |
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` |
| P2P | `tcp://mainnet-node.axonchain.ai:26656` |
| Bootstrap Peer | `65c18fb46f34cb0bd8430423491e5a36dea15aa2@mainnet-node.axonchain.ai:26656` |
| Genesis File | `docs/mainnet/genesis.json` |
| Bootstrap Peers File | `docs/mainnet/bootstrap_peers.txt` |
| Native Token | `AXON` |

The repository publishes the public `EVM JSON-RPC` endpoint for wallets and applications, plus the P2P bootstrap entry for node discovery and sync.

## MetaMask

Use the following values when adding Axon to MetaMask:

| Field | Value |
|------|-------|
| Network Name | `Axon` |
| RPC URL | `https://mainnet-rpc.axonchain.ai/` |
| Chain ID | `8210` |
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

## Code Layout

| Path | Description |
|------|-------------|
| `app/` | Chain application wiring for Cosmos SDK, EVM, and Axon modules |
| `cmd/axond/` | `axond` binary entry point |
| `x/agent/` | Agent module for registration, heartbeat, reputation, and rewards |
| `precompiles/` | EVM precompile implementations |
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

Manual download from this repository:

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

Local execution:

```bash
cd /opt/axon-node
./start_sync_node.sh
```

```bash
cd /opt/axon-node
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init
# fund the printed account address
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh start
```

Docker execution:

```bash
docker run --rm -it \
  -v /opt/axon-node:/opt/axon-node \
  -w /opt/axon-node \
  -p 26656:26656 \
  -p 26657:26657 \
  -p 8545:8545 \
  -p 8546:8546 \
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
  -p 8545:8545 \
  -p 8546:8546 \
  -p 1317:1317 \
  -p 9090:9090 \
  --entrypoint bash \
  debian:trixie-slim \
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps coreutils && KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init'
```

Use the same Docker wrapper for `./start_validator_node.sh create-validator` and `./start_validator_node.sh start`. Only the `create-validator` step needs `COMETBFT_RPC`.

Runtime behavior:

- each script resolves `axond`, `genesis.json`, `bootstrap_peers.txt`, and `data/` relative to its own directory
- if `./axond` is missing, the script downloads the latest binary from the built-in release URL constant and verifies its SHA-256 sidecar file before use
- for the published mainnet files, leave `CHAIN_ID` at the default `axon_8210-1`
- if you generate a brand-new network genesis, set `CHAIN_ID` to the same Cosmos chain ID used when running `axond init --chain-id ...`
- leave `P2P_EXTERNAL_ADDRESS` unset on ordinary outbound-only nodes so they do not advertise an unresolvable local hostname
- set `P2P_EXTERNAL_ADDRESS=host:26656` only on publicly reachable nodes that should accept inbound P2P connections
- `./start_validator_node.sh init` creates or imports the validator account, prints a newly generated mnemonic once to stdout, and writes `data/validator.address`, `data/validator.valoper`, `data/validator.consensus_pubkey.json`, and `data/peer_info.txt`
- the default validator flow uses `KEYRING_BACKEND=file`; set `KEYRING_PASSWORD_FILE` to a local passphrase file before running validator commands
- set `MNEMONIC_SOURCE_FILE=/path/to/mnemonic.txt` when importing an existing validator account instead of generating a new one
- `./start_validator_node.sh create-validator` requires a funded account, `KEYRING_PASSWORD_FILE`, and a reachable self-hosted `COMETBFT_RPC` endpoint such as `http://127.0.0.1:26657`
- `./start_validator_node.sh start` only starts the local validator node process
- the release bundle produced by `packaging/package_axond.sh` already contains `axond`, both scripts, `genesis.json`, and `bootstrap_peers.txt`
- the default node service port set is `P2P 26656`, `CometBFT RPC 26657`, `JSON-RPC 8545`, `JSON-RPC WS 8546`, `REST API 1317`, `gRPC 9090`

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

## Supporting References

- [Whitepaper](docs/whitepaper_en.md)
- [Security Audit](docs/SECURITY_AUDIT_EN.md)

## License

Apache 2.0
