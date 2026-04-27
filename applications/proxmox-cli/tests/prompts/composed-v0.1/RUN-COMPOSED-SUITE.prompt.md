# Composed v0.1 Suite Runner

## Purpose

Run virtual-workflow prompt chains to cover capabilities that are not directly exercised by the two e2e workflows.

## Required Execution Mode

- If sub-agents are supported, use sub-agents.
- Spawn one sub-agent per composed prompt file (C00..C05).
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
8) Generate one suite run marker for disposable resources, for example `SUITE_RUN_ID="$(date +%Y%m%d-%H%M%S)-$$"`.

## Mandatory Cleanup Contract

1) Every VM-writing chain (C00..C04) must run cleanup in a `finally` block (or equivalent) even when validation fails midway.
2) Cleanup must include VM destruction (not stop-only):
   - stop/shutdown best-effort;
   - call `capability destroy_vm --if-missing ok --purge 1 --destroy-unreferenced-disks 1` using current test credentials;
   - treat "already absent" as success.
3) Allowed preserve exception:
   - only for `C01-vm-lifecycle-chain` when migration is still healthy/in-progress after observation window and no explicit failure signal;
   - chain must return `preserve_for_validation=true` with vmid/node/reason.
4) At end of suite, run one extra sweep:
   - gather all VMIDs recorded by sub-agents;
   - exclude VMIDs explicitly marked `preserve_for_validation=true` by C01;
   - retry destroy for any other VM still present;
   - return `cleanup` diagnostics including attempted/destroyed/remaining counts and `preserved_vmids`.

## Prompt Files to Execute

- `C00-read-task-chain.prompt.md`
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
    "passed": 6,
    "failed": 0
  },
  "cleanup": {
    "attempted": 5,
    "destroyed": 5,
    "remaining": 0,
    "preserved_vmids": []
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
