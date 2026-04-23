# A45 ssh_inject_pubkey_qga

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A45 `ssh_inject_pubkey_qga` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE`, allocate `TEST_VMID`, clone template (`clone_template --wait`), and start VM.
5) Generate temporary key pair under `build/phase4-a45-id_ed25519`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_inject_pubkey_qga --node "$TEST_NODE" --vmid "$TEST_VMID" --username ubuntu --pub-key-file "build/phase4-a45-id_ed25519.pub"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_inject_pubkey_qga"`
- JSON field `ok == true`
- JSON contains `result.username == "ubuntu"`
- JSON contains `result.authorized_keys_path`
- JSON contains `result.fingerprint`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.
- Delete temporary key files created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Never depend on key material from other prompt runs.
```
