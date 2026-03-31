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
SYNC_NODE_PROFILE="${SYNC_NODE_PROFILE:-rpc-30d}"
RPC_LADDR=""
PROFILE_LABEL=""

AXOND_DOWNLOAD_URL_LINUX_AMD64="${AXOND_DOWNLOAD_URL_LINUX_AMD64:-https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64}"
AXOND_DOWNLOAD_URL_LINUX_ARM64="${AXOND_DOWNLOAD_URL_LINUX_ARM64:-https://github.com/axon-chain/axon/releases/latest/download/axond_linux_arm64}"
AXOND_DOWNLOAD_SHA256_URL_LINUX_AMD64="${AXOND_DOWNLOAD_SHA256_URL_LINUX_AMD64:-https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64.sha256}"
AXOND_DOWNLOAD_SHA256_URL_LINUX_ARM64="${AXOND_DOWNLOAD_SHA256_URL_LINUX_ARM64:-https://github.com/axon-chain/axon/releases/latest/download/axond_linux_arm64.sha256}"

log() {
    printf '==> %s\n' "$*"
}

die() {
    printf 'error: %s\n' "$*" >&2
    exit 1
}

normalize_sync_node_profile() {
    local raw="${1:-}"
    local normalized=""

    normalized="${raw,,}"
    case "$normalized" in
        rpc-30d|archive|p2p)
            printf '%s\n' "$normalized"
            ;;
        *)
            die "SYNC_NODE_PROFILE must be one of: rpc-30d, archive, p2p"
            ;;
    esac
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

checksum_url() {
    case "$(platform_key)" in
        linux/amd64) printf '%s\n' "$AXOND_DOWNLOAD_SHA256_URL_LINUX_AMD64" ;;
        linux/arm64) printf '%s\n' "$AXOND_DOWNLOAD_SHA256_URL_LINUX_ARM64" ;;
        *) die "unsupported platform: $(platform_key)" ;;
    esac
}

checksum_tool() {
    if command -v sha256sum >/dev/null 2>&1; then
        printf 'sha256sum\n'
        return 0
    fi
    if command -v shasum >/dev/null 2>&1; then
        printf 'shasum\n'
        return 0
    fi
    die "missing required command: sha256sum or shasum"
}

verify_checksum() {
    local file_path="$1"
    local checksum_path="$2"
    local expected=""

    expected="$(awk 'NF {print $1; exit}' "$checksum_path")"
    [ -n "$expected" ] || die "checksum file is empty: $checksum_path"

    case "$(checksum_tool)" in
        sha256sum)
            [ "$(sha256sum "$file_path" | awk '{print $1}')" = "$expected" ] || die "sha256 mismatch for $file_path"
            ;;
        shasum)
            [ "$(shasum -a 256 "$file_path" | awk '{print $1}')" = "$expected" ] || die "sha256 mismatch for $file_path"
            ;;
    esac
}

ensure_binary() {
    if [ -x "$BINARY" ]; then
        return 0
    fi

    need_cmd curl
    local tmp_binary=""
    local tmp_checksum=""
    tmp_binary="$(mktemp "$SCRIPT_DIR/.axond.XXXXXX")"
    tmp_checksum="$(mktemp "$SCRIPT_DIR/.axond.sha256.XXXXXX")"
    trap 'rm -f "$tmp_binary" "$tmp_checksum"' RETURN
    log "Downloading axond binary"
    curl -fsSL "$(download_url)" -o "$tmp_binary"
    curl -fsSL "$(checksum_url)" -o "$tmp_checksum"
    verify_checksum "$tmp_binary" "$tmp_checksum"
    mv "$tmp_binary" "$BINARY"
    chmod 0755 "$BINARY"
    rm -f "$tmp_checksum"
    trap - RETURN
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

node_pid() {
    if [ ! -f "$PID_FILE" ]; then
        return 1
    fi

    local pid=""
    pid="$(cat "$PID_FILE" 2>/dev/null || true)"
    [ -n "$pid" ] || return 1
    printf '%s\n' "$pid"
}

node_process_running() {
    local pid=""
    pid="$(node_pid)" || return 1
    kill -0 "$pid" >/dev/null 2>&1
}

command_status() {
    local process_status="stopped"
    local pid_display="not found"
    local home_status="missing"
    local log_status="missing"
    local pid_value=""

    if [ -f "$HOME_DIR/config/config.toml" ]; then
        home_status="$HOME_DIR"
    fi

    if [ -f "$LOG_FILE" ]; then
        log_status="$LOG_FILE"
    fi

    if node_process_running; then
        process_status="running"
        pid_display="$(node_pid)"
    elif pid_value="$(node_pid 2>/dev/null)"; then
        process_status="stale pid"
        pid_display="$pid_value"
    fi

    echo "Sync node status"
    echo "  Process: $process_status"
    echo "  PID:     $pid_display"
    echo "  Home:    $home_status"
    echo "  Log:     $log_status"
    echo "  Peer:    ${PEER_INFO_FILE}"
}

command_stop() {
    stop_existing_node
    echo "Sync node stop completed."
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

    [ -f "$HOME_DIR/config/app.toml" ] || die "missing generated app.toml at $HOME_DIR/config/app.toml after init"
    [ -f "$HOME_DIR/config/config.toml" ] || die "missing generated config.toml at $HOME_DIR/config/config.toml after init"

    python3 - \
        "$HOME_DIR/config/app.toml" \
        "$HOME_DIR/config/config.toml" \
        "$MIN_GAS_PRICES" \
        "$persistent_peers" \
        "$P2P_EXTERNAL_ADDRESS" \
        "$RPC_LADDR" \
        "$JSON_RPC_ADDRESS" \
        "$JSON_RPC_WS_ADDRESS" \
        "$API_ADDRESS" \
        "$GRPC_ADDRESS" \
        "$SYNC_NODE_PROFILE" <<'PYEOF'
from pathlib import Path
import re
import sys

app_path = Path(sys.argv[1])
config_path = Path(sys.argv[2])
minimum_gas_prices = sys.argv[3]
persistent_peers = sys.argv[4]
external_address = sys.argv[5]
rpc_laddr = sys.argv[6]
json_rpc_address = sys.argv[7]
json_rpc_ws_address = sys.argv[8]
api_address = sys.argv[9]
grpc_address = sys.argv[10]
profile = sys.argv[11]

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

def replace_root_int(text: str, key: str, value: int) -> str:
    pattern = rf'(^\s*{re.escape(key)}\s*=\s*)[0-9]+'
    updated, count = re.subn(pattern, lambda item: f'{item.group(1)}{value}', text, count=1, flags=re.MULTILINE)
    if count != 1:
        raise SystemExit(f"failed to update root int {key}")
    return updated

def replace_root_array(text: str, key: str, value_literal: str) -> str:
    pattern = rf'(^\s*{re.escape(key)}\s*=\s*)\[[^\n]*\]'
    updated, count = re.subn(pattern, lambda item: f'{item.group(1)}{value_literal}', text, count=1, flags=re.MULTILINE)
    if count != 1:
        raise SystemExit(f"failed to update root array {key}")
    return updated

def replace_section_value_str(text: str, section: str, key: str, value: str) -> str:
    return replace_section_value(text, section, key, value)

def replace_section_int(text: str, section: str, key: str, value: int) -> str:
    pattern = rf'(\[{re.escape(section)}\][\s\S]*?^\s*{re.escape(key)}\s*=\s*)[0-9]+'
    updated, count = re.subn(pattern, lambda item: f'{item.group(1)}{value}', text, count=1, flags=re.MULTILINE)
    if count != 1:
        raise SystemExit(f"failed to update [{section}] int {key}")
    return updated

app_text = app_path.read_text(encoding="utf-8")
app_text = replace_root_value(app_text, "minimum-gas-prices", minimum_gas_prices)
app_text = replace_section_int(app_text, "state-sync", "snapshot-interval", 0)
app_text = replace_section_int(app_text, "state-sync", "snapshot-keep-recent", 2)
app_text = replace_section_bool(app_text, "json-rpc", "enable-indexer", False)

if profile == "rpc-30d":
    app_text = replace_root_value(app_text, "pruning", "custom")
    app_text = replace_root_value(app_text, "pruning-keep-recent", "518400")
    app_text = replace_root_value(app_text, "pruning-interval", "10")
    app_text = replace_root_int(app_text, "min-retain-blocks", 518400)
    app_text = replace_section_bool(app_text, "api", "enable", True)
    app_text = replace_section_bool(app_text, "api", "swagger", False)
    app_text = replace_section_value_str(app_text, "api", "address", api_address)
    app_text = replace_section_bool(app_text, "grpc", "enable", True)
    app_text = replace_section_value_str(app_text, "grpc", "address", grpc_address)
    app_text = replace_section_bool(app_text, "grpc-web", "enable", True)
    app_text = replace_section_bool(app_text, "json-rpc", "enable", True)
    app_text = replace_section_value_str(app_text, "json-rpc", "address", json_rpc_address)
    app_text = replace_section_value_str(app_text, "json-rpc", "ws-address", json_rpc_ws_address)
elif profile == "archive":
    app_text = replace_root_value(app_text, "pruning", "nothing")
    app_text = replace_root_int(app_text, "min-retain-blocks", 0)
    app_text = replace_section_bool(app_text, "api", "enable", True)
    app_text = replace_section_bool(app_text, "api", "swagger", False)
    app_text = replace_section_value_str(app_text, "api", "address", api_address)
    app_text = replace_section_bool(app_text, "grpc", "enable", True)
    app_text = replace_section_value_str(app_text, "grpc", "address", grpc_address)
    app_text = replace_section_bool(app_text, "grpc-web", "enable", True)
    app_text = replace_section_bool(app_text, "json-rpc", "enable", True)
    app_text = replace_section_value_str(app_text, "json-rpc", "address", json_rpc_address)
    app_text = replace_section_value_str(app_text, "json-rpc", "ws-address", json_rpc_ws_address)
elif profile == "p2p":
    app_text = replace_root_value(app_text, "pruning", "everything")
    app_text = replace_root_value(app_text, "pruning-keep-recent", "0")
    app_text = replace_root_value(app_text, "pruning-interval", "10")
    app_text = replace_root_int(app_text, "min-retain-blocks", 0)
    app_text = replace_section_bool(app_text, "api", "enable", False)
    app_text = replace_section_bool(app_text, "grpc", "enable", False)
    app_text = replace_section_bool(app_text, "grpc-web", "enable", False)
    app_text = replace_section_bool(app_text, "json-rpc", "enable", False)
else:
    raise SystemExit(f"unsupported sync profile: {profile}")
app_path.write_text(app_text, encoding="utf-8")

config_text = config_path.read_text(encoding="utf-8")
config_text = replace_root_value(config_text, "external_address", external_address)
config_text = replace_root_value(config_text, "persistent_peers", persistent_peers)
config_text = replace_root_value(config_text, "laddr", rpc_laddr)
config_text = replace_section_bool(config_text, "storage", "discard_abci_responses", profile == "p2p")
config_text = replace_section_value(config_text, "tx_index", "indexer", "null" if profile == "p2p" else "kv")
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
        --rpc.laddr "$RPC_LADDR"
    )

    if [ -n "$P2P_EXTERNAL_ADDRESS" ]; then
        args+=(--p2p.external-address "$P2P_EXTERNAL_ADDRESS")
    fi

    exec "$BINARY" "${args[@]}"
}

usage() {
    cat <<'EOF'
Manage a sync node from the current directory.

Commands:
  start   initialize local state if needed and start the sync node
  status  show process, PID, home, and log paths
  stop    stop the locally started sync node process
  help    show this help message

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
  - set SYNC_NODE_PROFILE=rpc-30d (default), archive, or p2p
EOF
}

command_start() {
    need_cmd python3
    SYNC_NODE_PROFILE="$(normalize_sync_node_profile "$SYNC_NODE_PROFILE")"
    case "$SYNC_NODE_PROFILE" in
        rpc-30d)
            PROFILE_LABEL="rpc-30d"
            RPC_LADDR="tcp://0.0.0.0:${RPC_PORT}"
            ;;
        archive)
            PROFILE_LABEL="archive"
            RPC_LADDR="tcp://0.0.0.0:${RPC_PORT}"
            ;;
        p2p)
            PROFILE_LABEL="p2p"
            RPC_LADDR="tcp://127.0.0.1:${RPC_PORT}"
            ;;
    esac
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
    echo "  Profile:   $PROFILE_LABEL"
    echo "  Peer:      $(cat "$PEER_INFO_FILE")"
    echo "  Upstream:  $(bootstrap_peers_value)"
    if [ "$SYNC_NODE_PROFILE" = "p2p" ]; then
        echo "  Services:  external RPC/API disabled"
    else
        echo "  Services:  RPC/API enabled"
    fi
    echo

    stop_existing_node
    mkdir -p "$(dirname "$LOG_FILE")"
    echo "$$" >"$PID_FILE"
    exec > >(tee -a "$LOG_FILE") 2>&1
    start_node
}

COMMAND="${1:-start}"

case "$COMMAND" in
    start)
        command_start
        ;;
    status)
        command_status
        ;;
    stop)
        command_stop
        ;;
    help|-h|--help)
        usage
        ;;
    *)
        usage
        exit 1
        ;;
esac
