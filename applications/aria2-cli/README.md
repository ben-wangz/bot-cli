# aria2-cli

Agent-first CLI for operating an `aria2c` daemon via JSON-RPC.

## Scope

- capability-first command surface (`capability ...`)
- workflow wrappers for common chains (`workflow ...`)
- deterministic JSON output for agent consumption

## Quick Start

1. Build binary:

```bash
cd applications/aria2-cli/src
go build ./cmd/aria2-cli
```

2. Ensure daemon is started (idempotent, starts only once):

```bash
./aria2-cli capability ensure_daemon_started \
  --rpc-endpoint http://127.0.0.1:6800/jsonrpc \
  --rpc-secret your-secret
```

Or set env once:

```bash
export ARIA2_RPC_SECRET="your-secret"
./aria2-cli capability ensure_daemon_started --rpc-endpoint http://127.0.0.1:6800/jsonrpc
```

3. Run capability commands:

```bash
./aria2-cli capability get_global_stat --rpc-endpoint http://127.0.0.1:6800/jsonrpc --rpc-secret your-secret
```

4. For end-to-end validation, use prompt specs:

- `applications/aria2-cli/tests/prompts/basic-capability-chain.md`
- `applications/aria2-cli/tests/prompts/workflow-smoke.md`
- `applications/aria2-cli/tests/prompts/daemon-idempotent-with-secret.md`
- `applications/aria2-cli/tests/prompts/error-invalid-secret.md`
- `applications/aria2-cli/tests/prompts/error-invalid-endpoint.md`
- `applications/aria2-cli/tests/prompts/error-missing-gid.md`

## Core Commands

- `capability add_uri`
- `capability add_magnet`
- `capability add_torrent`
- `capability add_metalink`
- `capability ensure_daemon_started`
- `capability pause|pause_all|resume|resume_all|remove|remove_all`
- `capability tell_status|list_active|list_waiting|list_stopped|get_global_stat`
- `capability purge_download_result|rpc_call`

## Wait Behavior

- Global `--wait` is supported for mutating capabilities.
- Default is no wait; with `--wait`, CLI polls task status and writes wait diagnostics.
- Tune wait with `--wait-timeout <sec>` and `--wait-interval <ms>`.

## Daemon Idempotency Contract

- `ensure_daemon_started` first probes RPC with `aria2.getVersion`.
- If daemon is already running, it returns success with `already_running=true`.
- If daemon is not running, it starts `aria2c` in daemon mode and waits for RPC readiness.
- It must not start duplicate daemon processes for the same endpoint.

## Secret Input

- Supported sources: `--rpc-secret` and env `ARIA2_RPC_SECRET`
- Priority: `--rpc-secret` overrides `ARIA2_RPC_SECRET`

## Output Contract

Default output is JSON and follows this envelope:

```json
{
  "ok": true,
  "request": {},
  "result": {},
  "diagnostics": {}
}
```
