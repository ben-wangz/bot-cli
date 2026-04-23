# A48 ssh_print_connect_command

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase4-ssh-control-plane.shared-guest-ip` exists.
- `build/phase4-ssh-control-plane.shared-identity-file` exists.

## Prompt

```text
You are a test execution agent. Run the A48 `ssh_print_connect_command` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_GUEST_IP` from `build/phase4-ssh-control-plane.shared-guest-ip`.
3) Read `SHARED_IDENTITY_FILE` from `build/phase4-ssh-control-plane.shared-identity-file`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_print_connect_command --host "$SHARED_GUEST_IP" --port 22 --user cloud --identity-file "$SHARED_IDENTITY_FILE" --extra-args "-o StrictHostKeyChecking=no"

Success criteria:
- exit code = 0
- JSON field `action == "ssh_print_connect_command"`
- JSON field `ok == true`
- JSON contains `result.command`
- `result.command` includes `ssh`, `cloud@`, and `-p 22`

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs in this run.
- Do not attempt interactive SSH inside this prompt.
```
