# A31 sendkey

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable VM.

## Prompt

```text
You are a test execution agent. Run the A31 `sendkey` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.
4) Ensure VM is running.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action sendkey --node "$TEST_NODE" --vmid "$TEST_VMID" --key ret

Success criteria:
- exit code = 0
- JSON field `action == "sendkey"`
- JSON field `ok == true`
- `diagnostics.method == "PUT"`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
