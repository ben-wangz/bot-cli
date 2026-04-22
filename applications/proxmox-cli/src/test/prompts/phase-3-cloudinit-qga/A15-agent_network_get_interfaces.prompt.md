# A15 agent_network_get_interfaces

## Preconditions

- `build/pve-user.env` is loaded.
- A local Ubuntu template exists for disposable VM cloning on target node.

## Prompt

```text
You are a test execution agent. Run the A15 `agent_network_get_interfaces` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate a fresh `TEST_VMID` in allowed range, clone a disposable VM from local Ubuntu template, enable `agent=1`, then start/reboot it.
4) Bootstrap qga inside guest via Phase 4 actions:
   - use `open_vm_termproxy` to get serial session entry
   - use `serial_ws_session_control` script mode to run install/enable commands
5) Probe guest-agent readiness (retry with backoff). If still unavailable, return `blocked_qga_unavailable` and destroy `TEST_VMID`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action agent_network_get_interfaces --node "$TEST_NODE" --vmid "$TEST_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "agent_network_get_interfaces"`
- JSON field `ok == true`
- JSON contains `diagnostics.ipv4_addresses` (array, may be empty)

Teardown:
- Destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Resolve and clean up its own VMID during this prompt run.
```
