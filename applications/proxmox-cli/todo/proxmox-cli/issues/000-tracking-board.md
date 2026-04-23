# ISSUE-000 Tracking Board

- status: in_progress
- priority: high
- owner: ben.wangz

## Objective

维护 Phase 0-6 的全局进度、依赖关系和里程碑状态。

## Current Snapshot

- last_updated: 2026-04-23
- overall_progress: 5/8 issues completed
- action_coverage: 47/51
- prompt_coverage: 46/51

## Issue Status

| Issue | Phase | Status | Action/Prompt Target | Depends On |
| --- | --- | --- | --- | --- |
| ISSUE-001 | 0 | completed | foundation baseline | - |
| ISSUE-002 | 1 | completed | 9 actions / 9 prompts | ISSUE-001 |
| ISSUE-003 | 2 | completed | 13 actions / 13 prompts | ISSUE-002 |
| ISSUE-004 | 3 | completed | 6 actions / 6 prompts | ISSUE-003 |
| ISSUE-005 | 4 | completed | 8 actions / 8 prompts | ISSUE-004 |
| ISSUE-006 | 5 | in_progress | 4 actions / 4 prompts | ISSUE-005 |
| ISSUE-007 | 6 | open | 3 actions / 3 prompts | ISSUE-006 |
| ISSUE-008 | cross | open | E2E + full prompt sweep | ISSUE-001..007 |

> 注：A17 包含 `agent_exec` 与 `agent_exec_status` 两个可执行动作，计入 2 条 prompt；ISSUE-005 新增 A44-A51 后总计为 51。

## Milestones

- [x] M1: ISSUE-001 + ISSUE-002 完成
- [x] M2: ISSUE-003 + ISSUE-004 完成
- [x] M3: ISSUE-005 完成
- [ ] M4: ISSUE-006 + ISSUE-007 完成
- [ ] M5: ISSUE-008 完成（E2E 联调与回归）

## Active Blockers

- ISSUE-004 已完成：Phase 3 的 6 个 action / 6 条 prompt 已通过正向回归。
- QGA 安装路径依赖已记录：当无 qga-ready 基础镜像时，仍需 Phase 4 串口路径保障来宾内安装与启用。
- ISSUE-005 已完成：SSH 控制面 A44-A51 已实现并通过回归（8/8）。
- A22 保持 guard 职责；seed/snippet 上传需独立 action 承担（待新 issue/action 编排）。
- 存储上传实测：PVE upload API 当前仅接受 `iso|vztmpl|import`，不接受 `snippets`，因此 cloud-init snippet 自动落盘需依赖 Phase 5 root 路径。
- Phase 5 范围已重审：root 仅用于一次性 user 授权 bootstrap；常规 VM 管理与回归不再依赖 root action。
- ISSUE-006 当前目标为 user+ACL 管理 4 个 action：创建用户 + 授权关系查询/新增/撤销（变更使用 delete+add 组合）。
- 新增问题：当前 `build_ubuntu_autoinstall_iso` 路径为“整盘重打包”，输出 ISO 体积接近官方源 ISO，偏离最初“小 shim”目标。
- 决策记录：后续回到“小 shim + HTTP 安装源”方案；当前大 ISO 路径作为临时稳定回归手段保留，待 ISSUE-008 编排替换。

## Dependencies

1. `001-phase-0-foundation.md`
2. `002-phase-1-read-task-actions.md`
3. `003-phase-2-vm-lifecycle-config-actions.md`
4. `004-phase-3-cloudinit-qga-actions.md`
5. `005-phase-4-console-websocket.md`
6. `006-phase-5-privilege-root-ops.md`
7. `007-phase-6-policy-cleanup.md`
8. `008-e2e-workflow-and-prompt-coverage.md`

## Working Rules

- 每完成一个 issue，立即回写本看板的 `Current Snapshot` 与 `Issue Status`。
- 每完成一个 action，同时更新 action/prompt 覆盖计数。
- 若依赖阻塞超过 1 天，在对应 issue 增加 blocker 说明并同步到此看板。
- 任何资源策略变更（VM 数量、规格、清理规则）必须先更新 ISSUE-008，再更新本看板。

## Exit Criteria

- [ ] A01-A51 覆盖完成。
- [ ] 51 条 action prompt 覆盖完成并可执行。
- [ ] Ubuntu24 workflow 闭环验证通过。
