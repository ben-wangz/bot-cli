#!/usr/bin/env bash

set -euo pipefail

SCRIPT_PATH="$(readlink -f "$0")"
SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"
APP_DIR="$SCRIPT_DIR"
SRC_DIR="$APP_DIR/src"
OUT_DIR="$APP_DIR/build/bin"
OUT_BIN="$OUT_DIR/proxmox-cli"

mkdir -p "$OUT_DIR"

(cd "$SRC_DIR" && go build -o "$OUT_BIN" ./cmd/proxmox-cli)

echo "Built: $OUT_BIN"
