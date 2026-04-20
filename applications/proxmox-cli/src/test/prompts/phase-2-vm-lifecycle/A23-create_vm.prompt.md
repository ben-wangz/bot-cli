# A23 create_vm

## Preconditions

- `build/pve-user.env` is loaded.
- `TEST_NODE` and `TEST_STORAGE` are available for VM creation.

## Prompt

```text
You are a test execution agent. Run the A23 `create_vm` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE` from online nodes.
3) Resolve `TEST_VMID` via `get_next_vmid`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action create_vm --node "$TEST_NODE" --vmid "$TEST_VMID" --name "p2-a23-$TEST_VMID" --memory 2048 --cores 2

Success criteria:
- exit code = 0
- JSON field `action == "create_vm"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Destroy `TEST_VMID` in the same prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
