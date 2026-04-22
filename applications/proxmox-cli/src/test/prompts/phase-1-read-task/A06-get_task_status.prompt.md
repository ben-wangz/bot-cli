# A06 get_task_status

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A06 `get_task_status` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`) and capture returned `result.upid` as `CLONE_UPID`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action get_task_status --node "$TEST_NODE" --upid "$CLONE_UPID"

Success criteria:
- exit code = 0
- JSON field `action == "get_task_status"`
- JSON field `ok == true`
- `diagnostics` contains at least `status` or `exitstatus`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Resolve and clean up its own VMID and UPID during this prompt run.
```
