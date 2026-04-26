# C01 vm lifecycle chain

## Goal

Use one disposable VM lifecycle chain to cover:

- `clone_template`
- `update_vm_config`
- `review_install_tasks`
- `sendkey`
- `migrate_vm`

## Prompt

```text
You are a test execution agent. Execute vm lifecycle chain in one run.

Setup:
1) Load `build/pve-user.env`, switch to `applications/proxmox-cli/src`.
2) Resolve `TEMPLATE_VMID` from `build/ubuntu-24-with-agent.vm-template.id`.
3) Resolve source node by `list_cluster_resources --type vm`, then allocate a free in-range `TEST_VMID`.
4) Resolve a target node different from source for migration; if none, return `no_migrate_target_node`.

Chain:
1) `clone_template --wait --full 0 --source-vmid TEMPLATE_VMID --target-vmid TEST_VMID`.
2) `update_vm_config` with safe update (e.g., description/name).
3) `review_install_tasks`.
4) `sendkey --key ret`.
5) `migrate_vm --wait` to target node.

Validation:
- Every capability returns `ok == true`.
- Async capabilities finish with final task `status=stopped` and `exitstatus=OK`.

Cleanup:
- Stop and destroy TEST_VMID on final host.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
