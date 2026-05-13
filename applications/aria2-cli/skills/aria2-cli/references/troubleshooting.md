# Troubleshooting

## Daemon Will Not Start

Check `ensure_daemon_started` first.

Known constraint:

- `ensure_daemon_started` only supports localhost endpoints.

If you target a remote daemon, start or manage that daemon outside `aria2-cli`, then use the normal RPC capabilities.

## RPC Authentication Failures

If RPC returns auth errors:

1. Verify `--rpc-secret` or `ARIA2_RPC_SECRET` matches the daemon config.
2. Remember `--rpc-secret` overrides `ARIA2_RPC_SECRET`.

## Mutating Call Returned Before State Settled

Some capabilities return before download state becomes stable.

Use:

- `--wait`
- `--wait-timeout <sec>`
- `--wait-interval <ms>`

This is most useful for `add_*`, `pause`, `resume`, and `remove` style calls.

## Workflow Timed Out

`workflow queue_add_and_wait` reports timeout via diagnostics.

If that happens:

1. Increase `--wait_timeout` for the workflow.
2. Re-check with `capability tell_status --gid <gid>`.
3. Inspect daemon health with `capability get_global_stat`.

## Need Unsupported RPC Surface

Use `capability rpc_call` when a first-class capability is missing, but prefer named capabilities for common operations so output stays predictable.
