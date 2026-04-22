# A26 start_installer_and_console_ticket

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A26 `start_installer_and_console_ticket` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`).
7) Ensure VM is stopped before action execution (`vm_power --mode stop --desired-state stopped --wait`).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action start_installer_and_console_ticket --node "$TEST_NODE" --vmid "$TEST_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "start_installer_and_console_ticket"`
- JSON field `ok == true`
- `result.upid` is non-empty
- `diagnostics.wait_status.status == "stopped"`
- `diagnostics.wait_status.exitstatus == "OK"`

Teardown:
- Stop and destroy `TEST_VMID` inside this prompt run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
