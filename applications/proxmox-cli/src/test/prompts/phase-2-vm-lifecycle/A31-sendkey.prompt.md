# A31 sendkey

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase2-vm-lifecycle.shared-node` exists.
- `build/phase2-vm-lifecycle.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A31 `sendkey` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_NODE` from `build/phase2-vm-lifecycle.shared-node`.
3) Read `SHARED_VMID` from `build/phase2-vm-lifecycle.shared-vmid`.
4) Ensure VM is running (`vm_power --mode start --desired-state running --wait`).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action sendkey --node "$SHARED_NODE" --vmid "$SHARED_VMID" --key ret

Success criteria:
- exit code = 0
- JSON field `action == "sendkey"`
- JSON field `ok == true`
- `diagnostics.method == "PUT"`
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Teardown:
- Restore shared VM to `stopped` state (`vm_power --mode stop --desired-state stopped --wait`).

Independence rule:
- This test must be self-contained for action execution and order-independent.
- Do not create or destroy VMs inside this prompt run.
```
