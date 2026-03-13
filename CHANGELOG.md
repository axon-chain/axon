# Changelog

## v1.0.0 - Initial Public Release

### Core Chain
- Cosmos SDK v0.54 with official Cosmos EVM integration
- CometBFT consensus with Axon `x/agent` module
- Full EVM compatibility and JSON-RPC support
- EIP-1559 fee market with fee burn logic
- Agent-native registry, reputation, and wallet precompiles

### Economic Model
- Fixed-supply token model with zero pre-allocation
- Block rewards and contribution rewards managed by the Agent module
- Deflation paths for gas fees, registration, deployment, reputation loss, and cheating penalties
- Reputation and AI bonus adjustments integrated into validator reward weight

### Tooling
- `axond` node binary
- Agent heartbeat daemon in `tools/agent-daemon/`
- Python SDK in `sdk/python/`
- TypeScript SDK in `sdk/typescript/`
- Public node startup scripts in `scripts/`
- Multi-platform release packaging in `packaging/`

### Contracts
- Solidity interfaces for Agent registry, reputation, and wallet precompiles
- Example contracts for DAO, marketplace, vault, and trust channel workflows
- Hardhat-based contract test and deployment tooling

### Documentation
- Dual-language README files as the primary public documentation entrypoints
- Supplementary references in `docs/`

### Repository Organization
- Public startup scripts separated from operations utilities and packaging scripts
- Mainnet deployment parameters normalized to the published public network configuration
- Release workflow aligned with repository packaging scripts
