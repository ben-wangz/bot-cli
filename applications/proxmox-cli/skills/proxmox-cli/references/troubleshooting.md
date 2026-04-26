# Troubleshooting

Use this checklist for common proxmox-cli failures.

## 403 Permission Check Failed

Likely causes:

1. wrong auth scope (`user` vs `root`)
2. missing ACL grant for pool/path
3. operation targets VM or storage outside granted scope

Actions:

1. verify current scope and identity
2. re-check `get_user_acl_binding`
3. re-check `get_effective_permissions` on target path
4. ensure clone/create includes `--pool "$PVE_POOL"`

## Async Task Timeout

Likely causes:

1. normal long-running operation (migration/install)
2. hidden installer/guest failure

Actions:

1. inspect `get_task_status` and `list_tasks_by_vmid`
2. for provisioning, run `review_install_tasks`
3. resume supported workflow checkpoint when safe

## QGA Not Ready

Likely causes:

1. guest not fully booted
2. VM agent not enabled/configured
3. install stage incomplete

Actions:

1. verify VM power state
2. poll `agent_network_get_interfaces`
3. run `agent_exec --command hostname` once interfaces appear

## SSH Validation Fails

Likely causes:

1. key injection failed or wrong target user/path
2. guest networking not ready
3. SSH service not yet listening

Actions:

1. inject key again via `ssh_inject_pubkey_qga`
2. run `ssh_check_service`
3. verify guest IP via QGA interfaces
