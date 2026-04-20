# A25 set_net_boot_config

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable VM.

## Prompt

```text
You are a test execution agent. Run the A25 `set_net_boot_config` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.
3) Prepare URL-safe values for `NET0_VALUE` and `BOOT_VALUE`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action set_net_boot_config --node "$TEST_NODE" --vmid "$TEST_VMID" --net0 "$NET0_VALUE" --boot "$BOOT_VALUE"

Success criteria:
- exit code = 0
- JSON field `action == "set_net_boot_config"`
- JSON field `ok == true`
- `result.upid` is non-empty

Teardown:
- Destroy `TEST_VMID` in this run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
