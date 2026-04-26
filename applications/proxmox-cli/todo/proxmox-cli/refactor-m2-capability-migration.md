# proxmox-cli 重构 M2：能力域搬迁与目录收敛

- 状态: completed
- 日期: 2026-04-26
- 依赖: `refactor-m1-target-structure.md`

## 目标

在不改变 CLI 契约与业务语义的前提下，完成从 phase/action 历史组织到能力域组织的搬迁，并把高耦合大文件进一步拆分到稳定目录。

## 已完成项

1. action 执行核心迁移到能力域目录：`internal/capability/`，并保留 registry 分发模型。
2. phase dispatcher 与 phase 入口历史文件已清理，`internal` 主目录不再以 phase 组织。
3. wait 与策略横切能力独立：
   - `internal/taskwait/`
   - `internal/policy/`
4. capability 内完成二级目录化：
   - `internal/capability/vm/`
   - `internal/capability/console/`
   - `internal/capability/ssh/`
   - `internal/capability/storage/`
5. root capability 包收敛为桥接与注册：
   - `internal/capability/registry.go`
   - `internal/capability/*_bridge.go`
6. 保持 action 名称与输出契约不变（`ok/action/scope/request/result/diagnostics`）。
7. VNC action 已从运行时移除（`start_vnc_proxy`、`connect_vnc_websocket`），并同步清理相关 prompt/doc 残留。

## 验证

1. 构建验证通过：`go build ./...`。
2. registry 分发稳定，CLI 入口无需变更外部调用方式。

## 备注

1. 本阶段以“搬迁优先、行为不变”为原则，不引入新需求。
2. workflow 编排层仍位于 `internal/cli/workflow.go`，留给 M3 收敛。

## M2 -> M3 交接建议

1. 新建 `internal/workflow/`，将编排逻辑从 CLI 剥离。
2. 统一 workflow step 执行骨架（参数校验、诊断、错误包装、恢复点）。
3. 清理 workflow 中残留 phase 语义命名。
