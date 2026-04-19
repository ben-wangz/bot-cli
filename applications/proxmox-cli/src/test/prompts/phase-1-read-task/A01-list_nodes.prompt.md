# A01 list_nodes

## Preconditions

- `applications/proxmox-cli/src` can build successfully.
- Live env is available in `build/pve-user.env`.

## Prompt

```text
You are a test execution agent. Run the A01 `list_nodes` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`

Command:
go run ./cmd/proxmox-cli --output json action list_nodes

Success criteria:
- Process exit code is 0
- JSON field `action == "list_nodes"`
- JSON field `ok == true`
- `result` is an array with at least one node object

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
