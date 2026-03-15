# Scripts

Public node startup scripts:

- `start_validator_node.sh`
- `start_sync_node.sh`

Each script is self-contained and resolves paths relative to its own directory.

- `axond` is expected at `./axond` and is downloaded automatically when missing
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
- `./start_validator_node.sh create-validator` submits the on-chain validator registration with a funded account, `KEYRING_PASSWORD_FILE`, and a reachable self-hosted `COMETBFT_RPC`, for example `http://127.0.0.1:26657`
- `./start_validator_node.sh start` starts the local validator node process

Sync-node behavior:

- initializes `./data/node`
- writes `./data/peer_info.txt`

Default local node service ports:

- `P2P 26656`
- `CometBFT RPC 26657`
- `JSON-RPC 8545`
- `JSON-RPC WS 8546`
- `REST API 1317`
- `gRPC 9090`
