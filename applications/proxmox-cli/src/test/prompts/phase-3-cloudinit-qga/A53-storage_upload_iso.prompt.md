# A53 storage_upload_iso

## Preconditions

- `build/pve-user.env` is loaded.
- Source ISO exists at `build/ubuntu-24.04.2-live-server-amd64.iso`.

## Prompt

```text
You are a test execution agent. Run the A53 `storage_upload_iso` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase3-cloudinit-qga.shared-node`.
3) Use storage `local` and deterministic filename `phase3-a53-upload.iso`.
4) Run guard first for better diagnostics:
   - `storage_upload_guard --node "$SHARED_NODE" --storage local --content-type iso`

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action storage_upload_iso --node "$SHARED_NODE" --storage local --source-path build/ubuntu-24.04.2-live-server-amd64.iso --filename phase3-a53-upload.iso --if-exists skip

Success criteria:
- exit code = 0
- JSON field `action == "storage_upload_iso"`
- JSON field `ok == true`
- JSON contains non-empty `result.volid`
- JSON contains `result.uploaded == true`

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs.
- Prompt must be rerunnable (`--if-exists skip`).
```
