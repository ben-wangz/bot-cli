# A19 connect_vnc_websocket

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A19 `connect_vnc_websocket` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`).
7) Start cloned VM (`vm_power --mode start --desired-state running --wait`).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action connect_vnc_websocket --node "$TEST_NODE" --vmid "$TEST_VMID" --probe-seconds 2

Success criteria:
- exit code = 0
- JSON field `action == "connect_vnc_websocket"`
- JSON field `ok == true`
- JSON field `result.connected == true`
- JSON contains `result.websocket`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Never rely on another prompt to prepare vnc proxy metadata.
- Never reuse a VMID created by another prompt run.
```
