> 🌐 [中文版](README.md)

# Agent Heartbeat Daemon

A sidecar daemon for Axon nodes that automatically sends heartbeat transactions to the on-chain registry precompile contract, keeping the Agent's online status active.

Pass the published Axon EVM JSON-RPC endpoint: `https://mainnet-rpc.axonchain.ai/`.

## Features

- **Automatic Heartbeat**: Automatically sends heartbeat transactions every N blocks (default 100)
- **Registration Check**: Verifies whether the account is registered as an Agent at startup
- **Graceful Shutdown**: Supports safe exit via SIGINT / SIGTERM signals

## Usage

### Command-Line Arguments

| Argument | Default | Description |
|----------|---------|-------------|
| `--rpc` | (required) | JSON-RPC node address |
| `--private-key-file` | (required) | Path to a file containing the Agent account's hex private key |
| `--heartbeat-interval` | `100` | Number of blocks between each heartbeat |
| `--log-level` | `info` | Log level: debug, info, warn, error |

### Run Directly

```bash
go build -o agent-daemon .

./agent-daemon \
  --rpc https://mainnet-rpc.axonchain.ai/ \
  --private-key-file /path/to/private_key.txt \
  --heartbeat-interval 100
```

### Run with Docker

```bash
docker build -t agent-daemon .

docker run --rm \
  --network host \
  agent-daemon \
  --rpc https://mainnet-rpc.axonchain.ai/ \
  --private-key-file /run/secrets/agent_private_key
```

## How It Works

1. On startup, connects to the RPC node and retrieves the chain ID
2. Calls the registry precompile contract `isAgent(address)` to check if the current account is a registered Agent
3. Polls the latest block height; when the interval since the last heartbeat exceeds the configured threshold, constructs and signs a `heartbeat()` transaction
4. Sends the transaction via `eth_sendRawTransaction` and waits for receipt confirmation

## Important Notes

- Only `--private-key-file` is supported to avoid leaking secrets into shell history and process listings
- Keep your private key secure. In production environments, mount it from a secret store or key management service
- Ensure the Agent account has sufficient balance to pay gas fees
- Registry precompile contract address: `0x0000000000000000000000000000000000000801`
