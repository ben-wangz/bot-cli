# VM Lifecycle Playbook

This playbook is for disposable VM lifecycle operations: clone, configure, power, migrate, validate, cleanup.

## Typical Lifecycle Chain

1. Resolve template VMID and target VMID (`get_next_vmid` returns an ID in allowed range).
2. `clone_template --source-vmid <template> --target-vmid <vmid> --pool "$PVE_POOL"`
3. `update_vm_config` (name/description/resource knobs)
4. `vm_power` (start/stop/reboot)
5. Optional: `migrate_vm --online 1 --with-local-disks 1`
6. Validation checks (`get_vm_status`, task status polling)
7. Cleanup (`vm_power --mode stop` + `destroy_vm`)

## Task Completion Policy

For strict completion steps:

1. capture returned `upid`
2. poll `get_task_status`
3. require terminal `status=stopped` and `exitstatus=OK`

For long-running migration chains, use an observation window policy only when explicitly allowed by the test design.

## Safe Mutation Tips

1. Always set `--pool "$PVE_POOL"` for clone/create in scoped environments.
2. Never mutate unknown VMIDs discovered outside your assigned range.
3. Keep operation logs/JSON artifacts for failure diagnosis.

## Minimal Diagnostics Set

When a step fails, collect:

1. failing capability request args (without secret fields)
2. latest `get_task_status` result (if task-based)
3. `list_tasks_by_vmid` around failure time
4. `get_vm_status` and `get_vm_config` snapshot
