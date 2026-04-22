# Phase 3 Suite Runner

## Purpose

Run all Phase 3 action prompts under `applications/proxmox-cli/src/test/prompts/phase-3-cloudinit-qga/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A15, A17-exec, A17-status, A20, A22, A27).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- Every prompt must be order-independent and self-contained.
- Do not reuse VMIDs across prompt files.
- `agent_exec` mutates guest state; keep effective VM write concurrency = 1 for VM mutating prompts.

## Fallback Rule

- If sub-agents are not supported, stop and return `sub_agent_not_supported`.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.
6) Set VMID policy env vars for this suite: `PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000` (or approved override).
7) For A15/A17 prompts, each sub-agent must provision its own disposable VM in-range (clone from a local Ubuntu template), enable `agent=1`, boot/reboot, then bootstrap qga inside guest.
8) QGA bootstrap must use Phase 4 minimal console actions (`open_vm_termproxy` + `serial_ws_session_control`) to run install/enable commands inside guest.
9) If console bootstrap path or qga readiness probe fails, return `blocked_qga_unavailable` with setup diagnostics and still destroy the disposable VM.

## Prompt Files to Execute

- `A15-agent_network_get_interfaces.prompt.md`
- `A17-agent_exec.prompt.md`
- `A17-agent_exec_status.prompt.md`
- `A20-dump_cloudinit.prompt.md`
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
