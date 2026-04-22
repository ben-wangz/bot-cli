# A17 agent_exec

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A17 `agent_exec` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`) and start VM (`vm_power --mode start --desired-state running --wait`).
7) Probe guest-agent readiness with retry/backoff for `TEST_VMID`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action agent_exec --node "$TEST_NODE" --vmid "$TEST_VMID" --shell 1 --command "echo A17_EXEC_OK"

Success criteria:
- exit code = 0
- JSON field `action == "agent_exec"`
- JSON field `ok == true`
- JSON contains `result.pid`
- JSON contains `result.status.exited == true`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Resolve and clean up its own VMID during this prompt run.
```
