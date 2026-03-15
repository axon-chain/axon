#!/bin/bash
set -euo pipefail

PACKAGING_DOCKER_IMAGE="${PACKAGING_DOCKER_IMAGE:-golang:1.25.7-trixie}"
PACKAGING_WORKSPACE="/workspace"

packaging_need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

packaging_need_docker() {
  packaging_need_cmd docker
}

packaging_host_goos() {
  case "$(uname -s)" in
    Linux) printf 'linux\n' ;;
    Darwin) printf 'darwin\n' ;;
    *)
      echo "unsupported host OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

packaging_host_goarch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64\n' ;;
    aarch64|arm64) printf 'arm64\n' ;;
    *)
      echo "unsupported host architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

packaging_prepare_go_env() {
  mkdir -p \
    "$REPO_ROOT/.cache/go-build" \
    "$REPO_ROOT/.cache/gomod" \
    "$REPO_ROOT/.cache/packaging-bin"
}

packaging_relpath() {
  local input_path="$1"
  local abs_path=""

  abs_path="$(cd "$(dirname "$input_path")" && pwd)/$(basename "$input_path")"

  case "$abs_path" in
    "$REPO_ROOT")
      printf '.\n'
      ;;
    "$REPO_ROOT"/*)
      printf '%s\n' "${abs_path#$REPO_ROOT/}"
      ;;
    *)
      echo "path must be inside repository: $input_path" >&2
      exit 1
      ;;
  esac
}

packaging_linux_platform() {
  case "$1" in
    amd64) printf 'linux/amd64\n' ;;
    arm64) printf 'linux/arm64\n' ;;
    *)
      echo "unsupported Linux architecture for Docker build: $1" >&2
      exit 1
      ;;
  esac
}

packaging_assert_docker_target() {
  local goos="$1"
  local goarch="$2"

  if [[ "$goos" != "linux" ]]; then
    echo "dockerized source builds currently support linux targets only: ${goos}/${goarch}" >&2
    echo "use a prebuilt binary if you need to package a non-linux target" >&2
    exit 1
  fi

  packaging_linux_platform "$goarch" >/dev/null
}

packaging_run_go_container() {
  local goos="$1"
  local goarch="$2"
  local work_rel="$3"
  shift 3

  local docker_envs=()
  local command=""
  local platform=""
  local workdir="$PACKAGING_WORKSPACE"

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --)
        shift
        break
        ;;
      *=*)
        docker_envs+=(-e "$1")
        shift
        ;;
      *)
        echo "invalid docker environment argument: $1" >&2
        exit 1
        ;;
    esac
  done

  if [[ $# -ne 1 ]]; then
    echo "packaging_run_go_container expects exactly one shell command" >&2
    exit 1
  fi

  command="$1"
  packaging_assert_docker_target "$goos" "$goarch"
  packaging_need_docker
  packaging_prepare_go_env

  platform="$(packaging_linux_platform "$goarch")"
  if [[ -n "$work_rel" && "$work_rel" != "." ]]; then
    workdir="$PACKAGING_WORKSPACE/$work_rel"
  fi

  docker run --rm \
    --platform "$platform" \
    --user "$(id -u):$(id -g)" \
    -e HOME=/tmp \
    -e GOPATH=/tmp/go \
    -e GOOS="$goos" \
    -e GOARCH="$goarch" \
    -e GOCACHE="$PACKAGING_WORKSPACE/.cache/go-build" \
    -e GOMODCACHE="$PACKAGING_WORKSPACE/.cache/gomod" \
    "${docker_envs[@]}" \
    -v "$REPO_ROOT:$PACKAGING_WORKSPACE" \
    -w "$workdir" \
    "$PACKAGING_DOCKER_IMAGE" \
    bash -c "$command"
}
