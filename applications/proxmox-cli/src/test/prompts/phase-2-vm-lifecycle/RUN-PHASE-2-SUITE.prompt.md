# Phase 2 Suite Runner

## Purpose

Run all Phase 2 action prompts under `applications/proxmox-cli/src/test/prompts/phase-2-vm-lifecycle/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A07, A08, A10, A26, A30, A31).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- Prompts are split into two execution classes:
  - Independent-VM prompts: A07, A08 (each prompt provisions and destroys its own VM assets).
  - Shared-VM prompts: A10, A26, A30, A31 (consume one suite-level shared VM).
- Shared-VM prompts must run sequentially (shared VM write concurrency = 1).
- Independent-VM prompts may run concurrently up to the suite concurrency limit when infra allows.
- A prompt must never reuse a VMID produced by another prompt, except the suite-level shared VM artifacts explicitly defined below.
- For async actions, prompt commands must include `--wait`, and validation must assert final task success (`status=stopped` and `exitstatus=OK`).

## Fallback Rule

- If sub-agents are not supported, stop and return `sub_agent_not_supported`.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.
6) Set VMID policy env vars for this suite: `PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000` (or approved override).
7) Do not execute bootstrap prompt in this suite. Require setup prompt (`../setup.md`) to have prepared `build/ubuntu-24-with-agent.vm-template.id`; if missing, stop and return `setup_template_id_missing`.
8) Resolve `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id` and resolve `SHARED_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
9) Allocate one fresh in-range `SHARED_VMID` via `get_next_vmid`, then clone once: `clone_template --wait --full 0 --pool "$PVE_POOL"` from `TEMPLATE_VMID` to `SHARED_VMID` on `SHARED_NODE`.
10) Persist shared inputs for sub-agents:
   - `build/phase2-vm-lifecycle.shared-node`
   - `build/phase2-vm-lifecycle.shared-vmid`
11) Independent-VM prompts must still resolve their own `TEST_VMID` and self-destroy assets inside each prompt run.
12) For `A08-migrate_vm`, choose a target node different from template node and clean up VM on its final host.
13) For disposable clones, prefer linked clone (`full=0`, default) to minimize storage I/O.
14) After all prompts finish (success or failure), run suite teardown once: stop and destroy `SHARED_VMID`, then remove shared input files.
15) For `A08-migrate_vm`, do not rely on one-shot `--wait`; start migrate task first, then poll `get_task_status` (30s interval, up to 4h) until `stopped/OK`. On timeout, include `upid` and latest status snapshot in diagnostics.

## Prompt Files to Execute

- Independent-VM prompts:
  - `A07-clone_template.prompt.md`
  - `A08-migrate_vm.prompt.md`
- Shared-VM prompts:
  - `A10-update_vm_config.prompt.md`
  - `A26-start_installer_and_console_ticket.prompt.md`
  - `A30-review_install_tasks.prompt.md`
  - `A31-sendkey.prompt.md`

## Final Output Format

Return one JSON object:

```json
{
  "suite": "phase-2-vm-lifecycle",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 6,
    "failed": 0
  },
  "results": [
    {
      "action": "clone_template",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
