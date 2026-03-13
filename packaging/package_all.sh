#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/lib/docker_go.sh"

VERSION="${VERSION:-v1.0.0}"
GOOS_TARGET="${GOOS_TARGET:-$(packaging_host_goos)}"
GOARCH_TARGET="${GOARCH_TARGET:-$(packaging_host_goarch)}"
OUT_DIR="$REPO_ROOT/dist"
AXOND_BINARY="${AXOND_BINARY:-}"
AXOND_CGO_ENABLED="${AXOND_CGO_ENABLED:-}"

version_plain() {
  printf '%s\n' "${1#v}"
}

usage() {
  cat <<EOF
Usage:
  bash packaging/package_all.sh [options]

Options:
  --version <v>     package version (default: ${VERSION})
  --os <goos>       target GOOS (default: ${GOOS_TARGET})
  --arch <goarch>   target GOARCH (default: ${GOARCH_TARGET})
  --out <dir>       output directory (default: ${OUT_DIR})
  --axond-binary <path>  package prebuilt axond binary
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
    --axond-binary) AXOND_BINARY="$2"; shift 2 ;;
    --axond-cgo) AXOND_CGO_ENABLED="$2"; shift 2 ;;
    --docker-image) PACKAGING_DOCKER_IMAGE="$2"; shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *)
      echo "Unknown argument: $1"
      usage
      exit 1
      ;;
  esac
done

echo "Packaging Axon release bundle..."
echo "  Version: ${VERSION}"
echo "  Target:  ${GOOS_TARGET}/${GOARCH_TARGET}"
echo "  Output:  ${OUT_DIR}"
echo "  Builder: docker (${PACKAGING_DOCKER_IMAGE})"
if [[ -n "$AXOND_BINARY" ]]; then
  echo "  axond:   prebuilt binary ${AXOND_BINARY}"
elif [[ -n "$AXOND_CGO_ENABLED" ]]; then
  echo "  axond:   build from source (CGO_ENABLED=${AXOND_CGO_ENABLED})"
else
  echo "  axond:   build from source (CGO_ENABLED=auto)"
fi
echo

VERSION_PLAIN="$(version_plain "$VERSION")"
VALIDATOR_TARBALL="${OUT_DIR}/axond_${VERSION_PLAIN}_${GOOS_TARGET}_${GOARCH_TARGET}.tar.gz"
AGENT_TARBALL="${OUT_DIR}/agent-daemon_${VERSION_PLAIN}_${GOOS_TARGET}_${GOARCH_TARGET}.tar.gz"

echo "Packaging axond..."
if [[ -n "$AXOND_BINARY" ]]; then
  bash "$SCRIPT_DIR/package_axond.sh" \
    --version "$VERSION" \
    --os "$GOOS_TARGET" \
    --arch "$GOARCH_TARGET" \
    --out "$OUT_DIR" \
    --docker-image "$PACKAGING_DOCKER_IMAGE" \
    --axond-cgo "$AXOND_CGO_ENABLED" \
    --binary "$AXOND_BINARY"
else
  if [[ -n "$AXOND_CGO_ENABLED" ]]; then
    bash "$SCRIPT_DIR/package_axond.sh" \
      --version "$VERSION" \
      --os "$GOOS_TARGET" \
      --arch "$GOARCH_TARGET" \
      --out "$OUT_DIR" \
      --docker-image "$PACKAGING_DOCKER_IMAGE" \
      --axond-cgo "$AXOND_CGO_ENABLED"
  else
    bash "$SCRIPT_DIR/package_axond.sh" \
      --version "$VERSION" \
      --os "$GOOS_TARGET" \
      --arch "$GOARCH_TARGET" \
      --out "$OUT_DIR" \
      --docker-image "$PACKAGING_DOCKER_IMAGE"
  fi
fi

echo "Packaging agent-daemon..."
bash "$SCRIPT_DIR/package_agent.sh" \
  --version "$VERSION" \
  --os "$GOOS_TARGET" \
  --arch "$GOARCH_TARGET" \
  --out "$OUT_DIR" \
  --docker-image "$PACKAGING_DOCKER_IMAGE"

echo
echo "All packages created."
echo "  Axond:     $VALIDATOR_TARBALL"
echo "  Agent:     $AGENT_TARBALL"
echo "  Directory: $OUT_DIR"
