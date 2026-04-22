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
7) Clone `TEMPLATE_VMID` to `TEST_VMID` on `SOURCE_NODE` (`clone_template --wait`).
8) Ensure cloned VM is stopped before migration (`vm_power --mode stop --desired-state stopped --wait`).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action migrate_vm --node "$SOURCE_NODE" --vmid "$TEST_VMID" --target "$TARGET_NODE"

Success criteria:
- exit code = 0
- JSON field `action == "migrate_vm"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Destroy `TEST_VMID` in this prompt run, even when migration fails.
- Resolve final host node for `TEST_VMID` before destroy if migration completed.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
