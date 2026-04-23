# Phase 3 Suite Runner

## Purpose

Run all Phase 3 action prompts under `applications/proxmox-cli/src/test/prompts/phase-3-cloudinit-qga/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A15, A17-exec, A17-status, A20, A22, A27).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- Prompts are split into three execution classes:
  - Independent-VM prompts: A17-agent_exec, A17-agent_exec_status.
  - Shared-VM prompts: A15-agent_network_get_interfaces, A20-dump_cloudinit.
  - No-VM prompts: A22-storage_upload_guard, A27-render_and_serve_seed.
- Shared-VM prompts must run sequentially (shared VM write concurrency = 1).
- Independent-VM prompts may run concurrently up to the suite concurrency limit when infra allows.
- No-VM prompts may run concurrently.
- A prompt must never reuse a VMID produced by another prompt, except the suite-level shared VM artifacts explicitly defined below.

## Fallback Rule

- If sub-agents are not supported, stop and return `sub_agent_not_supported`.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.
6) Set VMID policy env vars for this suite: `PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000` (or approved override).
7) Before running action prompts, execute bootstrap prompt `../e2e/BOOTSTRAP-UBUNTU24-WITH-AGENT-TEMPLATE.prompt.md` once and ensure `build/ubuntu-24-with-agent.vm-template.id` exists.
8) Resolve `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id` and resolve `SHARED_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
9) Allocate one fresh in-range `SHARED_VMID` via `get_next_vmid`, then clone once: `clone_template --wait --full 0` from `TEMPLATE_VMID` to `SHARED_VMID` on `SHARED_NODE`.
10) Persist shared inputs for sub-agents:
   - `build/phase3-cloudinit-qga.shared-node`
   - `build/phase3-cloudinit-qga.shared-vmid`
11) Independent-VM prompts (A17*) must still resolve their own `TEST_VMID` and self-destroy assets inside each prompt run.
12) No-VM prompts (A22/A27) must not create/stop/destroy VMs.
13) For disposable clones, prefer linked clone (`full=0`, default) to minimize storage I/O.
14) After all prompts finish (success or failure), run suite teardown once: stop and destroy `SHARED_VMID`, then remove shared input files.

## Prompt Files to Execute

- Independent-VM prompts:
  - `A17-agent_exec.prompt.md`
  - `A17-agent_exec_status.prompt.md`
- Shared-VM prompts:
  - `A15-agent_network_get_interfaces.prompt.md`
  - `A20-dump_cloudinit.prompt.md`
- No-VM prompts:
  - `A22-storage_upload_guard.prompt.md`
  - `A27-render_and_serve_seed.prompt.md`

## Final Output Format

Return one JSON object:

```json
{
  "suite": "phase-3-cloudinit-qga",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 6,
    "failed": 0
  },
  "results": [
    {
      "action": "agent_network_get_interfaces",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
