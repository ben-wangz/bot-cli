# ISSUE-006 Phase 5 Privilege Ladder and Root Ops

- status: open
- priority: high
- phase: 5
- depends_on: ISSUE-005

## Goal

实现 root 专属能力路径与权限阶梯验证。

## Actions

- A36 enforce_args_permission_guard
- A37 extract_kernel_initrd_workaround
- A38 set_args_with_root_session
- A39 node_termproxy_shell_exec
- A41 revalidate_privilege_ladder
- A42 autoinstall_trigger_retest

## Tasks

- [ ] 实现 `args` 权限错误识别与引导。
- [ ] 实现 root session ticket 通道。
- [ ] 实现 node termproxy shell 执行器。
- [ ] 实现 kernel/initrd 提取与校验。
- [ ] 实现权限梯度复验报告输出。
- [ ] 为 6 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 6 个 action 可体现 user/root 差异。
- [ ] A38/A39 在 root 路径可成功执行。
- [ ] A41 产出可审计结论。
- [ ] 6 条 prompt 通过。
