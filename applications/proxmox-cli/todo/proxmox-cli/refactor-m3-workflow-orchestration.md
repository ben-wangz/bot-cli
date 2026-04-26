# proxmox-cli 重构 M3：workflow 编排层收敛

- 状态: completed
- 日期: 2026-04-26
- 依赖: `refactor-m2-capability-migration.md`

## 目标

在收敛 workflow 编排实现的前提下，完成内部语义统一（action/phase -> capability/step）：

1. 将编排逻辑从 `internal/cli/workflow.go` 拆分到 `internal/workflow/`。
2. 统一 step 执行骨架（参数校验、执行日志、错误包装、诊断输出）。
3. 清理 workflow 内部残留的 phase/action 语义命名。

## 约束

1. `workflow` 子命令与参数契约保持稳定。
2. 不引入新 workflow 功能。
3. 每批改动均可单独构建验证（`go build ./...`）。

## 拟拆分结构（草案）

1. `internal/workflow/engine.go`：workflow 分发入口。
2. `internal/workflow/bootstrap_user_pool_acl.go`：bootstrap 编排。
3. `internal/workflow/provision_template.go`：provision 编排。
4. `internal/workflow/helpers.go` / `logging.go` / `step_command_runner.go` / `storage_iso.go`：共享校验、日志、step 执行骨架与 ISO 校验桥接。

## 任务清单

1. [x] 抽取 workflow runtime 与 step 执行 helper。
2. [x] 迁移 `bootstrap-bot-user-pool-acl` 编排到 `internal/workflow/`。
3. [x] 迁移 `provision-template-from-artifact` 编排到 `internal/workflow/`。
4. [x] 清理 workflow 层 phase/action 残留命名。
5. [x] 保持 CLI 入口稳定，仅调整内部调用路径。

## 验收

1. [x] `go build ./...` 通过。
2. [x] `internal/cli/workflow.go` 仅保留薄入口，不再承载重编排逻辑。
3. [x] workflow 编排统一使用 step 语义（日志与诊断字段已去 action/phase 命名）。
4. [x] `capability` 命令面与输出契约已替代旧 action 语义（有意不保留兼容）。
5. [ ] workflow 回归脚本/场景验证（按当前决策暂缓执行）。
