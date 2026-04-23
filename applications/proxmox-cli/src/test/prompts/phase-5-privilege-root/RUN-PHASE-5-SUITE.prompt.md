# Phase 5 Suite Runner

## Purpose

Run all Phase 5 root-assisted user ACL bootstrap prompts under `applications/proxmox-cli/src/test/prompts/phase-5-privilege-root/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per prompt file (P5-01..P5-04).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- This suite has no VM dependency; prompts must not create/stop/destroy VMs.
- Because prompts operate on shared ACL entities, run prompts sequentially (effective write concurrency = 1).

## Fallback Rule

- If sub-agents are not supported, stop and return `sub_agent_not_supported`.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.
6) Use root-capable auth scope for this suite (`--auth-scope root` preferred).
7) Use deterministic test user id `botcli-phase5@pve` for idempotent reruns.

## Prompt Files to Execute

- `P5-01-create_pve_user_with_root.prompt.md`
- `P5-02-get_user_acl_binding.prompt.md`
- `P5-03-grant_user_acl.prompt.md`
- `P5-04-revoke_user_acl.prompt.md`

## Final Output Format

Return one JSON object:

```json
{
  "suite": "phase-5-privilege-root",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 4,
    "failed": 0
  },
  "results": [
    {
      "action": "create_pve_user_with_root",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
