# A49 ssh_tunnel_start

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A49 `ssh_tunnel_start` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Resolve template/node, allocate `TEST_VMID`, clone template, and start VM.
4) Resolve `GUEST_IP` via `agent_network_get_interfaces`.
5) Generate key pair and inject public key via `ssh_inject_pubkey_qga`.
6) Select `LOCAL_PORT` (unused local port) and set:
   - `PID_FILE=build/ssh-tunnels/phase4-a49.pid`
   - `LOG_FILE=build/ssh-tunnels/phase4-a49.log`

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_tunnel_start --host "$GUEST_IP" --port 22 --user ubuntu --identity-file "build/phase4-a49-id_ed25519" --local-port "$LOCAL_PORT" --remote-host 127.0.0.1 --remote-port 22 --pid-file "$PID_FILE" --log-file "$LOG_FILE"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_tunnel_start"`
- JSON field `ok == true`
- JSON contains `result.pid`
- JSON contains `result.pid_file` and `result.log_file`
- `PID_FILE` exists after command

Teardown:
- Stop tunnel via `ssh_tunnel_stop --pid-file "$PID_FILE"`.
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.
- Delete temporary key files and tunnel artifacts created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse pid/log file paths from other prompt runs.
```
