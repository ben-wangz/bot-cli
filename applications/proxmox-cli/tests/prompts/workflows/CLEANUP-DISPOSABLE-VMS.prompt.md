# Cleanup disposable VMs

## Goal

Remove disposable VMs left by regression/composed runs so pool state remains clean.

## Prompt

```text
You are a test execution agent. Clean up disposable VMs created by prompt suites.

Execution requirements:
1) Work from repository root.
2) Load `build/pve-user.env`.
3) Use API base `${PVE_API_BASE_URL%/}/api2/json`.
4) Read optional preserve list from env var `PRESERVE_VMIDS` (comma-separated VMID list). These VMIDs must be skipped.
5) Discover all VMs via `capability list_cluster_resources --type vm`.
6) Keep template VMID from `build/ubuntu-24-with-agent.vm-template.id` (never delete template VM).
7) Select cleanup candidates where all conditions are true:
   - VM belongs to current pool (`pool == $PVE_POOL`),
   - VM is not template (`template != 1`),
   - VMID is not in `PRESERVE_VMIDS`,
   - VM name starts with one of: `botcli-c00-`, `botcli-c01-`, `botcli-c02-`, `botcli-c03-`, `botcli-c04-`.
8) For each candidate VM:
    - best-effort stop with `capability vm_power --mode stop`;
    - destroy VM with `capability destroy_vm --node <node> --vmid <vmid> --if-missing ok --purge 1 --destroy-unreferenced-disks 1` using current user credentials;
    - if VM already missing, treat as success.
9) Re-read `list_cluster_resources --type vm` and verify no candidate VM remains.
10) Return JSON only:
   - `workflow`: `cleanup-disposable-vms`
   - `success`
   - `attempted`
   - `destroyed`
   - `remaining`
   - `remaining_vmids`
   - `preserved_vmids`
   - `diagnostics`

Safety:
- Never delete the template VM from `build/ubuntu-24-with-agent.vm-template.id`.
- Never delete VMs outside current `PVE_POOL`.
```
