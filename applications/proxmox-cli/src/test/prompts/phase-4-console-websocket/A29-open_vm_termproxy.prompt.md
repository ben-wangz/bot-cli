# A29 open_vm_termproxy

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A29 `open_vm_termproxy` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`).
7) Enable serial console (`enable_serial_console`) and start VM (`vm_power --mode start --desired-state running --wait`).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action open_vm_termproxy --node "$TEST_NODE" --vmid "$TEST_VMID" --serial serial0

Success criteria:
- exit code = 0
- JSON field `action == "open_vm_termproxy"`
- JSON field `ok == true`
- JSON contains `result.port`
- JSON contains `result.ticket`
- JSON contains `result.user`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Never depend on any previous websocket session.
- Never reuse a VMID created by another prompt run.
```
