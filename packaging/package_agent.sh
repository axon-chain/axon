#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
AGENT_DIR="$REPO_ROOT/tools/agent-daemon"
source "$SCRIPT_DIR/lib/docker_go.sh"

VERSION="${VERSION:-v1.0.0}"
GOOS_TARGET="${GOOS_TARGET:-$(packaging_host_goos)}"
GOARCH_TARGET="${GOARCH_TARGET:-$(packaging_host_goarch)}"
OUT_DIR="$REPO_ROOT/dist"
SKIP_BUILD=false
BUILD_MODE="docker"

version_plain() {
  printf '%s\n' "${1#v}"
}

binary_name() {
  if [[ "$1" == "windows" ]]; then
    printf 'agent-daemon.exe\n'
  else
    printf 'agent-daemon\n'
  fi
}

usage() {
  cat <<EOF
Usage:
  bash packaging/package_agent.sh [options]

Options:
  --version <v>     package version (default: ${VERSION})
  --os <goos>       target GOOS (default: ${GOOS_TARGET})
  --arch <goarch>   target GOARCH (default: ${GOARCH_TARGET})
  --out <dir>       output directory (default: ${OUT_DIR})
  --skip-build      skip binary build step
  --docker-image <image>  Go builder image (default: ${PACKAGING_DOCKER_IMAGE})
  --help            show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --os) GOOS_TARGET="$2"; shift 2 ;;
    --arch) GOARCH_TARGET="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --skip-build) SKIP_BUILD=true; shift ;;
    --docker-image) PACKAGING_DOCKER_IMAGE="$2"; shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "$AGENT_DIR/main.go" ]]; then
  echo "agent-daemon source not found: $AGENT_DIR/main.go" >&2
  exit 1
fi

VERSION_PLAIN="$(version_plain "$VERSION")"
DIST_NAME="agent-daemon_${VERSION_PLAIN}_${GOOS_TARGET}_${GOARCH_TARGET}"
STAGE_DIR="$OUT_DIR/$DIST_NAME"
PACKAGED_BINARY_NAME="$(binary_name "$GOOS_TARGET")"
PACKAGED_BINARY_PATH="$STAGE_DIR/$PACKAGED_BINARY_NAME"
ARCHIVE_PATH="$OUT_DIR/${DIST_NAME}.tar.gz"
CACHE_BIN_DIR="$REPO_ROOT/.cache/packaging-bin/agent-${GOOS_TARGET}-${GOARCH_TARGET}"
CACHE_BIN_PATH="$CACHE_BIN_DIR/$PACKAGED_BINARY_NAME"

mkdir -p "$OUT_DIR" "$CACHE_BIN_DIR"

echo "Packaging agent-daemon binary..."
echo "  Version: ${VERSION}"
echo "  Target:  ${GOOS_TARGET}/${GOARCH_TARGET}"
echo "  Output:  ${ARCHIVE_PATH}"
echo "  Builder: ${BUILD_MODE} (${PACKAGING_DOCKER_IMAGE})"
if [[ "$SKIP_BUILD" == true ]]; then
  echo "  agent:   skip build, expect existing binary in stage directory"
else
  echo "  agent:   build from source"
fi
echo

rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR"

if [[ "$SKIP_BUILD" != true ]]; then
  echo "Building agent-daemon for ${GOOS_TARGET}/${GOARCH_TARGET}..."
  rm -f "$CACHE_BIN_PATH"
  packaging_run_go_container \
    "$GOOS_TARGET" \
    "$GOARCH_TARGET" \
    "tools/agent-daemon" \
    "AXON_BUILD_OUTPUT=${PACKAGING_WORKSPACE}/$(packaging_relpath "$CACHE_BIN_PATH")" \
    "CGO_ENABLED=0" \
    -- \
    'go build -trimpath -ldflags "-s -w" -o "$AXON_BUILD_OUTPUT" .'
  cp "$CACHE_BIN_PATH" "$PACKAGED_BINARY_PATH"
fi

if [[ ! -x "$PACKAGED_BINARY_PATH" ]]; then
  echo "agent-daemon binary not found in $PACKAGED_BINARY_PATH" >&2
  exit 1
fi

(
  cd "$OUT_DIR"
  tar -czf "$ARCHIVE_PATH" "$DIST_NAME"
)

echo "agent-daemon package created:"
echo "  $ARCHIVE_PATH"
