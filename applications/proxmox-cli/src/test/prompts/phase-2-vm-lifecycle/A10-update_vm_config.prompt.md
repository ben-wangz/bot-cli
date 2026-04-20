# A10 update_vm_config

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable config target.

## Prompt

```text
You are a test execution agent. Run the A10 `update_vm_config` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` using `get_next_vmid`.
4) Create dedicated `TEST_VMID` inside this prompt run.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action update_vm_config --node "$TEST_NODE" --vmid "$TEST_VMID" --description "phase2-a10"

Success criteria:
- exit code = 0
- JSON field `action == "update_vm_config"`
- JSON field `ok == true`
- `result.upid` is non-empty

Teardown:
- Destroy `TEST_VMID` in this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
