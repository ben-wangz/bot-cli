# P5-00 create_pool_with_root

## Preconditions

- `build/pve-user.env` is loaded.
- Root-scope credentials are available for this prompt run (`--auth-scope root` or `root-token`).

## Prompt

```text
You are a test execution agent. Run the P5-00 `create_pool_with_root` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Build a deterministic test pool id, for example `botcli-phase5-pool`.
3) Use idempotent mode `--if-exists reuse` to avoid hard failure on reruns.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json action create_pool_with_root --poolid "botcli-phase5-pool" --comment "phase5 bootstrap pool" --if-exists reuse

Success criteria:
- exit code = 0
- JSON field `action == "create_pool_with_root"`
- JSON field `ok == true`
- JSON contains `result.poolid == "botcli-phase5-pool"`
- JSON contains `result.created` or `result.reused`

Independence rule:
- This is a no-VM prompt: do not create/stop/destroy VMs.
- Prompt must be rerunnable without manual cleanup.
```
