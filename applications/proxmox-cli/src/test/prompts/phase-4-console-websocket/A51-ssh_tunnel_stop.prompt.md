# A51 ssh_tunnel_stop

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase4-ssh-control-plane.shared-tunnel-pid-file` exists.

## Prompt

```text
You are a test execution agent. Run the A51 `ssh_tunnel_stop` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `PID_FILE` from `build/phase4-ssh-control-plane.shared-tunnel-pid-file`.
3) Validate `PID_FILE` exists before command execution.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_tunnel_stop --pid-file "$PID_FILE"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_tunnel_stop"`
- JSON field `ok == true`
- JSON contains `result.stopped == true`
- JSON contains `result.cleanup`
- `PID_FILE` no longer points to a running tunnel process

Teardown:
- Remove `build/phase4-ssh-control-plane.shared-tunnel-pid-file` after successful stop.

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs in this run.
- This prompt depends on tunnel artifacts created by `A49` in the same suite run.
```
