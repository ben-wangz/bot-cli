# A40 validate_serial_output_criterion2

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A40 `validate_serial_output_criterion2` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`).
7) Enable serial console (`enable_serial_console`) and start VM (`vm_power --mode start --desired-state running --wait`).
8) Prepare `build/regression-phase4` directory for capture logs.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action validate_serial_output_criterion2 --node "$TEST_NODE" --vmid "$TEST_VMID" --log-path build/regression-phase4/serial-$TEST_VMID.log --append 1 --script "<ENTER>" --timeout-seconds 45

Success criteria:
- exit code = 0
- JSON field `action == "validate_serial_output_criterion2"`
- JSON field `ok == true`
- JSON field `result.criterion2_passed == true`
- JSON contains `result.log_path`

Teardown:
- Stop and destroy `TEST_VMID` in this prompt run on both success and failure.

Independence rule:
- This test must be self-contained and order-independent.
- Do not overwrite shared log files from other prompt runs.
- Never reuse a VMID created by another prompt run.
```
