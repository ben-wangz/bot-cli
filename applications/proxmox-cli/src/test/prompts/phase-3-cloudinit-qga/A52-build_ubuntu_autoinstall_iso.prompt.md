# A52 build_ubuntu_autoinstall_iso

## Preconditions

- `build/pve-user.env` is loaded.
- Host runtime has required local tools for ISO build path (`mount`, `umount`, `cp`, `mkisofs`).
- Source ISO exists at `build/ubuntu-24.04.2-live-server-amd64.iso`.

## Prompt

```text
You are a test execution agent. Run the A52 `build_ubuntu_autoinstall_iso` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Ensure source ISO exists: `build/ubuntu-24.04.2-live-server-amd64.iso`.
3) Use deterministic output path: `build/phase3-a52-autoinstall.iso`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action build_ubuntu_autoinstall_iso --source-iso build/ubuntu-24.04.2-live-server-amd64.iso --output-iso build/phase3-a52-autoinstall.iso --work-dir build/autoinstall-iso-work/phase3-a52 --password "ChangeMe123!"

Success criteria:
- exit code = 0
- JSON field `action == "build_ubuntu_autoinstall_iso"`
- JSON field `ok == true`
- JSON contains non-empty `result.output_iso`
- output ISO file exists and is non-empty

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs.
- Prompt must be rerunnable and may overwrite build output under `build/`.
```
