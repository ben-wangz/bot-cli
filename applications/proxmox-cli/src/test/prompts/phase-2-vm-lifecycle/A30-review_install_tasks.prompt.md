# A30 review_install_tasks

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase2-vm-lifecycle.shared-node` exists.
- `build/phase2-vm-lifecycle.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A30 `review_install_tasks` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase2-vm-lifecycle.shared-node`.
3) Read `SHARED_VMID` from `build/phase2-vm-lifecycle.shared-vmid`.
4) Trigger at least one task for `SHARED_VMID` within this prompt run (for example `vm_power` start then stop).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action review_install_tasks --node "$SHARED_NODE" --vmid "$SHARED_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "review_install_tasks"`
- JSON field `ok == true`
- `result` is an array
- JSON contains numeric `diagnostics.active_task_count`

Teardown:
- Restore shared VM to `stopped` state (`vm_power --mode stop --desired-state stopped --wait`).

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
