# proxmox-cli 重构 M3：workflow 编排层收敛

- 状态: active
- 日期: 2026-04-26
- 依赖: `refactor-m2-capability-migration.md`

## 目标

在不改变 workflow 对外行为（命令名、参数名、输出结构、错误语义）的前提下，收敛 workflow 编排实现：

1. 将编排逻辑从 `internal/cli/workflow.go` 拆分到 `internal/workflow/`。
2. 统一 step 执行骨架（参数校验、执行日志、错误包装、诊断输出）。
3. 清理 workflow 内部残留的 phase 语义命名。

## 约束

1. CLI 契约保持兼容。
2. 不引入新 workflow 功能。
3. 每批改动均可单独构建验证（`go build ./...`）。

## 拟拆分结构（草案）

1. `internal/workflow/runner.go`：workflow 分发与 step 运行骨架。
2. `internal/workflow/bootstrap_user_pool_acl.go`：bootstrap 编排。
3. `internal/workflow/provision_template_from_artifact.go`：provision 编排。
4. `internal/workflow/common.go`：共享参数校验/日志/错误包装。

## 任务清单

1. [ ] 抽取 workflow runtime 与 step 执行 helper。
2. [ ] 迁移 `bootstrap-bot-user-pool-acl` 编排到 `internal/workflow/`。
3. [ ] 迁移 `provision-template-from-artifact` 编排到 `internal/workflow/`。
4. [ ] 清理 workflow 层 phase 残留命名。
5. [ ] 保持 CLI 入口不变，仅调整内部调用路径。

## 验收

1. `go build ./...` 通过。
2. workflow 命令对外 JSON 契约无变化。
3. `internal/cli/workflow.go` 仅保留薄入口，不再承载重编排逻辑。
