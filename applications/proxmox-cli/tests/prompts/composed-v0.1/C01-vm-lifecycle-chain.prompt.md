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
1) `clone_template --wait --full 0 --source-vmid TEMPLATE_VMID --target-vmid TEST_VMID --pool "$PVE_POOL" --name "botcli-c01-${TEST_VMID}"`.
2) `update_vm_config` with safe update (e.g., description/name).
3) `review_install_tasks`.
4) `sendkey --key ret`.
5) Start migration with `migrate_vm --online 1 --with-local-disks 1` (capture `upid`), then observe for up to 5 minutes:
   - poll `get_task_status --node <source-node> --upid <upid>` every 10s;
   - if task reaches `status=stopped` and `exitstatus=OK`, mark success;
   - if 5 minutes elapsed and task is still progressing, treat as success when recent task progress looks normal (no explicit error/failure signal).

Validation:
- Every capability returns `ok == true`.
- For `clone_template`: keep strict async completion (`status=stopped` and `exitstatus=OK`).
- For `migrate_vm`: success is either
  - strict completion within 5 minutes (`status=stopped` and `exitstatus=OK`), or
  - after 5 minutes, task still in progress with no explicit error/failure signal in observed task progress.

Cleanup:
- Default: cleanup runs in `finally` (stop best-effort + destroy VM).
- Use `capability destroy_vm --node <latest node for TEST_VMID> --vmid TEST_VMID --if-missing ok --purge 1 --destroy-unreferenced-disks 1` for destroy step.
- Exception (preserve for migration verification): if migrate did not reach terminal `exitstatus=OK` within observation window but still shows healthy in-progress signals (no explicit failure), skip cleanup intentionally.
- When cleanup is skipped by this exception, return explicit preserve metadata for operator follow-up:
  - `preserve_for_validation: true`
  - `preserved_vmid: TEST_VMID`
  - `preserved_node: <latest node for TEST_VMID>`
  - `cleanup_skipped_reason: migration_still_progressing_without_failure_signal`

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
- Include `cleanup` details (`performed/skipped`, reason, vmid/node when preserved).
```
