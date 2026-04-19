# A21 list_tasks_by_vmid

## Preconditions

- `build/pve-user.env` is loaded.
- At least one online node has at least one VM.

## Prompt

```text
You are a test execution agent. Run the A21 `list_tasks_by_vmid` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`
3) Discover `NODE` from online nodes.
4) Discover `VMID` by calling `list_vms_by_node --node "$NODE"` and selecting the first VM.
5) Fail if no VM is found.

Command:
go run ./cmd/proxmox-cli --output json action list_tasks_by_vmid --node "$NODE" --vmid "$VMID"

Success criteria:
- exit code = 0
- JSON field `action == "list_tasks_by_vmid"`
- JSON field `ok == true`
- `result` is an array

Independence rule:
- This test must be self-contained and must not depend on outputs from other prompt files.
- Resolve `NODE` and `VMID` locally inside this prompt execution.

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
