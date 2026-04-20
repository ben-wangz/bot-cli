# A26 start_installer_and_console_ticket

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable VM.

## Prompt

```text
You are a test execution agent. Run the A26 `start_installer_and_console_ticket` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.
4) Ensure VM is stopped before action execution.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action start_installer_and_console_ticket --node "$TEST_NODE" --vmid "$TEST_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "start_installer_and_console_ticket"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Stop and destroy `TEST_VMID` inside this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
