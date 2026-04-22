# A12 get_vm_status

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase1-read-task.shared-node` exists.
- `build/phase1-read-task.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A12 `get_vm_status` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase1-read-task.shared-node`.
3) Read `SHARED_VMID` from `build/phase1-read-task.shared-vmid`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action get_vm_status --node "$SHARED_NODE" --vmid "$SHARED_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "get_vm_status"`
- JSON field `ok == true`
- `request.node == SHARED_NODE` and `request.vmid == SHARED_VMID`

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create, stop, or destroy VMs inside this prompt run.
```
