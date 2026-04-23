# A15 agent_network_get_interfaces

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase3-cloudinit-qga.shared-node` exists.
- `build/phase3-cloudinit-qga.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A15 `agent_network_get_interfaces` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase3-cloudinit-qga.shared-node`.
3) Read `SHARED_VMID` from `build/phase3-cloudinit-qga.shared-vmid`.
4) Ensure shared VM is running (`vm_power --mode start --desired-state running --wait`).
5) Probe guest-agent readiness with retry/backoff for `SHARED_VMID`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action agent_network_get_interfaces --node "$SHARED_NODE" --vmid "$SHARED_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "agent_network_get_interfaces"`
- JSON field `ok == true`
- JSON contains `diagnostics.ipv4_addresses` (array, may be empty)

Teardown:
- Restore shared VM to `stopped` state (`vm_power --mode stop --desired-state stopped --wait`).

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
