# A22 storage_upload_guard

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase3-cloudinit-qga.shared-node` exists.

## Prompt

```text
You are a test execution agent. Run the A22 `storage_upload_guard` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase3-cloudinit-qga.shared-node`.
3) Resolve one target storage name as `TEST_STORAGE`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action storage_upload_guard --node "$SHARED_NODE" --storage "$TEST_STORAGE" --content-type snippets

Success criteria:
- exit code = 0
- JSON field `action == "storage_upload_guard"`
- JSON field `ok == true`
- JSON contains `result.upload_allowed`
- If `result.upload_allowed == false`, JSON contains a non-empty `result.hint`

Independence rule:
- This test must be self-contained and order-independent.
- Resolve its own storage input during this prompt run.
- Do not create, stop, or destroy VMs inside this prompt run.
```
