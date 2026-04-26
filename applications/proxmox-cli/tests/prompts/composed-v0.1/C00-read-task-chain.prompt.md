# C00 read task chain

## Goal

Run one read/task chain to cover:

- `list_nodes`
- `list_vms_by_node`
- `get_effective_permissions`
- `get_next_vmid`
- `list_tasks_by_vmid`
- `get_task_status`

## Prompt

```text
You are a test execution agent. Execute read/task chain with deterministic setup.

Setup:
1) Load `build/pve-user.env`, switch to `applications/proxmox-cli/src`.
2) Resolve `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
3) Resolve `SOURCE_NODE` from `list_cluster_resources --type vm` by `TEMPLATE_VMID`.
4) Resolve `TEST_VMID` via `get_next_vmid`.
5) Create one disposable VM by `clone_template --wait --full 0 --source-vmid TEMPLATE_VMID --target-vmid TEST_VMID --pool "$PVE_POOL"`.
6) Trigger at least one fresh VM task for `TEST_VMID` (for stable task-chain assertions), e.g. start then stop once via `vm_power --wait`.

Chain:
1) `list_nodes` and pick one online node as `NODE`.
2) `list_vms_by_node --node NODE`.
3) `get_effective_permissions --path /pool/$PVE_POOL`.
4) `get_next_vmid` and ensure value is positive integer.
5) `list_tasks_by_vmid --node SOURCE_NODE --vmid TEST_VMID --source all`, resolve one `UPID`.
6) `get_task_status --node SOURCE_NODE --upid UPID`.

Validation:
- All chain capabilities return `ok == true`.
- `list_nodes` has at least one node entry.
- `list_vms_by_node` returns JSON array.
- `get_effective_permissions` returns non-empty object for requested path.
- `list_tasks_by_vmid` contains at least one task with non-empty `upid`.
- `get_task_status` returns terminal or running task status payload for `UPID`.

Cleanup:
- Stop and destroy `TEST_VMID` on `SOURCE_NODE`.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
