# Bootstrap Ubuntu24 Agent Template

## Purpose

Create a fresh Ubuntu 24 VM template (from zero) with guest-agent enabled, and persist template VMID to:

- `build/ubuntu-24-with-agent.vm-template.id`

This prompt must be executed before action-level suites that require disposable VMs.

## Prompt

```text
You are a coding/test execution agent. Prepare the shared Ubuntu24-with-agent template baseline for prompt suites.

Execution requirements:
1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Set VMID policy bounds explicitly:
   - `PVE_ALLOWED_VMID_MIN=1001`
   - `PVE_ALLOWED_VMID_MAX=2000`
4) Use API base: `${PVE_API_BASE_URL%/}/api2/json`.
5) Resolve inputs for workflow execution:
   - Resolve `NODE` from the first online node in `action list_nodes` output.
   - Resolve `TARGET_VMID` with this priority:
     1. if `build/ubuntu-24-with-agent.vm-template.id` exists and is non-empty, use that value
     2. otherwise use `1300`
6) Execute workflow:
   - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json workflow ubuntu24-with-agent-template --node "$NODE" --target-vmid "$TARGET_VMID"`
7) Validate workflow result:
   - `ok == true`
   - `result.template_vmid == TARGET_VMID` and in allowed range
   - `result.template_id_path` exists
8) Validate output file:
   - read `build/ubuntu-24-with-agent.vm-template.id`
   - content equals `result.template_vmid`
9) Return only structured result:
   - workflow
   - command
   - success
   - template_vmid
   - template_id_file
   - diagnostics

Safety:
- Do not modify or delete non-test VMs.
- If workflow fails after creating intermediate VM, report VMID for manual cleanup.
```
