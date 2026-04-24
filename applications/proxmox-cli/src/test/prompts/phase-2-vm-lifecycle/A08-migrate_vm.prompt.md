# A08 migrate_vm

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.
- At least two online nodes are available.

## Prompt

```text
You are a test execution agent. Run the A08 `migrate_vm` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `SOURCE_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Resolve `TARGET_NODE` as a different online node (fail if none exists).
6) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
7) Clone `TEMPLATE_VMID` to `TEST_VMID` on `SOURCE_NODE` (`clone_template --wait --pool "$PVE_POOL"`).
8) Ensure cloned VM is stopped before migration (`vm_power --mode stop --desired-state stopped --wait`).

Command:
1) Start migrate task (no blocking wait):
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --timeout 10m --output json action migrate_vm --node "$SOURCE_NODE" --vmid "$TEST_VMID" --target "$TARGET_NODE"

2) Extract `UPID` from migrate result, then poll status with a long window:
- run `get_task_status --node "$SOURCE_NODE" --upid "$UPID"` every 30s
- max wait window: 4h
- success only when `status == "stopped"` and `exitstatus == "OK"`

Success criteria:
- exit code = 0
- JSON field `action == "migrate_vm"`
- JSON field `ok == true`
- `result.upid` is non-empty
- polled final task status is `stopped/OK`

Timeout diagnostics rule:
- If polling exceeds 4h, report timeout with `node`, `upid`, and latest status snapshot.
- Capture and return the latest `get_task_status` output for the same `upid` before teardown.

Teardown:
- Destroy `TEST_VMID` in this prompt run, even when migration fails.
- Resolve final host node for `TEST_VMID` before destroy if migration completed.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
