# Scripts

Public node startup scripts:

- `start_validator_node.sh`
- `start_sync_node.sh`

Each script is self-contained and resolves paths relative to its own directory.

- `axond` is expected at `./axond` and is downloaded automatically when missing
- `genesis.json` is expected at `./genesis.json`
- `bootstrap_peers.txt` is expected at `./bootstrap_peers.txt`
- runtime data is stored under `./data/`

Validator-specific behavior:

- `./start_validator_node.sh init` initializes `./data/node`, creates the validator account, and writes `./data/validator.mnemonic`, `./data/validator.address`, `./data/validator.valoper`, `./data/validator.consensus_pubkey.json`, and `./data/peer_info.txt`
- `./start_validator_node.sh create-validator` submits the on-chain validator registration with a funded account and a reachable `COMETBFT_RPC`, for example `http://127.0.0.1:26657`
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
