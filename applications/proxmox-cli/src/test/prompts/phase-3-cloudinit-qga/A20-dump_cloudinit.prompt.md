# A20 dump_cloudinit

## Preconditions

- `build/pve-user.env` is loaded.
- A VM with cloud-init config exists in allowed VMID range.

## Prompt

```text
You are a test execution agent. Run the A20 `dump_cloudinit` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Resolve one VMID in allowed range with cloud-init enabled.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action dump_cloudinit --node "$TEST_NODE" --vmid "$TEST_VMID" --type user

Success criteria:
- exit code = 0
- JSON field `action == "dump_cloudinit"`
- JSON field `ok == true`
- JSON contains `result.type == "user"`
- JSON contains `result.content`

Independence rule:
- This test must be self-contained and order-independent.
- Resolve its own VMID during this prompt run.
```
