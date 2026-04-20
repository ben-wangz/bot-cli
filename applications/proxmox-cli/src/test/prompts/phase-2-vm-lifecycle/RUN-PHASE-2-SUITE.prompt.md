# Phase 2 Suite Runner

## Purpose

Run all Phase 2 action prompts under `applications/proxmox-cli/src/test/prompts/phase-2-vm-lifecycle/`.

## Required Execution Mode

- If sub-agents are supported, you **must** use sub-agents.
- Spawn one sub-agent per action prompt file (A07, A08, A09, A10, A13, A14, A23, A24, A25, A26, A28, A30, A31).
- Sub-agent concurrency must be <= 2.
- Each sub-agent executes exactly one prompt file and returns the result in the required schema.
- Every prompt must be order-independent and must not consume outputs from other prompt runs.
- Phase 2 actions are mutating operations; enforce a single active VM-mutating prompt at a time (effective VM write concurrency = 1).
- A prompt must never reuse a VMID produced by another prompt, even inside the same suite run.
- For async actions, prompt commands must include `--wait`, and validation must assert final task success (`status=stopped` and `exitstatus=OK`).

## Fallback Rule

- If sub-agents are not supported, stop and return `sub_agent_not_supported`.

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir: `applications/proxmox-cli/src`.
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Always include `--insecure-tls --output json`.
6) For mutating prompts, allocate a fresh `TEST_VMID` via `get_next_vmid` inside that prompt.
7) Teardown must run on both success and failure paths (best-effort cleanup).

## Prompt Files to Execute

- `A07-clone_template.prompt.md`
- `A08-migrate_vm.prompt.md`
- `A09-convert_vm_to_template.prompt.md`
- `A10-update_vm_config.prompt.md`
- `A13-vm_power.prompt.md`
- `A14-set_vm_agent.prompt.md`
- `A23-create_vm.prompt.md`
- `A24-attach_cdrom_iso.prompt.md`
- `A25-set_net_boot_config.prompt.md`
- `A26-start_installer_and_console_ticket.prompt.md`
- `A28-enable_serial_console.prompt.md`
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
    "passed": 13,
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
