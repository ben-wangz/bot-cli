# Capability Catalog

## Core Capabilities

- `ensure_daemon_started`
  - idempotently probe/start local `aria2c`
- `get_global_stat`
  - inspect aggregate daemon status
- `get_global_option`
  - inspect daemon global options
- `change_global_option`
  - update daemon global options
- `tell_status`
  - inspect one download by `gid`
- `list_active`
  - list active downloads
- `list_waiting`
  - list waiting downloads
- `list_stopped`
  - list stopped/completed/error downloads

## Download Creation

- `add_uri`
- `add_magnet`
- `add_torrent`
- `add_metalink`

These enqueue new downloads and usually return a `gid`.

## Download Control

- `pause`
- `pause_all`
- `resume`
- `resume_all`
- `remove`
- `remove_all`
- `purge_download_result`

## Escape Hatch

- `rpc_call`
  - call raw aria2 JSON-RPC when no first-class capability fits

## Workflows

- `queue_add_and_wait`
  - add one URI and poll until terminal state or timeout
- `pause_resume_chain`
  - pause, unpause, then return current status
- `cleanup_completed`
  - remove stored stopped/completed results in bulk

## Notes

1. Use `capability describe [<name>]` for runtime argument details.
2. Use global `--wait` flags for mutating capabilities when you need stable follow-up state.
3. Prefer workflows only when their built-in chain already matches the task.
