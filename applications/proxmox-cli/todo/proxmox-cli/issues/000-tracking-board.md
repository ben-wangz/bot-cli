# ISSUE-000 Tracking Board

- status: in_progress
- priority: high
- owner: ben.wangz

## Objective

维护 Phase 0-6 的全局进度、依赖关系和里程碑状态。

## Current Snapshot

- last_updated: 2026-04-22
- overall_progress: 3/8 issues completed
- action_coverage: 35/43
- prompt_coverage: 34/43

## Issue Status

| Issue | Phase | Status | Action/Prompt Target | Depends On |
| --- | --- | --- | --- | --- |
| ISSUE-001 | 0 | completed | foundation baseline | - |
| ISSUE-002 | 1 | completed | 9 actions / 9 prompts | ISSUE-001 |
| ISSUE-003 | 2 | completed | 13 actions / 13 prompts | ISSUE-002 |
| ISSUE-004 | 3 | in_progress | 6 actions / 6 prompts | ISSUE-003 |
| ISSUE-005 | 4 | in_progress | 6 actions / 6 prompts | ISSUE-004 |
| ISSUE-006 | 5 | in_progress | 6 actions / 6 prompts | ISSUE-005 |
| ISSUE-007 | 6 | open | 3 actions / 3 prompts | ISSUE-006 |
| ISSUE-008 | cross | open | E2E + full prompt sweep | ISSUE-001..007 |

> 注：A17 包含 `agent_exec` 与 `agent_exec_status` 两个可执行动作，计入 2 条 prompt，总计保持 43。

## Milestones

- [x] M1: ISSUE-001 + ISSUE-002 完成
- [ ] M2: ISSUE-003 + ISSUE-004 完成
- [ ] M3: ISSUE-005 完成
- [ ] M4: ISSUE-006 + ISSUE-007 完成
- [ ] M5: ISSUE-008 完成（E2E 联调与回归）

## Active Blockers

- ISSUE-004 QGA 正向回归受环境阻塞：当前无 qga-ready 模板可用。
- 解除路径：提前落地 ISSUE-005 最小子集（A29 + A34）以通过串口在来宾内安装并启用 qemu-guest-agent。
- 当前进展：A18/A19/A29/A32/A34/A40 协议层已打通（含 keepalive 与串口输出判据校验）；当前阻塞为模板环境未稳定提供 VM 串口 login，仍不能完成来宾内 QGA 安装。
- A22 保持 guard 职责；seed/snippet 上传需独立 action 承担（待新 issue/action 编排）。
- 存储上传实测：PVE upload API 当前仅接受 `iso|vztmpl|import`，不接受 `snippets`，因此 cloud-init snippet 自动落盘需依赖 Phase 5 root 路径。
- Phase 5 进展：A39 `node_termproxy_shell_exec` 已接入并完成会话交互（脚本命令回显可验证）。
- 新增验证：A39 已支持 `cmd-opts`（null-terminated strings）并通过 root 会话实测 `upgrade/ceph_install` 提示链路。
- 新增约束：`root-token` 无法触发节点 `upgrade/ceph_install` 命令（后端仅允许 `root@pam` session），A39 命令执行闭环仍受凭据/协议双重阻塞。
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

- [ ] A01-A43 覆盖完成。
- [ ] 43 条 action prompt 覆盖完成并可执行。
- [ ] Ubuntu24 workflow 闭环验证通过。
