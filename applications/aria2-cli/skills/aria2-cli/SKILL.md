---
name: aria2-cli
description: Use aria2-cli capabilities and workflows to operate aria2c via JSON-RPC.
---

# aria2-cli skill

## Purpose

Use `applications/aria2-cli/src/cmd/aria2-cli` to run deterministic, JSON-first download operations.

## Principles

- Prefer `capability` for atomic operations.
- Use `workflow` for multi-step execution.
- Keep outputs in JSON and pass structured args.
- Always call `ensure_daemon_started` before download operations when daemon state is unknown.

## Common commands

- `capability ensure_daemon_started`
- `capability get_global_stat`
- `capability list_active`
- `capability add_uri --uri <url>`
- `capability pause --gid <gid>`
- `capability resume --gid <gid>`
- `capability remove --gid <gid>`
- `workflow queue_add_and_wait --uri <url>`

## Recommended startup sequence

1. `capability ensure_daemon_started`
2. `capability get_global_stat`
3. run task capabilities/workflows

## Required globals

- `--rpc-endpoint`
- `--rpc-secret` (when daemon requires token), or env `ARIA2_RPC_SECRET`

## Secret priority

- command flag `--rpc-secret` has higher priority than env `ARIA2_RPC_SECRET`
