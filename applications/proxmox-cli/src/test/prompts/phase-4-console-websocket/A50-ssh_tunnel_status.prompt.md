# A50 ssh_tunnel_status

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase4-ssh-control-plane.shared-tunnel-pid-file` exists.

## Prompt

```text
You are a test execution agent. Run the A50 `ssh_tunnel_status` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `PID_FILE` from `build/phase4-ssh-control-plane.shared-tunnel-pid-file`.
3) Validate `PID_FILE` exists before command execution.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_tunnel_status --pid-file "$PID_FILE"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_tunnel_status"`
- JSON field `ok == true`
- JSON contains `result.running == true`
- JSON contains `result.pid`

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs in this run.
- This prompt depends on tunnel artifacts created by `A49` in the same suite run.
```
