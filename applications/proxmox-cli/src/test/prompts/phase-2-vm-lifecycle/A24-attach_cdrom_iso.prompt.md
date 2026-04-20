# A24 attach_cdrom_iso

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable VM.
- A readable ISO path `TEST_ISO` exists (example: `local:iso/ubuntu.iso`).

## Prompt

```text
You are a test execution agent. Run the A24 `attach_cdrom_iso` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.
3) Resolve `TEST_ISO` from available ISO storage.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action attach_cdrom_iso --node "$TEST_NODE" --vmid "$TEST_VMID" --iso "$TEST_ISO" --slot ide2

Success criteria:
- exit code = 0
- JSON field `action == "attach_cdrom_iso"`
- JSON field `ok == true`
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Teardown:
- Destroy `TEST_VMID` in this run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
