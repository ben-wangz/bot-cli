# A47 ssh_scp_transfer

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A47 `ssh_scp_transfer` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Resolve template/node, allocate `TEST_VMID`, clone template, and start VM.
4) Resolve `GUEST_IP` via `agent_network_get_interfaces`.
5) Generate key pair and inject public key via `ssh_inject_pubkey_qga`.
6) Create a local test file `build/phase4-a47-upload.txt`.

Command (upload):
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_scp_transfer --direction upload --host "$GUEST_IP" --port 22 --user ubuntu --identity-file "build/phase4-a47-id_ed25519" --local-path "build/phase4-a47-upload.txt" --remote-path "/home/ubuntu/phase4-a47-upload.txt"

Command (download):
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_scp_transfer --direction download --host "$GUEST_IP" --port 22 --user ubuntu --identity-file "build/phase4-a47-id_ed25519" --remote-path "/home/ubuntu/phase4-a47-upload.txt" --local-path "build/phase4-a47-download.txt"

Success criteria:
- both commands exit code = 0
- upload JSON: `action == "ssh_scp_transfer"`, `ok == true`, `result.direction == "upload"`
- download JSON: `action == "ssh_scp_transfer"`, `ok == true`, `result.direction == "download"`
- downloaded file content equals uploaded file content

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.
- Delete temporary key files and local transfer test files created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Never rely on files from other prompt runs.
```
