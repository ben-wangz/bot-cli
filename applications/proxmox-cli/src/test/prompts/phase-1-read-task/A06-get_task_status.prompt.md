# A06 get_task_status

## Preconditions

- `build/pve-user.env` is loaded.
- At least one online node has at least one VM with at least one task record.

## Prompt

```text
You are a test execution agent. Run the A06 `get_task_status` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`
3) Discover `NODE` from online nodes.
4) Discover `VMID` by calling `list_vms_by_node --node "$NODE"` and selecting a VMID in allowed range (`PVE_ALLOWED_VMID_MIN..PVE_ALLOWED_VMID_MAX`, default `1001..2000`).
5) Discover `UPID` by calling `list_tasks_by_vmid --node "$NODE" --vmid "$VMID"` and taking the first item.
6) If the selected VM has no tasks, try the next VM on the same node.
7) Fail if no `UPID` can be discovered for an allowed-range VMID.

Command:
go run ./cmd/proxmox-cli --output json action get_task_status --node "$NODE" --upid "$UPID"

Success criteria:
- exit code = 0
- JSON field `action == "get_task_status"`
- JSON field `ok == true`
- `diagnostics` contains at least `status` or `exitstatus`

Independence rule:
- This test must be self-contained and must not depend on outputs from other prompt files.
- Resolve `NODE`, `VMID`, and `UPID` locally inside this prompt execution.

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
