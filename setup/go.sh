#!/bin/bash

set -o errexit -o nounset -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

DEFAULT_GO_VERSION="${GO_VERSION:-1.26.2}"
DEFAULT_INSTALL_ROOT="${GO_INSTALL_ROOT:-${PROJECT_ROOT}/build/tools}"
DEFAULT_DOWNLOAD_BASE="${GO_DOWNLOAD_BASE:-https://go.dev/dl}"

function usage() {
    cat <<'EOF'
Usage: setup/go.sh [go-version] [install-root]

Installs Go by downloading the official tarball and extracting it as <install-root>/go.

Arguments:
  go-version   Go version, with or without leading v (default: GO_VERSION or 1.26.2)
  install-root Install root directory (default: GO_INSTALL_ROOT or <repo>/build/tools)

Environment:
  GO_VERSION         Default Go version
  GO_INSTALL_ROOT    Default install root directory
  GO_DOWNLOAD_BASE   Download base URL (default: https://go.dev/dl)

Output:
  Prints resolved Go binary path on stdout.
EOF
}

function normalize_version() {
    local raw="$1"

    raw="${raw#v}"

    if [[ ! "$raw" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        return 1
    fi

    printf '%s\n' "$raw"
}

function detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"

    case "$os" in
        linux|darwin)
            printf '%s\n' "$os"
            ;;
        *)
            echo "Error: Unsupported OS: ${os}" >&2
            return 1
            ;;
    esac
}

function detect_arch() {
    local arch
    arch="$(uname -m)"

    case "$arch" in
        x86_64|amd64)
            printf 'amd64\n'
            ;;
        aarch64|arm64)
            printf 'arm64\n'
            ;;
        *)
            echo "Error: Unsupported architecture: ${arch}" >&2
            return 1
            ;;
    esac
}

function read_go_version() {
    local go_bin="$1"
    local output=""

    if ! output="$(${go_bin} version 2>/dev/null)"; then
        return 1
    fi

    if [[ "$output" =~ go([0-9]+\.[0-9]+\.[0-9]+) ]]; then
        printf '%s\n' "${BASH_REMATCH[1]}"
        return 0
    fi

    return 1
}

function download_with_curl() {
    local url="$1"
    local output_path="$2"
    curl -fsSL --retry 3 --connect-timeout 10 --max-time 300 -o "$output_path" "$url"
}

function install_go() {
    local version="$1"
    local install_root="$2"

    local os
    local arch
    os="$(detect_os)"
    arch="$(detect_arch)"

    local file="go${version}.${os}-${arch}.tar.gz"
    local url="${DEFAULT_DOWNLOAD_BASE%/}/${file}"

    local tmp_dir
    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "${tmp_dir:-}"; trap - RETURN' RETURN

    if ! command -v curl >/dev/null 2>&1; then
        echo "Error: curl is required to download Go" >&2
        return 1
    fi

    download_with_curl "$url" "${tmp_dir}/${file}"

    mkdir -p "$install_root"
    rm -rf "${install_root}/go"
    tar -C "$install_root" -xzf "${tmp_dir}/${file}"

    local go_bin="${install_root}/go/bin/go"
    if [[ ! -x "$go_bin" ]]; then
        echo "Error: failed to install Go, binary not found at ${go_bin}" >&2
        return 1
    fi

    mkdir -p "${PROJECT_ROOT}/build/bin"
    ln -sf "${go_bin}" "${PROJECT_ROOT}/build/bin/go"
    ln -sf "${install_root}/go/bin/gofmt" "${PROJECT_ROOT}/build/bin/gofmt"

    local installed_version
    installed_version="$(read_go_version "$go_bin")" || {
        echo "Error: failed to read installed Go version" >&2
        return 1
    }

    if [[ "$installed_version" != "$version" ]]; then
        echo "Error: installed Go version (${installed_version}) does not match requested (${version})" >&2
        return 1
    fi

    printf '%s\n' "$go_bin"
}

function main() {
    if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
        usage
        exit 0
    fi

    if [[ $# -gt 2 ]]; then
        usage >&2
        exit 1
    fi

    local version_input="${1:-${DEFAULT_GO_VERSION}}"
    local install_root="${2:-${DEFAULT_INSTALL_ROOT}}"

    local version
    version="$(normalize_version "$version_input")" || {
        echo "Error: Invalid go version: ${version_input}" >&2
        exit 1
    }

    local existing_go="${install_root}/go/bin/go"
    if [[ -x "$existing_go" ]]; then
        local existing_version
        if existing_version="$(read_go_version "$existing_go")" && [[ "$existing_version" == "$version" ]]; then
            mkdir -p "${PROJECT_ROOT}/build/bin"
            ln -sf "${existing_go}" "${PROJECT_ROOT}/build/bin/go"
            ln -sf "${install_root}/go/bin/gofmt" "${PROJECT_ROOT}/build/bin/gofmt"
            printf '%s\n' "$existing_go"
            exit 0
        fi
    fi

    install_go "$version" "$install_root"
}

main "$@"
