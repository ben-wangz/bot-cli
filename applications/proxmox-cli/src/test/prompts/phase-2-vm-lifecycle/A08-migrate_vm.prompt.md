# A08 migrate_vm

## Preconditions

- `build/pve-user.env` is loaded.
- Two online nodes are available.
- A disposable migration candidate can be created with a lightweight profile.

## Prompt

```text
You are a test execution agent. Run the A08 `migrate_vm` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `SOURCE_NODE` and `TARGET_NODE` from two different online nodes.
3) Allocate fresh `TEST_VMID` using `get_next_vmid`.
4) Create a disposable VM for this prompt only using a lightweight migration profile:
   - prefer no-disk config (config-only migration), or
   - if a disk is required by environment policy, use a small disk (for example 4G-8G).
5) Keep the VM in `stopped` state before migration.

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

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
