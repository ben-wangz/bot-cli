# A09 convert_vm_to_template

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable conversion target.

## Prompt

```text
You are a test execution agent. Run the A09 `convert_vm_to_template` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE` from online nodes.
3) Allocate fresh `TEST_VMID` using `get_next_vmid`.
4) Create a dedicated stopped VM `TEST_VMID` for this prompt.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action convert_vm_to_template --node "$TEST_NODE" --vmid "$TEST_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "convert_vm_to_template"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Delete template `TEST_VMID` in this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
