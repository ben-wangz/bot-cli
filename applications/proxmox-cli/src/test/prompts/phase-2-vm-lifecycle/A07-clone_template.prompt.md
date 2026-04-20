# A07 clone_template

## Preconditions

- `build/pve-user.env` is loaded.
- A disposable template VM exists on `TEST_NODE` as `SOURCE_VMID`.

## Prompt

```text
You are a test execution agent. Run the A07 `clone_template` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE` from an online node.
3) Resolve `TARGET_VMID` with `get_next_vmid`.
4) Ensure `SOURCE_VMID` is a template VM dedicated for this prompt run.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action clone_template --node "$TEST_NODE" --source-vmid "$SOURCE_VMID" --target-vmid "$TARGET_VMID" --name "p2-a07-$TARGET_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "clone_template"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Destroy cloned VM `TARGET_VMID` inside this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
