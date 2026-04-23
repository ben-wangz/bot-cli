# A51 ssh_tunnel_stop

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A51 `ssh_tunnel_stop` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Resolve template/node, allocate `TEST_VMID`, clone template, and start VM.
4) Resolve `GUEST_IP` via `agent_network_get_interfaces`.
5) Generate key pair and inject public key via `ssh_inject_pubkey_qga`.
6) Start tunnel first via `ssh_tunnel_start` and keep:
   - `PID_FILE=build/ssh-tunnels/phase4-a51.pid`
   - `LOCAL_PORT` used for forward target `127.0.0.1:22`.

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
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.
- Delete temporary key files and tunnel artifacts created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Do not assume external cleanup; always validate local cleanup in this run.
```
