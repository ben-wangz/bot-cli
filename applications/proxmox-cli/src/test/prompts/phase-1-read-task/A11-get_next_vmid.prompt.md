# A11 get_next_vmid

## Preconditions

- `build/pve-user.env` is loaded.

## Prompt

```text
You are a test execution agent. Run the A11 `get_next_vmid` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`

Command:
go run ./cmd/proxmox-cli --output json action get_next_vmid

Success criteria:
- exit code = 0
- JSON field `action == "get_next_vmid"`
- JSON field `ok == true`
- `result.next_vmid` is a positive integer

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
