# A05 get_effective_permissions

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase1-read-task.shared-node` exists.
- `build/phase1-read-task.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A05 `get_effective_permissions` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase1-read-task.shared-node`.
3) Read `SHARED_VMID` from `build/phase1-read-task.shared-vmid`.
4) Validate shared VM is reachable: `get_vm_status --node "$SHARED_NODE" --vmid "$SHARED_VMID"`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action get_effective_permissions --path /

Success criteria:
- exit code = 0
- JSON field `action == "get_effective_permissions"`
- JSON field `ok == true`
- `request.path` exists

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create, stop, or destroy VMs inside this prompt run.
```
