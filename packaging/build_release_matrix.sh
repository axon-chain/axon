#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/lib/docker_go.sh"

VERSION="${VERSION:-v1.0.0}"
OUT_DIR="${OUT_DIR:-$REPO_ROOT/dist/releases}"
TARGETS="${TARGETS:-linux/amd64 linux/arm64}"
PROJECT_NAME="axon"
RELEASE_NOTES_FILE="${RELEASE_NOTES_FILE:-}"
ALLOW_PARTIAL="${ALLOW_PARTIAL:-0}"
AXOND_CGO_ENABLED="${AXOND_CGO_ENABLED:-}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

normalize_version() {
  if [[ "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
    if [[ "$VERSION" == v* ]]; then
      printf '%s\n' "$VERSION"
    else
      printf 'v%s\n' "$VERSION"
    fi
    return
  fi

  echo "invalid version format: $VERSION" >&2
  echo "expected semantic version such as v1.0.0 or 1.0.0" >&2
  exit 1
}

write_checksums() {
  local target_dir="$1"
  local sums_file="$target_dir/SHA256SUMS"

  : >"$sums_file"
  (
    cd "$target_dir"
    for artifact in *.tar.gz *.zip; do
      if [[ ! -f "$artifact" ]]; then
        continue
      fi

      if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$artifact" >>"$sums_file"
      else
        shasum -a 256 "$artifact" >>"$sums_file"
      fi
    done
  )
}

usage() {
  cat <<EOF
Usage:
  bash packaging/build_release_matrix.sh [options]

Options:
  --version <v>      release version tag (default: ${VERSION})
  --out <dir>        output directory (default: ${OUT_DIR})
  --targets <list>   space-separated GOOS/GOARCH list
  --notes <file>     optional release notes file to include
  --allow-partial    keep successful artifacts even if some targets fail
  --axond-cgo <v>    set CGO_ENABLED explicitly for axond builds
  --docker-image <image>  Go builder image (default: ${PACKAGING_DOCKER_IMAGE})
  --help             show this help

Examples:
  bash packaging/build_release_matrix.sh --version v1.0.0
  bash packaging/build_release_matrix.sh --targets "linux/amd64 linux/arm64"
  bash packaging/build_release_matrix.sh --allow-partial
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --targets) TARGETS="$2"; shift 2 ;;
    --notes) RELEASE_NOTES_FILE="$2"; shift 2 ;;
    --allow-partial) ALLOW_PARTIAL=1; shift ;;
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

need_cmd tar
packaging_need_docker
VERSION_TAG="$(normalize_version)"
RELEASE_DIR="$OUT_DIR/$VERSION_TAG"
REPORT_FILE="$RELEASE_DIR/BUILD_REPORT.md"
LOG_DIR="$RELEASE_DIR/.logs"
SUCCESS_COUNT=0
FAIL_COUNT=0

if [[ -n "$RELEASE_NOTES_FILE" && ! -f "$RELEASE_NOTES_FILE" ]]; then
  echo "release notes file does not exist: $RELEASE_NOTES_FILE" >&2
  exit 1
fi

rm -rf "$LOG_DIR"
mkdir -p "$RELEASE_DIR" "$LOG_DIR"

if [[ -n "$RELEASE_NOTES_FILE" ]]; then
  cp "$RELEASE_NOTES_FILE" "$RELEASE_DIR/RELEASE_NOTES.md"
fi

cat >"$REPORT_FILE" <<EOF
# Release Build Report

- Project: ${PROJECT_NAME}
- Version: ${VERSION_TAG}
- Targets: ${TARGETS}
- Allow Partial: ${ALLOW_PARTIAL}
- Axond CGO_ENABLED override: ${AXOND_CGO_ENABLED:-auto}
- Builder Image: ${PACKAGING_DOCKER_IMAGE}

## Results

EOF

echo "Building release matrix for ${VERSION_TAG}"
echo "Output directory: $RELEASE_DIR"
echo "Builder image: $PACKAGING_DOCKER_IMAGE"

for target in $TARGETS; do
  goos="${target%/*}"
  goarch="${target#*/}"

  if [[ -z "$goos" || -z "$goarch" || "$goos" == "$goarch" ]]; then
    echo "invalid target: $target" >&2
    exit 1
  fi

  echo "==> Packaging axond for ${goos}/${goarch}"
  printf -- "- axond %s/%s: " "$goos" "$goarch" >>"$REPORT_FILE"
  axond_log="$LOG_DIR/axond_${goos}_${goarch}.log"
  axond_cmd=(
    bash "$SCRIPT_DIR/package_axond.sh"
    --version "$VERSION_TAG"
    --os "$goos"
    --arch "$goarch"
    --out "$RELEASE_DIR"
    --docker-image "$PACKAGING_DOCKER_IMAGE"
  )
  if [[ -n "$AXOND_CGO_ENABLED" ]]; then
    axond_cmd+=(--axond-cgo "$AXOND_CGO_ENABLED")
  fi
  if "${axond_cmd[@]}" >"$axond_log" 2>&1; then
    echo "success"
    echo "success" >>"$REPORT_FILE"
    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
  else
    echo "failed"
    echo "failed" >>"$REPORT_FILE"
    echo >>"$REPORT_FILE"
    echo '```text' >>"$REPORT_FILE"
    cat "$axond_log" >>"$REPORT_FILE"
    echo '```' >>"$REPORT_FILE"
    echo >>"$REPORT_FILE"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    if [[ "$ALLOW_PARTIAL" -ne 1 ]]; then
      write_checksums "$RELEASE_DIR"
      exit 1
    fi
  fi

  echo "==> Packaging agent-daemon for ${goos}/${goarch}"
  printf -- "- agent-daemon %s/%s: " "$goos" "$goarch" >>"$REPORT_FILE"
  agent_log="$LOG_DIR/agent-daemon_${goos}_${goarch}.log"
  agent_cmd=(
    bash "$SCRIPT_DIR/package_agent.sh"
    --version "$VERSION_TAG"
    --os "$goos"
    --arch "$goarch"
    --out "$RELEASE_DIR"
    --docker-image "$PACKAGING_DOCKER_IMAGE"
  )
  if "${agent_cmd[@]}" >"$agent_log" 2>&1; then
    echo "success"
    echo "success" >>"$REPORT_FILE"
    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
  else
    echo "failed"
    echo "failed" >>"$REPORT_FILE"
    echo >>"$REPORT_FILE"
    echo '```text' >>"$REPORT_FILE"
    cat "$agent_log" >>"$REPORT_FILE"
    echo '```' >>"$REPORT_FILE"
    echo >>"$REPORT_FILE"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    if [[ "$ALLOW_PARTIAL" -ne 1 ]]; then
      write_checksums "$RELEASE_DIR"
      exit 1
    fi
  fi
done

write_checksums "$RELEASE_DIR"
rm -rf "$LOG_DIR"

echo >>"$REPORT_FILE"
echo "## Summary" >>"$REPORT_FILE"
echo >>"$REPORT_FILE"
echo "- Successful artifacts: ${SUCCESS_COUNT}" >>"$REPORT_FILE"
echo "- Failed artifacts: ${FAIL_COUNT}" >>"$REPORT_FILE"

echo "Release artifacts created:"
find "$RELEASE_DIR" -maxdepth 1 -type f | sort

if [[ "$SUCCESS_COUNT" -eq 0 ]]; then
  echo "no release artifacts were created successfully" >&2
  exit 1
fi

if [[ "$FAIL_COUNT" -gt 0 && "$ALLOW_PARTIAL" -ne 1 ]]; then
  echo "release matrix completed with failures; rerun with --allow-partial to keep partial results" >&2
  exit 1
fi
