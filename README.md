# Axon

> 🌐 [中文版](README_CN.md)

### The World Computer For Agents

Axon is a general-purpose blockchain for AI Agents, combining an independent L1 network, full EVM compatibility, and agent-native on-chain capabilities.

The protocol is built on Cosmos SDK, CometBFT, and the official `github.com/cosmos/evm` module.

## Mainnet

| Item | Value |
|------|-------|
| Chain ID (Cosmos) | `axon_8210-1` |
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` |
| P2P | `tcp://mainnet-node.axonchain.ai:26656` |
| Bootstrap Peer | `65c18fb46f34cb0bd8430423491e5a36dea15aa2@mainnet-node.axonchain.ai:26656` |
| Genesis File | `docs/mainnet/genesis.json` |
| Bootstrap Peers File | `docs/mainnet/bootstrap_peers.txt` |
| Native Token | `AXON` |

The repository currently publishes the public `EVM JSON-RPC` endpoint and the P2P bootstrap entry.

RPC interface roles:

- `EVM JSON-RPC`: for wallets, MetaMask, contracts, and Ethereum-compatible clients
- `CometBFT RPC`: for node operators and Cosmos/CometBFT-side maintenance only

Ordinary users should use the Axon `EVM JSON-RPC` endpoint. `CometBFT RPC` is not part of the public wallet-facing access information.

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
sudo install -m 0755 ./build/axond /usr/local/bin/axond
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
PACKAGING_DOCKER_IMAGE=golang:1.25-bookworm bash packaging/build_release_matrix.sh --version v1.0.0
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
sudo mkdir -p /opt/axon-node
cd /opt/axon-node

sudo curl -fsSLo start_validator_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_validator_node.sh
sudo curl -fsSLo start_sync_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_sync_node.sh
sudo curl -fsSLo genesis.json https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json
sudo curl -fsSLo bootstrap_peers.txt https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/bootstrap_peers.txt
sudo chmod 0755 start_validator_node.sh start_sync_node.sh
```

Local execution:

```bash
cd /opt/axon-node
./start_sync_node.sh
```

```bash
cd /opt/axon-node
./start_validator_node.sh init
# fund the printed account address
COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
./start_validator_node.sh start
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
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps && ./start_sync_node.sh'
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
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps && ./start_validator_node.sh init'
```

Use the same Docker wrapper for `./start_validator_node.sh create-validator` and `./start_validator_node.sh start`. Only the `create-validator` step needs `COMETBFT_RPC`.

Runtime behavior:

- each script resolves `axond`, `genesis.json`, `bootstrap_peers.txt`, and `data/` relative to its own directory
- if `./axond` is missing, the script downloads the latest binary from the built-in release URL constant
- `./start_validator_node.sh init` creates the validator account and writes `data/validator.mnemonic`, `data/validator.address`, `data/validator.valoper`, `data/validator.consensus_pubkey.json`, and `data/peer_info.txt`
- `./start_validator_node.sh create-validator` requires a funded account and a reachable `COMETBFT_RPC` endpoint such as `http://127.0.0.1:26657`
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
tx = client.register_agent("nlp,reasoning", "gpt-4", stake_axon=100)
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
const tx = await client.registerAgent("nlp,reasoning", "gpt-4", "100");
await tx.wait();
```

Related implementations:

- Python client: `sdk/python/axon/client.py`
- TypeScript client: `sdk/typescript/src/client.ts`
- Agent daemon: `tools/agent-daemon/`

## Supporting References

- [Whitepaper](docs/whitepaper_en.md)
- [Mainnet Parameters](docs/MAINNET_PARAMS_EN.md)
- [Security Audit](docs/SECURITY_AUDIT_EN.md)

## License

Apache 2.0
