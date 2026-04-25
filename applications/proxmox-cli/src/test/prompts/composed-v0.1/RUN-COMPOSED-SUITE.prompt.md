# Composed v0.1 Suite Runner

## Purpose

Run virtual-workflow prompt chains to cover actions that are not directly exercised by the two e2e workflows.

## Required Execution Mode

- If sub-agents are supported, use sub-agents.
- Spawn one sub-agent per composed prompt file (C01..C05).
- Sub-agent concurrency must be <= 2.
- VM-writing chains must run sequentially (effective write concurrency = 1).

## Shared Setup

1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) For ACL chain only, additionally load `build/pve-root.env`.
4) Use source dir `applications/proxmox-cli/src`.
5) Use API base `${PVE_API_BASE_URL%/}/api2/json`.
6) Always include `--insecure-tls --output json`.
7) Set VMID policy env vars for this suite: `PVE_ALLOWED_VMID_MIN=1001`, `PVE_ALLOWED_VMID_MAX=2000`.

## Prompt Files to Execute

- `C01-vm-lifecycle-chain.prompt.md`
- `C02-qga-cloudinit-chain.prompt.md`
- `C03-serial-websocket-chain.prompt.md`
- `C04-ssh-control-plane-chain.prompt.md`
- `C05-acl-lifecycle-chain.prompt.md`

## Final Output Format

Return one JSON object:

```json
{
  "suite": "composed-v0.1",
  "mode": "sub-agent",
  "success": true,
  "summary": {
    "passed": 5,
    "failed": 0
  },
  "results": [
    {
      "chain": "vm-lifecycle",
      "command": "...",
      "success": true,
      "key_result": "...",
      "diagnostics": "..."
    }
  ]
}
```
