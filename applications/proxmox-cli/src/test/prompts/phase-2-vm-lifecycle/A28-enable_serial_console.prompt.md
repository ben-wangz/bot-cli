# A28 enable_serial_console

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable VM.

## Prompt

```text
You are a test execution agent. Run the A28 `enable_serial_console` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action enable_serial_console --node "$TEST_NODE" --vmid "$TEST_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "enable_serial_console"`
- JSON field `ok == true`
- `result.upid` is non-empty

Teardown:
- Destroy `TEST_VMID` in this run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
