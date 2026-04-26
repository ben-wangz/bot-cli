# proxmox-cli 重构 M1：目标结构与包边界定稿

- 状态: completed
- 日期: 2026-04-26
- 依赖: `refactor-m0-baseline-mapping.md`

## 1) M1 目标

在不改变现有需求与 CLI 契约的前提下，确定后续迁移使用的稳定代码结构，明确包职责、依赖方向与迁移守则。

## 2) 目标目录结构（定稿草案）

```text
applications/proxmox-cli/src/internal/
  cli/
    root/             # 全局参数、命令分发
    command/          # action/workflow/auth/console 子命令入口
    help/

  action/
    registry/         # action 注册表与分发（替代 phase 判定）
    contract/         # action 请求/响应契约与通用校验适配

  domain/
    inventory/        # list_nodes / list_cluster_resources / list_vms...
    task/             # task status / task list / next vmid 相关
    vm/               # clone/migrate/power/config/template/boot
    guest/            # qga network/exec/status
    storage/          # storage guard/upload
    console/          # termproxy/vnc/ws/serial capture
    ssh/              # ssh exec/scp/tunnel
    access/           # root bootstrap + acl

  workflow/
    bootstrap/        # bootstrap-bot-user-pool-acl
    provision/        # provision-template-from-artifact
    shared/           # workflow step helper、错误包装、日志

  platform/
    pveapi/           # 现有 client 保留并逐步内聚
    localexec/        # 本地命令执行抽象（后续承接 ISO 构建等）
    websocket/        # websocket 连接与基础读写工具

  shared/
    apperr/
    output/
    auth/
    redact/
    wait/             # 统一等待框架（M4 落地）
```

说明：

1. `domain/*` 只表达业务能力，不携带 phase 语义。
2. `action/registry` 作为兼容层，确保 CLI action 名称与对外契约不变。
3. `workflow/*` 与 `domain/*` 分离，避免编排逻辑与 action 细节混杂。

## 3) 包职责定义

1. `cli/*`
   - 仅负责入参解析、调用编排层、输出渲染。
   - 不直接持有业务分支逻辑。

2. `action/registry`
   - 维护 `action_name -> handler` 映射。
   - 负责 action 是否支持、wait 元数据读取、统一分发错误。

3. `domain/*`
   - 每个能力域独立维护 action 实现。
   - 可依赖 `platform/*` 与 `shared/*`，不可反向依赖 `cli/*`。

4. `workflow/*`
   - 负责多 action 编排顺序、恢复点、步骤诊断。
   - 通过 action registry 或显式 domain service 组合，不绕过契约层。

5. `platform/*`
   - 封装外部系统交互（PVE API、websocket、local exec）。
   - 不包含业务语义决策。

6. `shared/*`
   - 提供通用模型、错误、输出、wait 框架与工具函数。

## 4) 依赖方向与禁止项

允许方向：

1. `cli -> action/registry | workflow | shared`
2. `workflow -> action/registry | domain | shared | platform`
3. `action/registry -> domain | shared`
4. `domain -> platform | shared`
5. `platform -> shared`

禁止项：

1. `domain` 依赖 `cli`。
2. `platform` 依赖 `domain` 或 `workflow`。
3. `workflow` 直接写 HTTP 细节（应经 `platform/pveapi` 或 domain 封装）。
4. 新增 `phaseX` 命名文件/包作为主入口。

## 5) 迁移守则（M2 开始执行）

1. 搬迁优先，不先改业务逻辑。
2. 每批仅迁移一个能力域，迁移后立即做回归验证。
3. 对外兼容字段保持稳定：`ok/action/workflow/request/result/diagnostics`。
4. phase 文件允许短暂保留为桥接层，但不得继续新增业务逻辑。
5. 新代码禁止放入 `phase*.go`，统一落新结构。

## 6) Action Registry 目标契约（供 M2 使用）

建议最小接口：

```go
type Handler func(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error)

type Meta struct {
    Async bool
    Capability string
}

func Dispatch(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, Meta, error)
```

约束：

1. `Meta.Async` 用于统一 wait 决策，不再通过 `IsPhaseXAction` 推断。
2. `Meta.Capability` 仅用于诊断与文档，不影响输出契约。

## 7) Wait 框架接口前置约定（与 M4 对齐）

M1 先锁定契约，不在本阶段做全量迁移：

1. 输入：`timeout`、`interval`、`max_attempts`（可选）、`condition`。
2. 输出诊断：`poll_count`、`elapsed_ms`、`last_state`、`timed_out`。
3. 兼容要求：保留现有 `wait_skipped`、`wait_status` 行为语义。

## 8) M1 完成标准

1. 目录结构、包职责、依赖方向得到确认。
2. 迁移守则可直接指导 M2 分批实施。
3. Action registry 与 wait 契约冻结为后续实现输入。

## 9) M1 代码落地进度

已完成（M1 子步 1）：

1. 引入 action registry 分发骨架：`applications/proxmox-cli/src/internal/action/registry.go`。
2. CLI action 分发改为通过 registry：`applications/proxmox-cli/src/internal/cli/app.go`。
3. 保持现有 action 行为不变（phase 实现仍保留，当前仅替换分发入口）。

已完成（M1 子步 2）：

1. 移除旧 phase 判定函数在主分发路径上的残余依赖（`IsPhaseXAction`、`IsActionAsync`、`WaitSkipReason` 已从执行路径清退）。
2. 等待决策改为 registry 元数据驱动（`Meta.Async` + `Meta.WaitSkipReason`）。

已完成（M1 子步 3）：

1. action capability 已写入 diagnostics（由 registry `Meta.Capability` 注入），便于后续按能力域观测与回归。
2. action help 已切换为 capability 视图（不再以 phase 作为主展示口径）。

收尾说明：

1. wait 框架“可执行接口定义”放入 M4 实现阶段完成（M1 保持契约冻结与分发改造完成）。
