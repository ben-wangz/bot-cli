# Setup: Prepare `build/pve-user.env` from `build/pve-root.env`

## Purpose

Provide one stable setup prompt for regression runs. This setup uses root bootstrap credentials to create a fresh bot user + pool, grants required ACLs, and writes `build/pve-user.env` for subsequent suites.

## Prompt

```text
You are a coding/test execution agent. Prepare `build/pve-user.env` from `build/pve-root.env` for regression testing.

Execution requirements:
1) Work from repository root.
2) Load root credentials from `build/pve-root.env`.
3) Build deterministic-per-run identifiers:
   - `RUN_ID="$(date +%Y%m%d-%H%M%S)"`
   - `BOT_USERID="botcli-reg-${RUN_ID}@pve"`
   - `BOT_POOLID="botcli-reg-${RUN_ID}-pool"`
   - `BOT_PASSWORD="$(python3 -c 'import secrets,string; s=string.ascii_letters+string.digits; print("".join(secrets.choice(s) for _ in range(20)))')"`
4) If `build/pve-user.env` exists, back it up to `build/pve-user.env.bak-${RUN_ID}`.
5) Execute workflow:
   - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json workflow bootstrap-bot-user-pool-acl --userid "$BOT_USERID" --poolid "$BOT_POOLID" --password "$BOT_PASSWORD" --if-exists fail --user-comment "regression bot user" --pool-comment "regression bot pool" --sdn-acl-path "/sdn/zones/localnetwork"`
6) Validate workflow response:
   - `workflow == "bootstrap-bot-user-pool-acl"`
   - `ok == true`
   - `result.userid == BOT_USERID`
   - `result.poolid == BOT_POOLID`
    - `result.grants` contains:
      - `/pool/${BOT_POOLID}` + `PVEAdmin`
      - `/` + `PVEAuditor`
      - `/storage` + `PVEDatastoreAdmin`
      - `/sdn/zones/localnetwork` + `PVEAdmin`
7) Write `build/pve-user.env` with exactly these keys:
   - `PVE_USER=${BOT_USERID}`
   - `PVE_PASSWORD=${BOT_PASSWORD}`
   - `PVE_POOL=${BOT_POOLID}`
   - `PVE_ACL_PATH=/pool/${BOT_POOLID}`
   - `PVE_API_BASE_URL=${PVE_API_BASE_URL}`
   - `PVE_API_SSL_SELF_SIGNED=1`
8) Verify new user env works:
   - `source build/pve-user.env`
   - run `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action get_effective_permissions --path "/pool/${PVE_POOL}"`
   - ensure `ok == true` and result for `/pool/${PVE_POOL}` is non-empty.
9) Return only structured output:
   - setup
   - workflow
   - success
   - backup_file
   - pve_user_env_path
   - userid
   - poolid
   - diagnostics

Safety:
- This setup is the only prompt that depends on `build/pve-root.env`.
- Do not modify unrelated users/pools.
```
