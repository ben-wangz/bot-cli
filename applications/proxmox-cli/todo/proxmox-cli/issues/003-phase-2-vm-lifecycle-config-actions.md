# ISSUE-003 Phase 2 VM Lifecycle and Config Actions

- status: in_progress
- priority: high
- phase: 2
- depends_on: ISSUE-002

## Goal

实现 VM 生命周期与配置写操作，统一 UPID 与 `--wait` 行为。

## Actions

- A07 clone_template
- A08 migrate_vm
- A09 convert_vm_to_template
- A10 update_vm_config
- A13 vm_power
- A14 set_vm_agent
- A23 create_vm
- A24 attach_cdrom_iso
- A25 set_net_boot_config
- A26 start_installer_and_console_ticket
- A28 enable_serial_console
- A30 review_install_tasks
- A31 sendkey

## Tasks

- [x] 写操作统一返回 UPID 并支持 `--wait`。
- [x] 实现 `sendkey` 强制 PUT。
- [x] 实现 `net0`/`boot` URL 编码工具。
- [x] 为 13 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [x] 13 个 action 可稳定执行。
- [x] 异步写操作 `--wait` 可得最终状态。
- [x] A31 PUT 语义验证通过。
- [ ] 13 条 prompt 通过。

## Regression Note

- `migrate_vm` 改为轻量候选 VM（优先无盘）后，`--wait` 回归通过并未产生残留锁。
