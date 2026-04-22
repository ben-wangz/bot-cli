# A25 set_net_boot_config

## Preconditions

- `build/pve-user.env` is loaded.
- `build/ubuntu-24-with-agent.vm-template.id` exists and points to a valid template VM in policy VMID range.

## Prompt

```text
You are a test execution agent. Run the A25 `set_net_boot_config` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
4) Resolve `TEST_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
5) Allocate fresh `TEST_VMID` in-range via `get_next_vmid`.
6) Clone `TEMPLATE_VMID` to `TEST_VMID` on `TEST_NODE` (`clone_template --wait`).
7) Prepare URL-safe values for `NET0_VALUE` and `BOOT_VALUE`.
   - Recommended `BOOT_VALUE`: `order=scsi0;net0`

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --wait --output json action set_net_boot_config --node "$TEST_NODE" --vmid "$TEST_VMID" --net0 "$NET0_VALUE" --boot "$BOOT_VALUE"

Success criteria:
- exit code = 0
- JSON field `action == "set_net_boot_config"`
- JSON field `ok == true`
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Teardown:
- Destroy `TEST_VMID` in this run.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
