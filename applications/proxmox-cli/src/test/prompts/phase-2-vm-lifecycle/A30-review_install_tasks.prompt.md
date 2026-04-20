# A30 review_install_tasks

## Preconditions

- `build/pve-user.env` is loaded.
- A template VM exists for creating a disposable VM.

## Prompt

```text
You are a test execution agent. Run the A30 `review_install_tasks` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Resolve `TEST_NODE`.
3) Allocate fresh `TEST_VMID` and create a disposable VM.
4) Trigger at least one task for `TEST_VMID` within this prompt run (for example a start/stop cycle).

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action review_install_tasks --node "$TEST_NODE" --vmid "$TEST_VMID"

Success criteria:
- exit code = 0
- JSON field `action == "review_install_tasks"`
- JSON field `ok == true`
- `result` is an array
- JSON contains `diagnostics.wait_skipped == "action is synchronous"`

Teardown:
- Stop and destroy `TEST_VMID` created for this prompt.

Independence rule:
- This test must be self-contained and order-independent.
- Never reuse a VMID created by another prompt.
```
