# A03 list_vms_by_node

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase1-read-task.shared-node` exists.

## Prompt

```text
You are a test execution agent. Run the A03 `list_vms_by_node` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase1-read-task.shared-node`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action list_vms_by_node --node "$SHARED_NODE"

Success criteria:
- exit code = 0
- JSON field `action == "list_vms_by_node"`
- JSON field `ok == true`
- `request.node == SHARED_NODE`

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create, stop, or destroy VMs inside this prompt run.
```
