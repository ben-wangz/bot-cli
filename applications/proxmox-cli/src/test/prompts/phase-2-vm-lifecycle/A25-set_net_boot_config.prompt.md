# A25 set_net_boot_config

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase2-vm-lifecycle.shared-node` exists.
- `build/phase2-vm-lifecycle.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A25 `set_net_boot_config` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase2-vm-lifecycle.shared-node`.
3) Read `SHARED_VMID` from `build/phase2-vm-lifecycle.shared-vmid`.
4) Prepare URL-safe values for `NET0_VALUE` and `BOOT_VALUE`.
   - Recommended `BOOT_VALUE`: `order=scsi0;net0`

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action set_net_boot_config --node "$SHARED_NODE" --vmid "$SHARED_VMID" --net0 "$NET0_VALUE" --boot "$BOOT_VALUE"

Success criteria:
- exit code = 0
- JSON field `action == "set_net_boot_config"`
- JSON field `ok == true`
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
