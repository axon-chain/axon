#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="$SCRIPT_DIR/data"
HOME_DIR="$DATA_DIR/node"
BINARY="$SCRIPT_DIR/axond"
GENESIS_FILE="$SCRIPT_DIR/genesis.json"
BOOTSTRAP_PEERS_FILE="$SCRIPT_DIR/bootstrap_peers.txt"
PEER_INFO_FILE="$DATA_DIR/peer_info.txt"
PID_FILE="$DATA_DIR/node.pid"
LOG_FILE="$DATA_DIR/node.log"
MNEMONIC_FILE="$DATA_DIR/validator.mnemonic"
ADDRESS_FILE="$DATA_DIR/validator.address"
VALOPER_FILE="$DATA_DIR/validator.valoper"
CONSENSUS_PUBKEY_FILE="$DATA_DIR/validator.consensus_pubkey.json"

CHAIN_ID="${CHAIN_ID:-axon_8210-1}"
DENOM="${DENOM:-aaxon}"
MIN_GAS_PRICES="${MIN_GAS_PRICES:-0${DENOM}}"
GAS_PRICES="${GAS_PRICES:-0${DENOM}}"
VALIDATOR_STAKE="${VALIDATOR_STAKE:-100000000000000000000${DENOM}}"
P2P_EXTERNAL_ADDRESS="${P2P_EXTERNAL_ADDRESS:-}"
P2P_PORT="${P2P_PORT:-26656}"
RPC_PORT="${RPC_PORT:-26657}"
JSON_RPC_ADDRESS="${JSON_RPC_ADDRESS:-0.0.0.0:8545}"
JSON_RPC_WS_ADDRESS="${JSON_RPC_WS_ADDRESS:-0.0.0.0:8546}"
API_ADDRESS="${API_ADDRESS:-tcp://0.0.0.0:1317}"
GRPC_ADDRESS="${GRPC_ADDRESS:-0.0.0.0:9090}"
MONIKER="${MONIKER:-axon-validator}"
KEY_NAME="${KEY_NAME:-validator}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
COMETBFT_RPC="${COMETBFT_RPC:-}"

AXOND_DOWNLOAD_URL_LINUX_AMD64="${AXOND_DOWNLOAD_URL_LINUX_AMD64:-https://assets.axonchain.ai/axond/latest/axond_linux_amd64}"
AXOND_DOWNLOAD_URL_LINUX_ARM64="${AXOND_DOWNLOAD_URL_LINUX_ARM64:-https://assets.axonchain.ai/axond/latest/axond_linux_arm64}"

log() {
    printf '==> %s\n' "$*"
}

die() {
    printf 'error: %s\n' "$*" >&2
    exit 1
}

need_cmd() {
    command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

platform_key() {
    local os=""
    local arch=""

    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) die "unsupported architecture: $(uname -m)" ;;
    esac

    printf '%s/%s\n' "$os" "$arch"
}

download_url() {
    case "$(platform_key)" in
        linux/amd64) printf '%s\n' "$AXOND_DOWNLOAD_URL_LINUX_AMD64" ;;
        linux/arm64) printf '%s\n' "$AXOND_DOWNLOAD_URL_LINUX_ARM64" ;;
        *) die "unsupported platform: $(platform_key)" ;;
    esac
}

ensure_binary() {
    if [ -x "$BINARY" ]; then
        return 0
    fi

    need_cmd curl
    log "Downloading axond binary"
    curl -fsSL "$(download_url)" -o "$BINARY"
    chmod 0755 "$BINARY"
}

require_file() {
    [ -f "$1" ] || die "required file does not exist: $1"
}

require_initialized_home() {
    [ -f "$HOME_DIR/config/config.toml" ] || die "validator home not initialized: $HOME_DIR (run ./start_validator_node.sh init first)"
}

stop_existing_node() {
    if [ ! -f "$PID_FILE" ]; then
        return 0
    fi

    local pid=""
    pid="$(cat "$PID_FILE" 2>/dev/null || true)"
    if [ -z "$pid" ]; then
        rm -f "$PID_FILE"
        return 0
    fi

    if ! kill -0 "$pid" >/dev/null 2>&1; then
        rm -f "$PID_FILE"
        return 0
    fi

    log "Stopping existing node process: $pid"
    kill "$pid" >/dev/null 2>&1 || true
    sleep 2
    rm -f "$PID_FILE"
}

bootstrap_peers_value() {
    python3 - "$BOOTSTRAP_PEERS_FILE" <<'PYEOF'
from pathlib import Path
import sys

path = Path(sys.argv[1])
lines = [line.strip() for line in path.read_text(encoding="utf-8").splitlines() if line.strip()]
if not lines:
    raise SystemExit(1)
print(",".join(lines))
PYEOF
}

configure_runtime_files() {
    local persistent_peers="$1"

    python3 - \
        "$HOME_DIR/config/app.toml" \
        "$HOME_DIR/config/config.toml" \
        "$MIN_GAS_PRICES" \
        "$persistent_peers" \
        "$P2P_EXTERNAL_ADDRESS" \
        "$RPC_PORT" \
        "$JSON_RPC_ADDRESS" \
        "$JSON_RPC_WS_ADDRESS" \
        "$API_ADDRESS" \
        "$GRPC_ADDRESS" <<'PYEOF'
from pathlib import Path
import re
import sys

app_path = Path(sys.argv[1])
config_path = Path(sys.argv[2])
minimum_gas_prices = sys.argv[3]
persistent_peers = sys.argv[4]
external_address = sys.argv[5]
rpc_port = sys.argv[6]
json_rpc_address = sys.argv[7]
json_rpc_ws_address = sys.argv[8]
api_address = sys.argv[9]
grpc_address = sys.argv[10]

def replace_root_value(text: str, key: str, value: str) -> str:
    pattern = rf'(^\s*{re.escape(key)}\s*=\s*)".*?"'
    updated, count = re.subn(pattern, rf'\1"{value}"', text, count=1, flags=re.MULTILINE)
    if count != 1:
        raise SystemExit(f"failed to update root key {key}")
    return updated

def replace_section_value(text: str, section: str, key: str, value: str) -> str:
    pattern = rf'(\[{re.escape(section)}\][\s\S]*?^\s*{re.escape(key)}\s*=\s*)".*?"'
    updated, count = re.subn(pattern, rf'\1"{value}"', text, count=1, flags=re.MULTILINE)
    if count != 1:
        raise SystemExit(f"failed to update [{section}] {key}")
    return updated

def replace_section_bool(text: str, section: str, key: str, value: bool) -> str:
    rendered = "true" if value else "false"
    pattern = rf'(\[{re.escape(section)}\][\s\S]*?^\s*{re.escape(key)}\s*=\s*)(true|false)'
    updated, count = re.subn(pattern, rf'\1{rendered}', text, count=1, flags=re.MULTILINE)
    if count != 1:
        raise SystemExit(f"failed to update [{section}] {key}")
    return updated

app_text = app_path.read_text(encoding="utf-8")
app_text = replace_root_value(app_text, "minimum-gas-prices", minimum_gas_prices)
app_text = replace_section_bool(app_text, "api", "enable", True)
app_text = replace_section_value(app_text, "api", "address", api_address)
app_text = replace_section_bool(app_text, "grpc", "enable", True)
app_text = replace_section_value(app_text, "grpc", "address", grpc_address)
app_text = replace_section_bool(app_text, "json-rpc", "enable", True)
app_text = replace_section_value(app_text, "json-rpc", "address", json_rpc_address)
app_text = replace_section_value(app_text, "json-rpc", "ws-address", json_rpc_ws_address)
app_path.write_text(app_text, encoding="utf-8")

config_text = config_path.read_text(encoding="utf-8")
config_text = replace_root_value(config_text, "external_address", external_address)
config_text = replace_root_value(config_text, "persistent_peers", persistent_peers)
config_text = replace_root_value(config_text, "laddr", f"tcp://0.0.0.0:{rpc_port}")
config_path.write_text(config_text, encoding="utf-8")
PYEOF
}

ensure_runtime_prereqs() {
    need_cmd python3
    ensure_binary
    require_file "$GENESIS_FILE"
    require_file "$BOOTSTRAP_PEERS_FILE"
    mkdir -p "$DATA_DIR"
}

ensure_home_exists() {
    if [ -f "$HOME_DIR/config/config.toml" ]; then
        return 0
    fi

    log "Initializing validator home: $HOME_DIR"
    "$BINARY" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null
}

install_genesis_and_configure() {
    log "Installing genesis"
    cp "$GENESIS_FILE" "$HOME_DIR/config/genesis.json"
    configure_runtime_files "$(bootstrap_peers_value)"
}

has_validator_key() {
    "$BINARY" keys show "$KEY_NAME" --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR" >/dev/null 2>&1
}

ensure_validator_key() {
    if has_validator_key; then
        return 0
    fi

    log "Creating validator account"
    python3 - "$MNEMONIC_FILE" "$ADDRESS_FILE" "$("$BINARY" keys add "$KEY_NAME" --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR" --output json)" <<'PYEOF'
from pathlib import Path
import json
import sys

mnemonic_path = Path(sys.argv[1])
address_path = Path(sys.argv[2])
payload = json.loads(sys.argv[3])
mnemonic_path.write_text(payload["mnemonic"] + "\n", encoding="utf-8")
address_path.write_text(payload["address"] + "\n", encoding="utf-8")
mnemonic_path.chmod(0o600)
PYEOF
}

require_validator_key() {
    has_validator_key || die "validator account not initialized (run ./start_validator_node.sh init first)"
}

account_address() {
    "$BINARY" keys show "$KEY_NAME" -a --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR"
}

validator_address() {
    "$BINARY" keys show "$KEY_NAME" --bech val -a --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR"
}

consensus_pubkey() {
    "$BINARY" comet show-validator --home "$HOME_DIR"
}

write_peer_info() {
    local node_id=""
    node_id="$("$BINARY" comet show-node-id --home "$HOME_DIR")"
    if [ -n "$P2P_EXTERNAL_ADDRESS" ]; then
        printf '%s@%s\n' "$node_id" "$P2P_EXTERNAL_ADDRESS" >"$PEER_INFO_FILE"
        return 0
    fi

    printf 'not advertised (node_id=%s)\n' "$node_id" >"$PEER_INFO_FILE"
}

write_validator_metadata() {
    printf '%s\n' "$(account_address)" >"$ADDRESS_FILE"
    printf '%s\n' "$(validator_address)" >"$VALOPER_FILE"
    printf '%s\n' "$(consensus_pubkey)" >"$CONSENSUS_PUBKEY_FILE"
}

validator_exists_on_chain() {
    local valoper_addr="$1"
    "$BINARY" query staking validator "$valoper_addr" --node "$COMETBFT_RPC" --output json >/dev/null 2>&1
}

submit_create_validator() {
    local validator_json="$DATA_DIR/create-validator.json"
    local pubkey=""

    pubkey="$(consensus_pubkey)"

    cat >"$validator_json" <<EOF
{
  "pubkey": ${pubkey},
  "amount": "${VALIDATOR_STAKE}",
  "moniker": "${MONIKER}",
  "identity": "",
  "website": "",
  "security": "",
  "details": "",
  "commission-rate": "0.10",
  "commission-max-rate": "0.20",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1"
}
EOF

    "$BINARY" tx staking create-validator "$validator_json" \
        --from "$KEY_NAME" \
        --keyring-backend "$KEYRING_BACKEND" \
        --home "$HOME_DIR" \
        --chain-id "$CHAIN_ID" \
        --node "$COMETBFT_RPC" \
        --gas auto \
        --gas-adjustment 1.5 \
        --gas-prices "$GAS_PRICES" \
        -y
}

print_init_summary() {
    echo
    echo "Validator node initialized."
    echo "  Home:              $HOME_DIR"
    echo "  Chain ID:          $CHAIN_ID"
    echo "  Account address:   $(cat "$ADDRESS_FILE")"
    echo "  Validator address: $(cat "$VALOPER_FILE")"
    echo "  Consensus pubkey:  $CONSENSUS_PUBKEY_FILE"
    echo "  Mnemonic:          $MNEMONIC_FILE"
    echo "  Peer:              $(cat "$PEER_INFO_FILE")"
    echo
    echo "Next steps:"
    echo "  1. Fund the account address shown above."
    echo "  2. Run: COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator"
    echo "  3. Run: ./start_validator_node.sh start"
    echo
}

print_create_validator_summary() {
    echo
    echo "Create-validator submitted."
    echo "  Validator address: $(cat "$VALOPER_FILE")"
    echo "  CometBFT RPC:      $COMETBFT_RPC"
    echo "  Stake:             $VALIDATOR_STAKE"
    echo
    echo "Next step:"
    echo "  ./start_validator_node.sh start"
    echo
}

command_init() {
    ensure_runtime_prereqs
    ensure_home_exists
    install_genesis_and_configure
    ensure_validator_key
    write_peer_info
    write_validator_metadata
    print_init_summary
}

command_create_validator() {
    ensure_runtime_prereqs
    require_initialized_home
    require_validator_key
    [ -n "$COMETBFT_RPC" ] || die "COMETBFT_RPC must be set, for example COMETBFT_RPC=http://127.0.0.1:26657"

    write_peer_info
    write_validator_metadata

    if validator_exists_on_chain "$(cat "$VALOPER_FILE")"; then
        log "Validator already exists on chain: $(cat "$VALOPER_FILE")"
        return 0
    fi

    submit_create_validator
    print_create_validator_summary
}

start_node() {
    local args=(
        start
        --home "$HOME_DIR" \
        --chain-id "$CHAIN_ID" \
        --minimum-gas-prices "$MIN_GAS_PRICES" \
        --p2p.laddr "tcp://0.0.0.0:${P2P_PORT}" \
        --p2p.persistent_peers "$(bootstrap_peers_value)" \
        --rpc.laddr "tcp://0.0.0.0:${RPC_PORT}"
    )

    if [ -n "$P2P_EXTERNAL_ADDRESS" ]; then
        args+=(--p2p.external-address "$P2P_EXTERNAL_ADDRESS")
    fi

    exec "$BINARY" "${args[@]}"
}

command_start() {
    ensure_runtime_prereqs
    require_initialized_home
    require_validator_key
    install_genesis_and_configure
    write_peer_info
    write_validator_metadata

    echo
    echo "Starting validator node."
    echo "  Home:              $HOME_DIR"
    echo "  Chain ID:          $CHAIN_ID"
    echo "  Account address:   $(cat "$ADDRESS_FILE")"
    echo "  Validator address: $(cat "$VALOPER_FILE")"
    echo "  Peer:              $(cat "$PEER_INFO_FILE")"
    echo "  Bootstrap:         $(bootstrap_peers_value)"
    echo

    stop_existing_node
    mkdir -p "$(dirname "$LOG_FILE")"
    echo "$$" >"$PID_FILE"
    exec > >(tee -a "$LOG_FILE") 2>&1
    start_node
}

usage() {
    cat <<'EOF'
Manage a validator node from the current directory.

Commands:
  init              initialize validator home, create account, and write local metadata
  create-validator  submit the on-chain create-validator transaction
  start             start the validator node process
  help              show this help message

Expected files in the script directory:
  - axond (optional; downloaded automatically when missing)
  - genesis.json
  - bootstrap_peers.txt

Runtime data:
  - data/node
  - data/validator.mnemonic
  - data/validator.address
  - data/validator.valoper
  - data/validator.consensus_pubkey.json
  - data/peer_info.txt
  - data/node.log

Typical flow:
  1. ./start_validator_node.sh init
  2. Fund the generated account address
  3. COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
  4. ./start_validator_node.sh start

Optional:
  - set P2P_EXTERNAL_ADDRESS=host:26656 only on publicly reachable nodes
EOF
}

COMMAND="${1:-help}"

case "$COMMAND" in
    init)
        command_init
        ;;
    create-validator)
        command_create_validator
        ;;
    start)
        command_start
        ;;
    help|-h|--help)
        usage
        ;;
    *)
        usage
        exit 1
        ;;
esac
