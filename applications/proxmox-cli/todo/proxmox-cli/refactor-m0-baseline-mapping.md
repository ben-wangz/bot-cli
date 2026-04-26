# proxmox-cli 重构 M0：基线冻结与映射清单

- 状态: completed
- 日期: 2026-04-26
- 基线 commit: `319931aa186fb2a82a4e7620e7087fb972284321`
- 范围: `applications/proxmox-cli/src`

## 1) 当前结构快照（按 phase）

主问题：业务代码按 `phase` 聚类，跨能力改动需要在多个 phase 文件跳转。

关键文件（按行数）：

1. `internal/action/phase4.go` (998)
2. `internal/action/phase3.go` (828)
3. `internal/cli/workflow.go` (696)
4. `internal/action/phase4_ssh.go` (692)
5. `internal/cli/app.go` (597)
6. `internal/action/phase2.go` (494)
7. `internal/action/phase5_acl.go` (348)
8. `internal/pveapi/client.go` (311)

说明：这些也是后续“Go 单文件 <= 250 行”治理的首批高风险点。

## 2) 旧 phase -> 新能力域映射（M0 草案）

### 2.1 Action 层映射

1. `phase1.go`（read/task）
   - 目标域: `domain/inventory` + `domain/task`
   - 典型 action: `list_nodes`, `list_cluster_resources`, `get_task_status`, `list_tasks_by_vmid`

2. `phase2.go`（vm lifecycle/config）
   - 目标域: `domain/vm`
   - 子能力: `lifecycle`（clone/migrate/power/template）, `config`（update/agent/boot/cdrom）, `install-diagnostics`（review/sendkey）

3. `phase3.go`（qga/cloud-init/storage upload）
   - 目标域: `domain/guest` + `domain/storage` + `platform/localiso`
   - 典型 action: `agent_exec*`, `storage_upload_*`, `build_ubuntu_autoinstall_iso`

4. `phase4.go` + `phase4_ssh.go`（console/ws + ssh）
   - 目标域: `domain/console`, `domain/ssh`
   - 子能力: websocket 会话、串口采集、SSH 执行、SCP、tunnel 生命周期

5. `phase5.go` + `phase5_acl.go`（root bootstrap/acl）
   - 目标域: `domain/access`
   - 子能力: user/pool bootstrap, ACL get/grant/revoke

6. `wait.go`（任务等待）
   - 目标域: `shared/wait`（统一等待框架入口之一）

### 2.2 CLI/编排层映射

1. `internal/cli/app.go`
   - 目标域: `cli/root`, `cli/action`, `cli/workflow`, `cli/auth`
   - 备注: 当前 `executeActionByPhase` 直接绑定 phase，需要替换为 capability registry

2. `internal/cli/workflow.go`
   - 目标域: `workflow/provision`, `workflow/bootstrap`, `workflow/shared`
   - 备注: 当前 workflow 与 action 调用/参数拼装混杂在单文件

3. `internal/cli/help.go`
   - 目标域: `cli/help`

## 3) wait 行为盘点（M0 输入）

wait 逻辑已分散在 action 与 workflow，多种实现并存：

1. 全局 `--wait` 路径
   - 位置: `internal/cli/app.go`
   - 机制: `applyActionWait` -> `action.WaitTask`

2. 通用任务等待
   - 位置: `internal/action/wait.go`
   - 机制: 轮询 task status + timeout + log tail

3. action 内自轮询
   - 位置: `internal/action/phase3.go`
   - 机制: `pollAgentExecStatus`（独立 timeout/poll 参数）

4. 会话/串口等待
   - 位置: `internal/action/phase4.go`
   - 机制: websocket 读循环 + 多处 sleep/retry

5. workflow 级等待编排
   - 位置: `internal/cli/workflow.go`
   - 机制: `runWorkflowAction(..., wait bool)`，按步骤显式开关 wait

结论：M4 统一 wait 框架需要覆盖“task wait + condition wait + session wait”三类模式，并保留现有输出/错误语义兼容。

## 4) 迁移顺序（先低耦合后高耦合）

1. Step A（低风险）
   - 拆出 `domain/access`（phase5）与 `domain/inventory`（phase1 read-only）

2. Step B（中风险）
   - 拆出 `domain/vm`（phase2）与 `domain/guest`（phase3 agent 部分）

3. Step C（中高风险）
   - 拆出 `domain/storage` + `platform/localiso`（phase3 upload/build）

4. Step D（高风险）
   - 拆出 `domain/console` + `domain/ssh`（phase4/phase4_ssh）

5. Step E（高风险）
   - workflow 编排层与 action registry 重构（去 phase 分发）

6. Step F（横切）
   - wait 统一框架落地并回迁各域实现

## 5) 风险点列表（M0）

1. 分发耦合风险
   - `executeActionByPhase` 依赖 `IsPhaseXAction`，重构时容易造成 action 漏注册。

2. 行数超限风险
   - 多个核心文件远超 250 行，若不先拆边界，lint 强化会一次性阻断开发。

3. wait 行为漂移风险
   - 当前存在多套等待语义，统一时容易引入 timeout 边界差异。

4. workflow 回归风险
   - `workflow.go` 与 action 交互紧耦合，迁移期间可能影响步骤诊断字段。

5. 兼容性风险
   - 需要确保错误码、diagnostics 字段、wait_skipped/wait_status 行为保持兼容。

## 6) M1 输入清单

1. 锁定新目录结构与 import 方向（domain/workflow/cli/platform/shared）。
2. 确定 action registry 方案（替代 phase 判定函数）。
3. 确定 wait 框架最小契约：
   - 输入: deadline/interval/backoff/condition
   - 输出: wait diagnostics（poll_count/elapsed/last_state）
   - 兼容: 保留现有字段与错误语义
