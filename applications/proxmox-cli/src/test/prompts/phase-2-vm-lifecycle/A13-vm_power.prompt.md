# A13 vm_power

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase2-vm-lifecycle.shared-node` exists.
- `build/phase2-vm-lifecycle.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A13 `vm_power` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase2-vm-lifecycle.shared-node`.
3) Read `SHARED_VMID` from `build/phase2-vm-lifecycle.shared-vmid`.
4) Ensure shared VM starts from `stopped` state (`vm_power --mode stop --desired-state stopped --wait`).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action vm_power --node "$SHARED_NODE" --vmid "$SHARED_VMID" --mode start --desired-state running

Success criteria:
- exit code = 0
- JSON field `action == "vm_power"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Restore shared VM to `stopped` state (`vm_power --mode stop --desired-state stopped --wait`).

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
