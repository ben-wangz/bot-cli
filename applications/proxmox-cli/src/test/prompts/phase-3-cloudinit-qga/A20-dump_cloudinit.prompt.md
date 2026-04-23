# A20 dump_cloudinit

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase3-cloudinit-qga.shared-node` exists.
- `build/phase3-cloudinit-qga.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A20 `dump_cloudinit` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase3-cloudinit-qga.shared-node`.
3) Read `SHARED_VMID` from `build/phase3-cloudinit-qga.shared-vmid`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action dump_cloudinit --node "$SHARED_NODE" --vmid "$SHARED_VMID" --type user

Success criteria:
- exit code = 0
- JSON field `action == "dump_cloudinit"`
- JSON field `ok == true`
- JSON contains `result.type == "user"`
- JSON contains `result.content`

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
