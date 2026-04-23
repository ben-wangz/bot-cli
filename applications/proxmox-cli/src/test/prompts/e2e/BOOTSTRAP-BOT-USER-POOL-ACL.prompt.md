# Bootstrap Bot User Pool ACL

## Purpose

Run workflow `bootstrap-bot-user-pool-acl` to create/reuse one bot user and one pool, then grant required ACLs for:

- pool-scoped VM admin (`/pool/<poolid>` + `PVEAdmin`)
- cluster/node read baseline for A01/A02 (`/` + `PVEAuditor`)
- datastore operations for A22 and ISO upload chain (`/storage` + `PVEDatastoreAdmin`)

## Prompt

```text
You are a coding/test execution agent. Run workflow-level bootstrap for bot user + pool ACL.

Execution requirements:
1) Work from repository root.
2) Load root credentials from `build/pve-root.env`.
3) Use API base `${PVE_API_BASE_URL%/}/api2/json` and always include `--insecure-tls`.
4) Resolve deterministic test inputs:
   - `BOT_USERID="botcli-bootstrap@pve"`
   - `BOT_POOLID="botcli-bootstrap-pool"`
5) Execute workflow:
   - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json workflow bootstrap-bot-user-pool-acl --userid "$BOT_USERID" --poolid "$BOT_POOLID" --if-exists reuse --user-comment "bot bootstrap user" --pool-comment "bot bootstrap pool"`
6) Validate workflow response:
   - top-level `workflow == "bootstrap-bot-user-pool-acl"`
   - top-level `ok == true`
   - `result.userid == BOT_USERID`
   - `result.poolid == BOT_POOLID`
   - `result.grants` contains all required tuples:
     - `/pool/botcli-bootstrap-pool` + `PVEAdmin`
     - `/` + `PVEAuditor`
     - `/storage` + `PVEDatastoreAdmin`
7) Validate ACL binding by follow-up action:
   - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json action get_user_acl_binding --userid "$BOT_USERID"`
   - assert binding list contains the three path/role tuples above.
8) Return only structured result:
   - workflow
   - command
   - success
   - userid
   - poolid
   - grants
   - diagnostics

Safety:
- Do not modify or delete non-test users/pools.
- Use idempotent mode (`--if-exists reuse`) to support reruns.
```
