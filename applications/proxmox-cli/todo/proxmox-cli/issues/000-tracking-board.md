# ISSUE-000 Tracking Board

- status: in_progress
- priority: high
- owner: ben.wangz

## Objective

维护 Phase 0-6 的全局进度、依赖关系和里程碑状态。

## Current Snapshot

- last_updated: 2026-04-19
- overall_progress: 2/8 issues completed
- action_coverage: 22/43
- prompt_coverage: 22/43
- blocker: ISSUE-003 live migrate task may hang and keep `lock: migrate` (current locked VMIDs: 106, 107 on eva003)

## Issue Status

| Issue | Phase | Status | Action/Prompt Target | Depends On |
| --- | --- | --- | --- | --- |
| ISSUE-001 | 0 | completed | foundation baseline | - |
| ISSUE-002 | 1 | completed | 9 actions / 9 prompts | ISSUE-001 |
| ISSUE-003 | 2 | in_progress | 13 actions / 13 prompts | ISSUE-002 |
| ISSUE-004 | 3 | open | 6 actions / 6 prompts | ISSUE-003 |
| ISSUE-005 | 4 | open | 6 actions / 6 prompts | ISSUE-004 |
| ISSUE-006 | 5 | open | 6 actions / 6 prompts | ISSUE-005 |
| ISSUE-007 | 6 | open | 3 actions / 3 prompts | ISSUE-006 |
| ISSUE-008 | cross | open | E2E + full prompt sweep | ISSUE-001..007 |

> 注：A17 包含 `agent_exec` 与 `agent_exec_status` 两个可执行动作，计入 2 条 prompt，总计保持 43。

## Milestones

- [x] M1: ISSUE-001 + ISSUE-002 完成
- [ ] M2: ISSUE-003 + ISSUE-004 完成
- [ ] M3: ISSUE-005 完成
- [ ] M4: ISSUE-006 + ISSUE-007 完成
- [ ] M5: ISSUE-008 完成（E2E 联调与回归）

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

- [ ] A01-A43 覆盖完成。
- [ ] 43 条 action prompt 覆盖完成并可执行。
- [ ] Ubuntu24 workflow 闭环验证通过。
