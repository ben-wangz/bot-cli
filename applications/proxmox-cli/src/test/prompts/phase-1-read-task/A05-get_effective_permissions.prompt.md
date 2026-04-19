# A05 get_effective_permissions

## Preconditions

- `build/pve-user.env` is loaded.
- `<PATH>` is a valid ACL path (default `/`).

## Prompt

```text
You are a test execution agent. Run the A05 `get_effective_permissions` positive-path test in the bot-cli repository.

Setup:
1) Load env vars: `source build/pve-user.env`
2) Change directory to `applications/proxmox-cli/src`

Command:
go run ./cmd/proxmox-cli --output json action get_effective_permissions --path <PATH>

Success criteria:
- exit code = 0
- JSON field `action == "get_effective_permissions"`
- JSON field `ok == true`
- `request.path` exists

Return only this structure:
- action
- command
- success
- key_result
- diagnostics
```
