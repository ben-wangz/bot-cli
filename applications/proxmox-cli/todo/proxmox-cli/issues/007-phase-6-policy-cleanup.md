# ISSUE-007 Phase 6 Policy and Cleanup

- status: open
- priority: medium
- phase: 6
- depends_on: ISSUE-006

## Goal

完成策略类 action 产品化与清理动作，收敛 A01-A43 全量覆盖。

## Actions

- A33 record_mcp_gap
- A35 disable_blind_grub_injection
- A43 cleanup_obsolete_artifact

## Tasks

- [ ] 固化策略类 action 输出结构（reason/evidence/recommendation）。
- [ ] 实现 mcp gap 结论输出。
- [ ] 实现禁用盲注 GRUB 的策略动作输出。
- [ ] 实现 obsolete artifact 清理与幂等处理。
- [ ] 为 3 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 3 个 action 输出可机读且稳定。
- [ ] A43 可重复执行不报错。
- [ ] 3 条 prompt 通过。
