# Phase 4 Suite Runner (SSH Control Plane)

## Purpose

Run all Phase 4 SSH-control-plane action prompts under `applications/proxmox-cli/src/test/prompts/phase-4-console-websocket/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A44-A51).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- Prompts are split into three execution classes:
  - Independent-VM prompts: A45, A47 (each prompt provisions and destroys its own VM assets).
  - Shared-VM prompts: A44, A46, A49 (consume one suite-level shared VM).
  - No-VM prompts: A48, A50, A51 (must not create/stop/destroy VMs in prompt body).
- Shared-VM prompts must run sequentially (shared VM write concurrency = 1).
- Independent-VM prompts may run concurrently up to the suite concurrency limit when infra allows.
- No-VM prompts may run concurrently, but A50/A51 must run after A49 because they depend on tunnel pid artifacts.
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
10) Start `SHARED_VMID` and resolve one reachable IPv4 as `SHARED_GUEST_IP` via `agent_network_get_interfaces`.
11) Generate one suite-level SSH key pair for shared prompts and inject pubkey via `ssh_inject_pubkey_qga`.
12) Persist shared inputs for sub-agents:
   - `build/phase4-ssh-control-plane.shared-node`
   - `build/phase4-ssh-control-plane.shared-vmid`
   - `build/phase4-ssh-control-plane.shared-guest-ip`
   - `build/phase4-ssh-control-plane.shared-identity-file`
13) Run A49 before A50/A51 and persist tunnel pid file path to:
   - `build/phase4-ssh-control-plane.shared-tunnel-pid-file`
14) Independent-VM prompts (A45/A47) must still resolve their own `TEST_VMID` and self-destroy assets inside each prompt run.
15) No-VM prompts (A48/A50/A51) must not create/stop/destroy VMs.
16) For disposable clones, prefer linked clone (`full=0`, default) to minimize storage I/O.
17) After all prompts finish (success or failure), run suite teardown once: stop and destroy `SHARED_VMID`, remove shared input files, and ensure tunnel process is not left running.

## Prompt Files to Execute

- Independent-VM prompts:
  - `A45-ssh_inject_pubkey_qga.prompt.md`
  - `A47-ssh_scp_transfer.prompt.md`
- Shared-VM prompts:
  - `A44-ssh_check_service.prompt.md`
  - `A46-ssh_exec.prompt.md`
  - `A49-ssh_tunnel_start.prompt.md`
- No-VM prompts:
  - `A48-ssh_print_connect_command.prompt.md`
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
