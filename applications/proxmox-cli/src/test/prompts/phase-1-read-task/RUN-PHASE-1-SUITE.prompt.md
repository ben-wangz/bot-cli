# Phase 1 Suite Runner

## Purpose

Run all Phase 1 action prompts under `applications/proxmox-cli/src/test/prompts/phase-1-read-task/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A01, A02, A03, A04, A05, A06, A11, A12, A21).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- The parent agent aggregates all sub-agent outputs into a single final report.
- Every prompt must be order-independent: no prompt may consume outputs or side effects produced by another prompt.
- Each prompt must resolve its own runtime inputs (`NODE`, `VMID`, `UPID`, etc.) inside the same prompt execution.
- Phase 1 actions are read-only; `--wait` is not required in Phase 1 prompts.

## Fallback Rule

- If the runtime does not support sub-agents, stop immediately and return: `sub_agent_not_supported`.
- Do not run the suite sequentially in a single agent when sub-agents are unavailable.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.

## Prompt Files to Execute

- `A01-list_nodes.prompt.md`
- `A02-list_cluster_resources.prompt.md`
- `A03-list_vms_by_node.prompt.md`
- `A04-get_vm_config.prompt.md`
- `A05-get_effective_permissions.prompt.md`
- `A06-get_task_status.prompt.md`
- `A11-get_next_vmid.prompt.md`
- `A12-get_vm_status.prompt.md`
- `A21-list_tasks_by_vmid.prompt.md`

## Final Output Format

Return one JSON object with this schema:

```json
{
  "suite": "phase-1-read-task",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 9,
    "failed": 0
  },
  "results": [
    {
      "action": "list_nodes",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
