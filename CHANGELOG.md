# Changelog

## Unreleased

## v1.1.1 - Security Hardening

### Consensus Upgrade

- **Upgrade Name**: `v1.1.1`
- **Mainnet Chain ID**: `axon_8210-1`
- **Upgrade Height**: `295500` (estimated 2026-04-04 20:00 CST, based on mainnet avg block time 5.39s)
- **Upgrade Type**: consensus-breaking (F1/F2/F3/F5/F8/F9), requires coordinated `software-upgrade` governance proposal
- **Historical Replay**: previously activated v1.1.0 behavior remains gated by `IsV110UpgradeActivated()` at height `259051`; new v1.1.1 behavior is gated separately by `IsV111UpgradeActivated()` at height `295500`

New v1.1.1 consensus changes activate at block 295500. Non-mainnet chains (`chainID != axon_8210-1`) activate both v1.1.0 and v1.1.1 behavior immediately for testnet convenience.

### Security Fixes
- [F1] AI challenge anti-cheat no longer penalizes validators for submitting the exact canonical normalized answer stored in the challenge pool; collusion detection now only triggers on identical non-canonical answers. `detectCheaters` accepts an `expectedHash` parameter and only skips the answer group whose `SHA256(normalizeAnswer(revealData))` exactly matches the stored `expectedHash`.
- [F2] Private identity (`0x0812`) and reputation report (`0x0807`) precompiles now bind mutations to the immediate contract caller (`contract.Caller()`) after the v1.1.1 upgrade, removing `tx.origin` confused-deputy behavior while preserving historical replay compatibility.
- [F3] L2 reputation evidence now requires a valid 32-byte transaction hash format and a chain-indexed EVM transaction record before full evidence weight is granted. Evidence normalization is applied only post-upgrade to preserve exact v1.0.0 weight semantics for replay.
- [F9] Private identity registrations now store an `agent -> commitment` reverse index after the v1.1.1 upgrade and remove identity state during `deregister` queue execution. Historical replay is preserved by keeping the legacy one-byte marker format before the upgrade height and deleting only the agent-side marker when old data lacks a reverse index.

### API Hardening
- [F4] Public transaction search API now validates `sender`, `recipient`, and `type` query parameters against format whitelists (`isValidAddress`, `isValidActionType`) and strips single quotes via `sanitizeQueryStringValue` before constructing CometBFT queries.
- [F6] Public API `simulate` and `broadcast` endpoints now enforce a 2 MB request body limit via `http.MaxBytesReader`.
- [F7] Public API response cache now enforces a maximum entry count of 10,000 with expired-entry pruning before insertion.
- Unknown API endpoints now return HTTP 404 instead of gRPC 501 (UNIMPLEMENTED).

### Observability
- Peer version visibility: each node injects a geth-style client name (`axond/<version>/<os>-<arch>/<go>`) into its CometBFT moniker at startup, making peer software versions visible via p2p handshake.
- `/chain/status` API now exposes `client_name` and a `peers` array with each peer's `node_id`, `name`, `moniker`, `remote_ip`, `network`, and `is_outbound`.

### Optimization
- [F5] Epoch-scoped Agent KV data (Challenge, AIResponse, EpochActivity, DeployCount, ContractCall) is cleaned up 2 epochs after settlement. Stale daily registration counters are cleaned up in batches (max 5,000 per block) to avoid gas spikes on first run. Evidence tx hashes are retained for 1 day (17,280 blocks) and cleaned up with height-indexed lookup.
- [F8] Contribution scoring now caps deploy and call counters at 10,000 before `int64` conversion and weight calculation to prevent extreme-value overflow paths.

### Economic Model
- AI performance rewards now distribute by eligible stake weight instead of per-Agent count.
- Contribution reward caps now scale by stake share, closing the fixed per-Agent `2%` Sybil bypass.
- Registered Agents can now increase stake in place via `AddStake`.

### EVM / SDK
- `IAgentRegistry.register(...)` and `IAgentRegistry.addStake()` now use `payable` + `msg.value` semantics consistently across the precompile and both SDKs.
- EVM `PostTxProcessing` hook now records all transaction hashes post-upgrade for L2 evidence validation.

### Design
- Added ADR 0001 for Agent economics hardening and the deferred privacy/ZK roadmap.

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
