# Phase 4 Suite Runner

## Purpose

Run all Phase 4 action prompts under `applications/proxmox-cli/src/test/prompts/phase-4-console-websocket/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A18, A19, A29, A32, A34, A40).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- Every prompt must be order-independent and self-contained.
- Do not reuse VMIDs across prompt files.

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
8) Each action prompt must clone from the template VMID in that file to a fresh `TEST_VMID`, then self-check and self-destroy VM assets.
9) If websocket setup fails, return actionable diagnostics and still clean up the disposable VM.
10) For disposable clones, prefer linked clone (`full=0`, default) to minimize storage I/O.

## Prompt Files to Execute

- `A18-start_vnc_proxy.prompt.md`
- `A19-connect_vnc_websocket.prompt.md`
- `A29-open_vm_termproxy.prompt.md`
- `A32-validate_k1_serial_readable.prompt.md`
- `A34-serial_ws_session_control.prompt.md`
- `A40-validate_serial_output_criterion2.prompt.md`

## Final Output Format

Return one JSON object:

```json
{
  "suite": "phase-4-console-websocket",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 6,
    "failed": 0
  },
  "results": [
    {
      "action": "start_vnc_proxy",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
