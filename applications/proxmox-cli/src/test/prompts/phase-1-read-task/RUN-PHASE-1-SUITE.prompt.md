# Phase 1 Suite Runner

## Purpose

Run all Phase 1 action prompts under `applications/proxmox-cli/src/test/prompts/phase-1-read-task/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A01, A02, A03, A04, A05, A06, A11, A12, A21).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- The parent agent aggregates all sub-agent outputs into a single final report.
- Every prompt must be order-independent for action execution and must remain read-only.
- Prompts may consume shared suite bootstrap artifacts (`SHARED_NODE`, `SHARED_VMID`) produced once before sub-agent fan-out.
- Phase 1 action commands are read-only; mutating helper actions are allowed only in shared suite setup/teardown.

## Fallback Rule

- If the runtime does not support sub-agents, stop immediately and return: `sub_agent_not_supported`.
- Do not run the suite sequentially in a single agent when sub-agents are unavailable.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.
6) Set VMID policy env vars for this suite: `PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000` (or project-approved override).
7) Do not execute bootstrap prompt in this suite. Require setup prompt (`../setup.md`) to have prepared `build/ubuntu-24-with-agent.vm-template.id`; if missing, stop and return `setup_template_id_missing`.
8) Resolve `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id` and resolve `SHARED_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
9) Allocate one fresh in-range `SHARED_VMID` via `get_next_vmid`, then clone once: `clone_template --wait --full 0` from `TEMPLATE_VMID` to `SHARED_VMID` on `SHARED_NODE`.
10) Persist shared inputs for sub-agents:
   - `build/phase1-read-task.shared-node`
   - `build/phase1-read-task.shared-vmid`
11) Empty-cluster guard: do not rely on pre-existing ordinary VMs; the shared clone in step 9 guarantees at least one disposable VM exists.
12) After all sub-agents finish (success or failure), run suite teardown once: stop and destroy `SHARED_VMID`, then remove shared input files.

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
