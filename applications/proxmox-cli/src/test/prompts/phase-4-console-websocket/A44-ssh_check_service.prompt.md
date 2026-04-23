# A44 ssh_check_service

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase4-ssh-control-plane.shared-node` exists.
- `build/phase4-ssh-control-plane.shared-vmid` exists.
- `build/phase4-ssh-control-plane.shared-guest-ip` exists.
- `build/phase4-ssh-control-plane.shared-identity-file` exists.

## Prompt

```text
You are a test execution agent. Run the A44 `ssh_check_service` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Set VMID policy env vars (`PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`).
3) Read `SHARED_NODE` from `build/phase4-ssh-control-plane.shared-node`.
4) Read `SHARED_VMID` from `build/phase4-ssh-control-plane.shared-vmid`.
5) Read `SHARED_GUEST_IP` from `build/phase4-ssh-control-plane.shared-guest-ip`.
6) Read `SHARED_IDENTITY_FILE` from `build/phase4-ssh-control-plane.shared-identity-file`.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action ssh_check_service --host "$SHARED_GUEST_IP" --port 22 --user cloud --identity-file "$SHARED_IDENTITY_FILE" --connect-timeout-seconds 5

Success criteria:
- exit code = 0
- JSON field `action == "ssh_check_service"`
- JSON field `ok == true`
- JSON contains `result.reachable == true`
- JSON contains `result.auth_ok == true`

Independence rule:
- This test consumes suite-level shared VM artifacts and must not create/stop/destroy VMs.
- This test must not mutate shared key assets.
```
