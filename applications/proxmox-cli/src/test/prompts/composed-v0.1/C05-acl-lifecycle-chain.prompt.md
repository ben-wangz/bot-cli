# C05 acl lifecycle chain

## Goal

Use one ACL chain to cover `revoke_user_acl` with idempotent lifecycle checks.

## Prompt

```text
You are a test execution agent. Execute ACL lifecycle chain.

Setup:
1) Load `build/pve-root.env`, switch to `applications/proxmox-cli/src`.
2) Ensure user/pool bootstrap by running workflow `bootstrap-bot-user-pool-acl` in `--if-exists reuse` mode for deterministic userid/poolid.

Chain:
1) Ensure `/vms + PVEAuditor` grant exists via `grant_user_acl`.
2) Revoke it via `revoke_user_acl`.
3) Re-run `get_user_acl_binding` and verify tuple absent.

Validation:
- All actions return `ok == true`.
- Revoke result contains `revoked == true`.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
