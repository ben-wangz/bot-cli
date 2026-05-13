---
name: aria2-cli
description: |
  Operate aria2c via aria2-cli using JSON-RPC capabilities and workflows with
  deterministic JSON output for agent-driven download tasks.
license: MIT
compatibility: opencode
metadata:
  audience: coding-agents
  tool: aria2-cli
---

# aria2-cli Skill

Use this skill when an agent needs reliable, script-friendly download control against an `aria2c` daemon.

## What aria2-cli Provides

`aria2-cli` supports both capability and workflow entry points:

1. `capability ...`: run one atomic JSON-RPC operation.
2. `workflow ...`: run a small multi-step chain for common download tasks.
3. `capability describe [<name>]`: inspect capability contract and args.

Key behaviors:

- Emits deterministic JSON envelopes for machine parsing.
- Can start a local daemon idempotently with `ensure_daemon_started`.
- Supports wait/polling for mutating capabilities via global `--wait` flags.
- Supports common queue, pause/resume, query, and cleanup operations.

## Operating Principles for Agents

1. Always use `--output json` for deterministic parsing.
2. When daemon state is unknown, call `capability ensure_daemon_started` first.
3. For mutating capabilities that return `gid`, use `--wait` when you need stable follow-up state.
4. Prefer workflows only when they match the desired chain exactly; otherwise compose capabilities directly.

## Base Command Template

Prefer resolved binary path for repeatable runs:

```bash
ARIA2_CLI_BIN="${ARIA2_CLI_BIN:-$(command -v aria2-cli 2>/dev/null || true)}"

if [ -z "${ARIA2_CLI_BIN}" ]; then
  echo "aria2-cli not found. Resolve/download binary first." >&2
  echo "See: references/binary-bootstrap-and-release-download.md" >&2
  exit 1
fi

"${ARIA2_CLI_BIN}" \
  --rpc-endpoint "http://127.0.0.1:6800/jsonrpc" \
  --rpc-secret "$ARIA2_RPC_SECRET" \
  --output json \
  capability get_global_stat
```

Fallback for source-only environments:

```bash
cd "<aria2-cli-source-root>/src"
go run ./cmd/aria2-cli ...
```

## Quick Task Routing

- Need binary bootstrap + release download pattern? -> [Binary bootstrap and release download](references/binary-bootstrap-and-release-download.md)
- Need minimal first-run daemon + download flow? -> [Quickstart](references/quickstart.md)
- Need capability/workflow overview? -> [Capability catalog](references/capability-catalog.md)
- Need failure handling for daemon start, RPC auth, wait behavior, or localhost constraints? -> [Troubleshooting](references/troubleshooting.md)

## Safety Checklist

Before executing:

1. `--rpc-endpoint` points at the intended daemon.
2. `ARIA2_RPC_SECRET` or `--rpc-secret` is set when daemon auth is enabled.
3. Output paths and download directories are deterministic.
4. Every step checks `ok == true` in JSON.
