# A46 ssh_exec

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A46 `ssh_exec` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID`, resolve `TEST_NODE`, allocate `TEST_VMID`, clone template, and start VM.
4) Resolve `GUEST_IP` via `agent_network_get_interfaces`.
5) Generate key pair in `build/` and inject public key via `ssh_inject_pubkey_qga`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_exec --host "$GUEST_IP" --port 22 --user ubuntu --identity-file "build/phase4-a46-id_ed25519" --timeout-seconds 20 --command "hostname"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_exec"`
- JSON field `ok == true`
- JSON contains `result.exit_code == 0`
- JSON contains non-empty `result.stdout`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.
- Delete temporary key files created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Resolve all runtime inputs inside this prompt run.
```
