# A28 enable_serial_console

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase2-vm-lifecycle.shared-node` exists.
- `build/phase2-vm-lifecycle.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A28 `enable_serial_console` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase2-vm-lifecycle.shared-node`.
3) Read `SHARED_VMID` from `build/phase2-vm-lifecycle.shared-vmid`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action enable_serial_console --node "$SHARED_NODE" --vmid "$SHARED_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "enable_serial_console"`
- JSON field `ok == true`
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
