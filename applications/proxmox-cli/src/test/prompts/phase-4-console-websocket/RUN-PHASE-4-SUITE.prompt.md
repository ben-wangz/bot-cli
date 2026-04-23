# Phase 4 Suite Runner (SSH Control Plane)

## Purpose

Run all Phase 4 SSH-control-plane action prompts under `applications/proxmox-cli/src/test/prompts/phase-4-console-websocket/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A44-A51).
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
8) Each action prompt must be self-contained; if a prompt needs VM context, it must allocate and clean up its own disposable assets.
9) For disposable clones, prefer linked clone (`full=0`, default) to minimize storage I/O.

## Prompt Files to Execute

- `A44-ssh_check_service.prompt.md`
- `A45-ssh_inject_pubkey_qga.prompt.md`
- `A46-ssh_exec.prompt.md`
- `A47-ssh_scp_transfer.prompt.md`
- `A48-ssh_print_connect_command.prompt.md`
- `A49-ssh_tunnel_start.prompt.md`
- `A50-ssh_tunnel_status.prompt.md`
- `A51-ssh_tunnel_stop.prompt.md`

## Final Output Format

Return one JSON object:

```json
{
  "suite": "phase-4-ssh-control-plane",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 8,
    "failed": 0
  },
  "results": [
    {
      "action": "ssh_check_service",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
