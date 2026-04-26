# proxmox-cli 重构 M5：lint 行数限制治理

- 状态: completed
- 日期: 2026-04-26
- 依赖: `refactor-m4-wait-unification.md`

## 目标

为 `applications/proxmox-cli/src/**/*.go` 引入行数上限检查（<= 250），并分批治理当前超限文件。

## 当前基线（M5-0）

当前超限文件（>250 行）：

1. `internal/cli/app.go` (599)
2. `internal/pveapi/client.go` (311)
3. `internal/workflow/provision_template.go` (293)
4. `internal/capability/console/support.go` (264)
5. `internal/capability/ssh/tunnel.go` (261)
6. `internal/capability/ssh/support.go` (252)

## M5-1 已落地

1. 在 `lint.yaml` 增加 proxmox-cli Go 文件行数检查命令。
2. 首轮采用“基线放行 + 新增超限阻断”策略：
   - 基线清单内文件暂时允许超限；
   - 非基线文件若超过 250 行立即 lint 失败。

## M5-2 已落地

1. `internal/workflow/provision_template.go` 已拆分至 <=250 行（当前 206 行）。
2. 相关辅助逻辑迁移至 `internal/workflow/provision_template_support.go`。
3. 行数 lint 基线清单已移除 `internal/workflow/provision_template.go`。

## M5-3 已落地

1. `internal/capability/console/support.go` 已拆分至 <=250 行（当前 227 行）。
2. 调试辅助函数迁移至 `internal/capability/console/support_debug.go`。
3. 行数 lint 基线清单已移除 `internal/capability/console/support.go`。

## M5-4 已落地（部分）

1. `internal/capability/ssh/support.go` 已拆分至 <=250 行（当前 244 行）。
2. `shellJoin` 迁移至 `internal/capability/ssh/command_join.go`。
3. 行数 lint 基线清单已移除 `internal/capability/ssh/support.go`。

## M5-5 已落地

1. `internal/capability/ssh/tunnel.go` 已拆分至 <=250 行（当前 242 行）。
2. 文件 I/O 与清理辅助逻辑迁移至 `internal/capability/ssh/tunnel_io.go`。
3. 行数 lint 基线清单已移除 `internal/capability/ssh/tunnel.go`。

## M5-6 已落地（部分）

1. `internal/pveapi/client.go` 已拆分至 <=250 行（当前 220 行）。
2. multipart 与 websocket 逻辑分别迁移至 `internal/pveapi/client_multipart.go` 与 `internal/pveapi/client_websocket.go`。
3. 行数 lint 基线清单已移除 `internal/pveapi/client.go`。

## M5-7 已落地

1. `internal/cli/app.go` 已拆分至 <=250 行（当前 35 行）。
2. 主入口与职责拆分至 `internal/cli/app_run.go`、`internal/cli/app_flags.go`、`internal/cli/app_commands.go`。
3. 行数 lint 基线放行清单已清空（`baseline_allow = set()`），规则切换为严格模式。

## 任务清单

1. [x] 引入行数 lint 规则（覆盖 proxmox-cli Go 文件）。
2. [x] 记录并固化首轮超限基线清单。
3. [x] 拆分 `internal/workflow/provision_template.go` 到 <=250 行。
4. [x] 拆分 `internal/capability/console/support.go` 到 <=250 行。
5. [x] 拆分 `internal/capability/ssh/tunnel.go` 与 `internal/capability/ssh/support.go` 到 <=250 行。
6. [x] 拆分 `internal/pveapi/client.go` 与 `internal/cli/app.go` 到 <=250 行。
7. [x] 清空基线放行清单，切换为严格模式（所有 Go 文件 <=250 行）。

## 验收

1. lint 能阻断新增超限文件。
2. 现有超限文件按批次清零。
3. 最终无基线放行项，规则进入长期严格执行。
