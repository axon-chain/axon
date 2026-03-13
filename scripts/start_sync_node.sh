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

CHAIN_ID="${CHAIN_ID:-axon_8210-1}"
DENOM="${DENOM:-aaxon}"
MIN_GAS_PRICES="${MIN_GAS_PRICES:-0${DENOM}}"
P2P_EXTERNAL_ADDRESS="${P2P_EXTERNAL_ADDRESS:-}"
P2P_PORT="${P2P_PORT:-26656}"
RPC_PORT="${RPC_PORT:-26657}"
JSON_RPC_ADDRESS="${JSON_RPC_ADDRESS:-0.0.0.0:8545}"
JSON_RPC_WS_ADDRESS="${JSON_RPC_WS_ADDRESS:-0.0.0.0:8546}"
API_ADDRESS="${API_ADDRESS:-tcp://0.0.0.0:1317}"
GRPC_ADDRESS="${GRPC_ADDRESS:-0.0.0.0:9090}"
MONIKER="${MONIKER:-axon-sync}"

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
    mkdir -p "$SCRIPT_DIR"
    log "Downloading axond binary"
    curl -fsSL "$(download_url)" -o "$BINARY"
    chmod 0755 "$BINARY"
}

require_file() {
    [ -f "$1" ] || die "required file does not exist: $1"
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

write_peer_info() {
    local node_id=""
    node_id="$("$BINARY" comet show-node-id --home "$HOME_DIR")"
    if [ -n "$P2P_EXTERNAL_ADDRESS" ]; then
        printf '%s@%s\n' "$node_id" "$P2P_EXTERNAL_ADDRESS" >"$PEER_INFO_FILE"
        return 0
    fi

    printf 'not advertised (node_id=%s)\n' "$node_id" >"$PEER_INFO_FILE"
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

usage() {
    cat <<'EOF'
Start a sync node from the current directory.

Expected files in the script directory:
  - axond (optional; downloaded automatically when missing)
  - genesis.json
  - bootstrap_peers.txt

Runtime data:
  - data/node
  - data/node.log
  - data/peer_info.txt

Optional:
  - set P2P_EXTERNAL_ADDRESS=host:26656 only on publicly reachable nodes
EOF
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    usage
    exit 0
fi

need_cmd python3
ensure_binary
require_file "$GENESIS_FILE"
require_file "$BOOTSTRAP_PEERS_FILE"

mkdir -p "$DATA_DIR"
if [ ! -f "$HOME_DIR/config/config.toml" ]; then
    log "Initializing node home: $HOME_DIR"
    "$BINARY" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null
fi

log "Installing genesis"
cp "$GENESIS_FILE" "$HOME_DIR/config/genesis.json"
configure_runtime_files "$(bootstrap_peers_value)"
write_peer_info

echo
echo "Sync node is configured."
echo "  Home:      $HOME_DIR"
echo "  Chain ID:  $CHAIN_ID"
echo "  Peer:      $(cat "$PEER_INFO_FILE")"
echo "  Upstream:  $(bootstrap_peers_value)"
echo

stop_existing_node
mkdir -p "$(dirname "$LOG_FILE")"
echo "$$" >"$PID_FILE"
exec > >(tee -a "$LOG_FILE") 2>&1
start_node
