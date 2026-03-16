#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/lib/docker_go.sh"

VERSION="${VERSION:-v1.0.0}"
GOOS_TARGET="${GOOS_TARGET:-$(packaging_host_goos)}"
GOARCH_TARGET="${GOARCH_TARGET:-$(packaging_host_goarch)}"
OUT_DIR="$REPO_ROOT/dist"
SKIP_BUILD=false
BINARY_PATH=""
AXOND_CGO_ENABLED="${AXOND_CGO_ENABLED:-}"
BUILD_MODE="docker"

version_plain() {
  printf '%s\n' "${1#v}"
}

binary_name() {
  if [[ "$1" == "windows" ]]; then
    printf 'axond.exe\n'
  else
    printf 'axond\n'
  fi
}

usage() {
  cat <<EOF
Usage:
  bash packaging/package_axond.sh [options]

Options:
  --version <v>     package version (default: ${VERSION})
  --os <goos>       target GOOS (default: ${GOOS_TARGET})
  --arch <goarch>   target GOARCH (default: ${GOARCH_TARGET})
  --out <dir>       output directory (default: ${OUT_DIR})
  --binary <path>   use prebuilt axond binary instead of building
  --skip-build      skip binary build step
  --axond-cgo <v>   set CGO_ENABLED explicitly for axond builds
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
    --binary) BINARY_PATH="$2"; shift 2 ;;
    --skip-build) SKIP_BUILD=true; shift ;;
    --axond-cgo) AXOND_CGO_ENABLED="$2"; shift 2 ;;
    --docker-image) PACKAGING_DOCKER_IMAGE="$2"; shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

VERSION_PLAIN="$(version_plain "$VERSION")"
DIST_NAME="axond_${VERSION_PLAIN}_${GOOS_TARGET}_${GOARCH_TARGET}"
STAGE_DIR="$OUT_DIR/$DIST_NAME"
PACKAGED_BINARY_NAME="$(binary_name "$GOOS_TARGET")"
PACKAGED_BINARY_PATH="$STAGE_DIR/$PACKAGED_BINARY_NAME"
ARCHIVE_PATH="$OUT_DIR/${DIST_NAME}.tar.gz"
CACHE_BIN_DIR="$REPO_ROOT/.cache/packaging-bin/axond-${GOOS_TARGET}-${GOARCH_TARGET}"
CACHE_BIN_PATH="$CACHE_BIN_DIR/$PACKAGED_BINARY_NAME"

mkdir -p "$OUT_DIR" "$CACHE_BIN_DIR"

echo "Packaging axond binary..."
echo "  Version: ${VERSION}"
echo "  Target:  ${GOOS_TARGET}/${GOARCH_TARGET}"
echo "  Output:  ${ARCHIVE_PATH}"
echo "  Builder: ${BUILD_MODE} (${PACKAGING_DOCKER_IMAGE})"
if [[ -n "$BINARY_PATH" ]]; then
  echo "  axond:   prebuilt binary ${BINARY_PATH}"
elif [[ "$SKIP_BUILD" == true ]]; then
  echo "  axond:   skip build, expect existing binary in stage directory"
elif [[ -n "$AXOND_CGO_ENABLED" ]]; then
  echo "  axond:   build from source (CGO_ENABLED=${AXOND_CGO_ENABLED})"
else
  echo "  axond:   build from source (CGO_ENABLED=auto)"
fi
echo

rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR"

COMMIT="$(git -C "$REPO_ROOT" log -1 --format='%H' 2>/dev/null || echo unknown)"
LDFLAGS="-X github.com/cosmos/cosmos-sdk/version.Name=axon \
-X github.com/cosmos/cosmos-sdk/version.AppName=axond \
-X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
-X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT}"

if [[ -n "$BINARY_PATH" ]]; then
  if [[ ! -x "$BINARY_PATH" ]]; then
    echo "Provided binary is not executable: $BINARY_PATH" >&2
    exit 1
  fi
  cp "$BINARY_PATH" "$PACKAGED_BINARY_PATH"
elif [[ "$SKIP_BUILD" != true ]]; then
  echo "Building axond for ${GOOS_TARGET}/${GOARCH_TARGET}..."
  rm -f "$CACHE_BIN_PATH"
  if [[ -n "$AXOND_CGO_ENABLED" ]]; then
    packaging_run_go_container \
      "$GOOS_TARGET" \
      "$GOARCH_TARGET" \
      "." \
      "AXON_BUILD_LDFLAGS=$LDFLAGS" \
      "AXON_BUILD_OUTPUT=${PACKAGING_WORKSPACE}/$(packaging_relpath "$CACHE_BIN_PATH")" \
      "CGO_ENABLED=$AXOND_CGO_ENABLED" \
      -- \
      'go build -trimpath -mod=readonly -ldflags "-s -w $AXON_BUILD_LDFLAGS" -o "$AXON_BUILD_OUTPUT" ./cmd/axond'
  else
    packaging_run_go_container \
      "$GOOS_TARGET" \
      "$GOARCH_TARGET" \
      "." \
      "AXON_BUILD_LDFLAGS=$LDFLAGS" \
      "AXON_BUILD_OUTPUT=${PACKAGING_WORKSPACE}/$(packaging_relpath "$CACHE_BIN_PATH")" \
      -- \
      'go build -trimpath -mod=readonly -ldflags "-s -w $AXON_BUILD_LDFLAGS" -o "$AXON_BUILD_OUTPUT" ./cmd/axond'
  fi
  cp "$CACHE_BIN_PATH" "$PACKAGED_BINARY_PATH"
fi

if [[ ! -x "$PACKAGED_BINARY_PATH" ]]; then
  echo "axond binary not found in $PACKAGED_BINARY_PATH" >&2
  exit 1
fi

cp "$REPO_ROOT/scripts/start_validator_node.sh" "$STAGE_DIR/start_validator_node.sh"
cp "$REPO_ROOT/scripts/start_sync_node.sh" "$STAGE_DIR/start_sync_node.sh"
cp "$REPO_ROOT/docs/mainnet/genesis.json" "$STAGE_DIR/genesis.json"
cp "$REPO_ROOT/docs/mainnet/bootstrap_peers.txt" "$STAGE_DIR/bootstrap_peers.txt"
chmod 0755 "$STAGE_DIR/start_validator_node.sh" "$STAGE_DIR/start_sync_node.sh"

(
  cd "$OUT_DIR"
  tar -czf "$ARCHIVE_PATH" "$DIST_NAME"
)

echo "axond package created:"
echo "  $ARCHIVE_PATH"
