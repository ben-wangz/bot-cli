# A27 render_and_serve_seed

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A27 `render_and_serve_seed` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`).
7) Prepare `build/regression-phase3/seed` for output.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action render_and_serve_seed --vmid "$TEST_VMID" --seed-dir build/regression-phase3/seed --seed-name "p3-$TEST_VMID" --host 127.0.0.1 --port 18088

Success criteria:
- exit code = 0
- JSON field `action == "render_and_serve_seed"`
- JSON field `ok == true`
- JSON contains `result.seed_path`
- JSON contains `result.seed_url`
- JSON contains `result.serve_address`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Never depend on files created by other prompt runs.
- Never reuse a VMID created by another prompt run.
```
