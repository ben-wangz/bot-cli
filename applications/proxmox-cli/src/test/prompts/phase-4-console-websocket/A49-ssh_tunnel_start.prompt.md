# A49 ssh_tunnel_start

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase4-ssh-control-plane.shared-node` exists.
- `build/phase4-ssh-control-plane.shared-vmid` exists.
- `build/phase4-ssh-control-plane.shared-guest-ip` exists.
- `build/phase4-ssh-control-plane.shared-identity-file` exists.

## Prompt

```text
You are a test execution agent. Run the A49 `ssh_tunnel_start` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `SHARED_NODE` from `build/phase4-ssh-control-plane.shared-node`.
4) Read `SHARED_VMID` from `build/phase4-ssh-control-plane.shared-vmid`.
5) Read `SHARED_GUEST_IP` from `build/phase4-ssh-control-plane.shared-guest-ip`.
6) Read `SHARED_IDENTITY_FILE` from `build/phase4-ssh-control-plane.shared-identity-file`.
7) Select `LOCAL_PORT` (unused local port) and set:
   - `PID_FILE=build/ssh-tunnels/phase4-shared.pid`
   - `LOG_FILE=build/ssh-tunnels/phase4-shared.log`
8) Persist tunnel pid file path to `build/phase4-ssh-control-plane.shared-tunnel-pid-file`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_tunnel_start --host "$SHARED_GUEST_IP" --port 22 --user cloud --identity-file "$SHARED_IDENTITY_FILE" --local-port "$LOCAL_PORT" --remote-host 127.0.0.1 --remote-port 22 --pid-file "$PID_FILE" --log-file "$LOG_FILE"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_tunnel_start"`
- JSON field `ok == true`
- JSON contains `result.pid`
- JSON contains `result.pid_file` and `result.log_file`
- `PID_FILE` exists after command
- `build/phase4-ssh-control-plane.shared-tunnel-pid-file` exists and points to `PID_FILE`

Independence rule:
- This test consumes suite-level shared VM artifacts and must not create/stop/destroy VMs.
- Keep tunnel alive for downstream `A50`/`A51`; do not stop tunnel on success path.
```
