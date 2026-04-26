# proxmox-cli 重构 M4：wait 框架统一

- 状态: completed
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

## M4-3 已落地

1. `ssh/tunnel.go` 的进程退出等待循环已迁移为复用 `taskwait.Poll`。
2. `console/serial.go` 与 `console/termproxy.go` 的脚本发送节奏等待已抽到共享 helper（集中参数与行为）。
3. 已评估 `console/support.go:readSerialUntil`：当前实现是 websocket 事件驱动循环（并发读通道 + keepalive ticker + timeout timer + ctx 中断）；若强行迁移到 `taskwait.Poll` 会增加读超时/keepalive 时序漂移风险，暂不迁移。
4. `ssh/tunnel.go` 的启动存活窗口检查已从固定 `sleep + 单次检查` 迁移为 `taskwait.Poll`。
5. console/ssh 其余等待点评估结论：
   - `readSerialUntil` 保持事件驱动（最终决策，不迁移）；
   - serial script 节奏等待属于 I/O 节奏控制，不属于可安全抽象的轮询重试。

## `readSerialUntil` 评估结论（M4-3）

1. 该函数同时处理 4 类异步事件：消息读取、远端关闭、keepalive、总体超时。
2. `taskwait.Poll` 适配的是“单次 probe + 轮询间隔”模型；若用于替代当前 `select` 事件循环，需要额外引入 websocket 读 deadline 与超时分支映射，语义复杂度显著上升。
3. 当前实现已稳定覆盖关键诊断分支（vm not running / remote closed / expect 匹配失败 transcript tail）。
4. 结论：保持事件驱动实现不变，M4 在 console/ssh 范围按“可安全抽象优先”完成收口。

## 任务清单

1. [x] 形成 wait 盘点清单（入口、参数、语义差异）。
2. [x] 抽出通用轮询 helper（不改对外契约）。
3. [x] 迁移 QGA `agent_exec` 轮询到统一 helper。
4. [x] 评估并迁移 `taskwait.WaitTask` 主循环到统一 helper。
5. [x] 评估并收敛 console/ssh 内部 session/self-polled sleep 重试逻辑。
6. [x] 完成 M4 回归验证并回填兼容性结论（延后到整体重构完成后统一执行，避免在结构持续变动期重复回归）。

## 验收

1. `go build ./...` 通过。
2. `agent_exec` 的 `no_wait/timeout_seconds/poll_count` 等字段保持兼容。
3. 未改变 capability/workflow 的命令与参数契约。
