# Scripts

Public node startup scripts:

- `start_validator_node.sh`
- `start_sync_node.sh`

Each script is self-contained and resolves paths relative to its own directory.

- `axond` is expected at `./axond`; for mainnet or production nodes, pre-download it from the latest GitHub Release asset and verify the checksum before first run
- if `./axond` is missing, the script falls back to downloading it automatically from the latest GitHub Release asset
- downloaded binaries are verified against the matching `.sha256` sidecar file before use
- `genesis.json` is expected at `./genesis.json`
- `bootstrap_peers.txt` is expected at `./bootstrap_peers.txt`
- runtime data is stored under `./data/`
- the published mainnet files use Cosmos chain ID `axon_8210-1`
- if you generate a brand-new network genesis, the script `CHAIN_ID` must match the `chain_id` written into `genesis.json`
- leave `P2P_EXTERNAL_ADDRESS` unset on ordinary outbound-only nodes to avoid advertising an unresolvable local hostname
- set `P2P_EXTERNAL_ADDRESS=host:26656` only on publicly reachable nodes that should be dialed by other peers

Validator-specific behavior:

- the default validator keyring backend is `file`; set `KEYRING_PASSWORD_FILE=/path/to/passphrase` before running validator commands
- set `MNEMONIC_SOURCE_FILE=/path/to/mnemonic.txt` when importing an existing validator account
- `./start_validator_node.sh init` initializes `./data/node`, creates or imports the validator account, prints a newly generated mnemonic once to stdout, and writes `./data/validator.address`, `./data/validator.valoper`, `./data/validator.consensus_pubkey.json`, and `./data/peer_info.txt`
- the public mainnet validator script defaults `GAS_PRICES` to `1000000000aaxon` for Cosmos staking transactions; override it explicitly if the chain fee floor changes
- `./start_validator_node.sh start` applies the `validator-min` profile: aggressive state pruning, `tx_index=null`, `discard_abci_responses=true`, and JSON-RPC / REST / gRPC disabled for minimal disk growth
- CometBFT RPC stays bound to `127.0.0.1` by default so local validator operations still work
- `./start_validator_node.sh start` starts the local validator node process
- `./start_validator_node.sh status` reads the current official runtime paths under `./data/` and reports process/PID/home/log state
- `./start_validator_node.sh stop` stops the locally started validator node process
- `./start_validator_node.sh create-validator` submits the on-chain validator registration with a funded account, `KEYRING_PASSWORD_FILE`, and a reachable self-hosted `COMETBFT_RPC`, for example `http://127.0.0.1:26657`; when using the local RPC example, start the validator node first and run `create-validator` from another terminal

Sync-node behavior:

- initializes `./data/node`
- writes `./data/peer_info.txt`
- defaults to `SYNC_NODE_PROFILE=rpc-30d`, which retains about 30 days of state/block history and keeps the query interfaces needed for public RPC/API service
- supports `SYNC_NODE_PROFILE=archive` for full retained history
- supports `SYNC_NODE_PROFILE=p2p` for public P2P ingress only without JSON-RPC / REST / gRPC exposure
- `./start_sync_node.sh status` reads the current official runtime paths under `./data/` and reports process/PID/home/log state
- `./start_sync_node.sh stop` stops the locally started sync node process

Default local node service ports:

- `P2P 26656`
- `CometBFT RPC 26657`
- `JSON-RPC 8545`
- `JSON-RPC WS 8546`
- `REST API 1317`
- `gRPC 9090`

Profile notes:

- `validator-min` exposes only P2P plus local CometBFT RPC
- `rpc-30d` is the default sync-node profile for public RPC/API service
- `archive` exposes the same service surface while keeping full history
- `p2p` is optimized for peer ingress and does not expose JSON-RPC / REST / gRPC
