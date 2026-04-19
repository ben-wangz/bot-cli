# A02 list_cluster_resources

## Preconditions

- `build/pve-user.env` is present and valid.
- At least one cluster resource exists.

## Prompt

```text
You are a test execution agent. Run the A02 `list_cluster_resources` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`

Command:
go run ./cmd/proxmox-cli --output json action list_cluster_resources --type vm

Success criteria:
- exit code = 0
- JSON field `action == "list_cluster_resources"`
- JSON field `ok == true`
- `result` is an array

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
