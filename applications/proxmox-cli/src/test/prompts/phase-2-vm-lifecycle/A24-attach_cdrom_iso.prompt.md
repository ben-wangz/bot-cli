# A24 attach_cdrom_iso

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase2-vm-lifecycle.shared-node` exists.
- `build/phase2-vm-lifecycle.shared-vmid` exists.
- A readable ISO path `TEST_ISO` exists.

## Prompt

```text
You are a test execution agent. Run the A24 `attach_cdrom_iso` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase2-vm-lifecycle.shared-node`.
3) Read `SHARED_VMID` from `build/phase2-vm-lifecycle.shared-vmid`.
4) Resolve `TEST_ISO` from available ISO storage.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action attach_cdrom_iso --node "$SHARED_NODE" --vmid "$SHARED_VMID" --iso "$TEST_ISO" --slot ide2 --media cdrom

Success criteria:
- exit code = 0
- JSON field `action == "attach_cdrom_iso"`
- JSON field `ok == true`
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
