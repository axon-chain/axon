# Axon Python SDK

Python SDK for interacting with the Axon AI Agent blockchain.

Use the published Axon EVM JSON-RPC endpoint: `https://mainnet-rpc.axonchain.ai/`.

## Installation

```bash
pip install -e sdk/python
```

## Quick Start

```python
from axon import AgentClient
import os

client = AgentClient("https://mainnet-rpc.axonchain.ai/")

# Query chain info
print(f"Chain ID: {client.chain_id}")
print(f"Block: {client.block_number}")

# Check if address is an agent
is_agent = client.is_agent("0x1234...")

# Query reputation
rep = client.get_reputation("0x1234...")

# Register as agent (requires private key)
client.set_account(os.environ["AXON_PRIVATE_KEY"])
tx = client.register_agent("nlp,vision", "axon-demo-model", stake_axon=100)

# Add more stake later without re-registering
top_up = client.add_stake(500)
```

## Features

- Agent registration, stake top-ups, heartbeat, deregistration
- Reputation and AI bonus queries
- Smart contract deployment and interaction
- Wallet management via precompiles
- Full EVM compatibility (works with any ERC-20, NFT, etc.)
