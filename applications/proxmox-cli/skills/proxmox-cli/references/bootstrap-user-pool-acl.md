# Bootstrap User/Pool ACL

Use this flow to create or reuse a bot user and pool, then grant required ACL bindings.

## When to Use

- New environment with only root credentials
- Need to switch to least-privilege operations for agent runtime

## Workflow Command

Run with root scope using resolved binary path:

```bash
"${PROXMOX_CLI_BIN}" \
  --api-base "${PVE_API_BASE_URL%/}/api2/json" \
  --insecure-tls \
  --output json \
  --auth-scope root \
  --auth-user "$PVE_ROOT_USER" \
  --auth-password "$PVE_ROOT_PASSWORD" \
  workflow bootstrap-bot-user-pool-acl \
  --userid "$BOT_USERID" \
  --poolid "$BOT_POOLID" \
  --password "$BOT_PASSWORD" \
  --if-exists reuse \
  --user-comment "bot bootstrap user" \
  --pool-comment "bot bootstrap pool" \
  --sdn-acl-path "/sdn/zones/localnetwork"
```

## Validation

After workflow success, validate both binding and effective permissions:

```bash
"${PROXMOX_CLI_BIN}" ... capability get_user_acl_binding --userid "$BOT_USERID"
"${PROXMOX_CLI_BIN}" ... capability get_effective_permissions --path "/pool/${BOT_POOLID}"
```

Expect:

- user exists and is bound to intended ACL paths
- pool path permissions are effective for runtime actions

## Common Mistakes

1. Using user scope for bootstrap steps.
2. Missing pool parameter in later clone/create calls.
3. Reusing a user without rechecking ACL bindings.
