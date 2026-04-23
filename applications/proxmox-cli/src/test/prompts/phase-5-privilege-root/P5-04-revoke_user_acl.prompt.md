# P5-04 revoke_user_acl

## Preconditions

- `build/pve-user.env` is loaded.
- Root-scope credentials are available for this prompt run (`--auth-scope root` or `root-token`).

## Prompt

```text
You are a test execution agent. Run the P5-04 `revoke_user_acl` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Ensure test user exists: `create_pve_user_with_root --userid botcli-phase5@pve --if-exists reuse`.
3) Ensure target ACL exists first by running:
   - `grant_user_acl --userid botcli-phase5@pve --path /vms --role PVEAuditor --propagate 1`

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json action revoke_user_acl --userid "botcli-phase5@pve" --path "/vms" --role "PVEAuditor"

Success criteria:
- exit code = 0
- JSON field `action == "revoke_user_acl"`
- JSON field `ok == true`
- JSON contains `result.userid == "botcli-phase5@pve"`
- JSON contains `result.path == "/vms"` and `result.role == "PVEAuditor"`
- JSON contains `result.revoked == true`

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs.
- Prompt must be rerunnable (already-absent binding should not hard fail).
```
