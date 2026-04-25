# C01 vm lifecycle chain

## Goal

Use one disposable VM lifecycle chain to cover:

- `clone_template`
- `update_vm_config`
- `start_installer_and_console_ticket`
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
3) `start_installer_and_console_ticket`.
4) `review_install_tasks`.
5) `sendkey --key ret`.
6) `migrate_vm --wait` to target node.

Validation:
- Every action returns `ok == true`.
- Async actions finish with final task `status=stopped` and `exitstatus=OK`.

Cleanup:
- Stop and destroy TEST_VMID on final host.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
