# A03 list_vms_by_node

## Preconditions

- `build/pve-user.env` is loaded.
- At least one online node is available.

## Prompt

```text
You are a test execution agent. Run the A03 `list_vms_by_node` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`
3) Discover `NODE` from `list_nodes` output:
   - pick the first node where `status == "online"`
   - fail if no online node is found

Command:
go run ./cmd/proxmox-cli --output json action list_vms_by_node --node "$NODE"

Success criteria:
- exit code = 0
- JSON field `action == "list_vms_by_node"`
- JSON field `ok == true`
- `request.node == NODE`

Independence rule:
- This test must be self-contained and must not depend on outputs from other prompt files.
- Resolve `NODE` locally inside this prompt execution.

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
