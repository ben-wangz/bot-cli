# P5-02 get_user_acl_binding

## Preconditions

- `build/pve-user.env` is loaded.
- Root-scope credentials are available for this prompt run (`--auth-scope root` or `root-token`).

## Prompt

```text
You are a test execution agent. Run the P5-02 `get_user_acl_binding` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Ensure test user exists by running `create_pve_user_with_root --userid botcli-phase5@pve --if-exists reuse`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json action get_user_acl_binding --userid "botcli-phase5@pve"

Success criteria:
- exit code = 0
- JSON field `action == "get_user_acl_binding"`
- JSON field `ok == true`
- JSON contains `result.userid == "botcli-phase5@pve"`
- JSON contains `result.bindings` (list) and `result.count`

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs.
- Prompt must be rerunnable.
```
