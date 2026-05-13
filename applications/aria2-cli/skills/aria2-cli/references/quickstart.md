# Quickstart

Use this sequence for a minimal first-run validation.

## Prerequisites

- `aria2-cli` binary is available locally.
- Local RPC endpoint: `http://127.0.0.1:6800/jsonrpc`
- If daemon auth is enabled, set `ARIA2_RPC_SECRET` or pass `--rpc-secret`.

## Minimal Startup Check

```bash
ARIA2_CLI_BIN="${ARIA2_CLI_BIN:-$(command -v aria2-cli 2>/dev/null || true)}"

"${ARIA2_CLI_BIN}" \
  --rpc-endpoint "http://127.0.0.1:6800/jsonrpc" \
  --output json \
  capability ensure_daemon_started
```

Then verify daemon state:

```bash
"${ARIA2_CLI_BIN}" \
  --rpc-endpoint "http://127.0.0.1:6800/jsonrpc" \
  --output json \
  capability get_global_stat
```

## Minimal Download Flow

```bash
"${ARIA2_CLI_BIN}" \
  --rpc-endpoint "http://127.0.0.1:6800/jsonrpc" \
  --output json \
  workflow queue_add_and_wait \
  --uri "https://github.com/ben-wangz/bot-cli/releases/download/v0.1.0/aria2-cli-linux-amd64"
```

Expected checks:

1. `ok == true`
2. `result.gid` is non-empty
3. `result.final_status` is one of `complete`, `error`, or `removed`
