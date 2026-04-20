# A13 vm_power

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable power-control target.

## Prompt

```text
You are a test execution agent. Run the A13 `vm_power` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.
4) Ensure VM starts from `stopped` state.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action vm_power --node "$TEST_NODE" --vmid "$TEST_VMID" --mode start

Success criteria:
- exit code = 0
- JSON field `action == "vm_power"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
