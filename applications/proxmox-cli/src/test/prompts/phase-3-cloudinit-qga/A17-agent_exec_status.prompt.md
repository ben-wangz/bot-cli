# A17 agent_exec_status

## Preconditions

- `build/pve-user.env` is loaded.
- A local Ubuntu template exists for disposable VM cloning on target node.

## Prompt

```text
You are a test execution agent. Run the A17 `agent_exec_status` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate a fresh `TEST_VMID` in allowed range, clone a disposable VM from local Ubuntu template, enable `agent=1`, then start/reboot it.
4) Bootstrap qga inside guest via Phase 4 actions:
   - use `open_vm_termproxy` to get serial session entry
   - use `serial_ws_session_control` script mode to run install/enable commands
5) Probe guest-agent readiness (retry with backoff). If still unavailable, return `blocked_qga_unavailable` and destroy `TEST_VMID`.
6) First run `agent_exec` to obtain a `PID` for this same VM:
   go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action agent_exec --node "$TEST_NODE" --vmid "$TEST_VMID" --command /usr/bin/true

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action agent_exec_status --node "$TEST_NODE" --vmid "$TEST_VMID" --pid "$PID"

Success criteria:
- exit code = 0
- JSON field `action == "agent_exec_status"`
- JSON field `ok == true`
- JSON contains `result.exited`

Teardown:
- Destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Produce and use its own PID within this prompt run.
```
