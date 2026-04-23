# A44 ssh_check_service

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A44 `ssh_check_service` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` via `get_next_vmid`, clone from template (`clone_template --wait`), and start VM (`vm_power --mode start --desired-state running --wait`).
6) Use `agent_network_get_interfaces` to resolve a reachable guest IPv4 as `GUEST_IP`.
7) Prepare SSH key pair in `build/` and inject public key by `ssh_inject_pubkey_qga`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_check_service --host "$GUEST_IP" --port 22 --user ubuntu --identity-file "build/phase4-a44-id_ed25519" --connect-timeout-seconds 5

Success criteria:
- exit code = 0
- JSON field `action == "ssh_check_service"`
- JSON field `ok == true`
- JSON contains `result.reachable == true`
- JSON contains `result.auth_ok == true`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.
- Delete temporary key files created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse VMID, IP, or key files from other prompt runs.
```
