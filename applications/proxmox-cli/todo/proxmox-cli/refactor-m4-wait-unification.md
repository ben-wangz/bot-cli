# proxmox-cli 重构 M4：wait 框架统一

- 状态: active
- 日期: 2026-04-26
- 依赖: `refactor-m3-workflow-orchestration.md`

## 目标

在不引入新功能的前提下，统一 capability/workflow 的等待实现骨架，减少分散轮询逻辑并保持现有错误语义与诊断字段稳定。

## 范围

1. 盘点当前 wait 入口、参数与行为差异。
2. 提供可复用轮询骨架（timeout/interval/context 中断）。
3. 先迁移高频链路一批（M4-1）：QGA `agent_exec` 状态轮询。
4. 后续批次迁移 task/wait 与 session/self-polled 场景。

## 当前盘点（M4-0）

1. CLI 全局 wait：`internal/cli/app.go` 通过 `--wait` 调用 `taskwait.WaitTask`，仅适用于 registry 标记 `Async=true` 的 capability。
2. task wait：`internal/taskwait/wait.go` 自带轮询循环（timeout、interval、tail log、status 摘要）。
3. QGA wait：`internal/capability/guest_agent_support.go` 的 `pollAgentExecStatus` 使用独立轮询循环（`timeout-seconds`、`poll-interval-ms`、`--no-wait`）。
4. console/ssh 内存在 session/self-polled 的 sleep/重试实现，尚未纳入统一骨架。

## M4-1 已落地

1. 新增通用轮询骨架：`internal/taskwait/poller.go`。
2. 将 `pollAgentExecStatus` 迁移为复用统一轮询骨架，保持原有 timeout/中断错误语义。

## M4-2 已落地

1. 将 `taskwait.WaitTask` 主循环迁移为复用 `taskwait.Poll`。
2. 保留 `wait_task timeout exceeded` 诊断增强语义（状态摘要 + task log tail）。

## 任务清单

1. [x] 形成 wait 盘点清单（入口、参数、语义差异）。
2. [x] 抽出通用轮询 helper（不改对外契约）。
3. [x] 迁移 QGA `agent_exec` 轮询到统一 helper。
4. [x] 评估并迁移 `taskwait.WaitTask` 主循环到统一 helper。
5. [ ] 评估并收敛 console/ssh 内部 session/self-polled sleep 重试逻辑。
6. [ ] 完成 M4 回归验证并回填兼容性结论。

## 验收

1. `go build ./...` 通过。
2. `agent_exec` 的 `no_wait/timeout_seconds/poll_count` 等字段保持兼容。
3. 未改变 capability/workflow 的命令与参数契约。
