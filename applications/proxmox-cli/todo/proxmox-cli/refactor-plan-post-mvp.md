# proxmox-cli 重构计划（Post-MVP，需求不变）

- 状态: active
- 优先级: high
- 创建日期: 2026-04-26
- 目标版本: v0.2（内部重构版本，不引入新需求）

## 1. 背景与问题

当前代码组织主要按 `phase` 演进历史分层，MVP 阶段可快速推进，但在后续维护中出现问题：

1. 目录和包边界按历史阶段而非稳定业务域划分，定位功能与变更影响面成本高。
2. 新增/修改 action 时容易跨 phase 跳转，认知负担大。
3. lint 目前缺少 Go 文件行数上限约束，代码体量增长后可读性和可维护性风险上升。

## 2. 重构目标（不改需求）

1. 在保持对外行为不变（命令、参数、输出、workflow 语义不变）的前提下，重组代码结构。
2. 将“按 phase 组织”迁移为“按稳定能力域组织”（示例：vm lifecycle、guest/qga、console/ws、ssh、acl/auth、workflow orchestration）。
3. 为 forgekit lint 增加 Go 文件行数限制：**单文件 <= 250 行**。
4. 保持现有回归入口和覆盖策略可用（setup -> workflows -> composed suite）。
5. 梳理并统一 `wait` 行为：为 action/workflow 提供可复用的等待框架（轮询、超时、重试、诊断输出）。

## 3. 非目标

1. 不新增产品功能。
2. 不调整 action/workflow 的业务语义。
3. 不引入额外发布流程复杂度。

## 4. 约束与验收口径

### 4.1 约束

1. 对外 CLI 契约不变：命令名、参数名、JSON 结构、错误码语义保持兼容。
2. 每次重构改动必须可回归验证，禁止一次性大搬迁后再集中修复。
3. Go 文件行数规则以 lint 强制执行（非人工约定）。
4. 现有 wait 语义保持兼容（默认超时、返回字段、错误语义不破坏）。

### 4.2 验收

1. 编译通过，核心回归通过。
2. phase 命名不再作为主组织方式（仅允许在兼容层短暂存在）。
3. lint 能稳定拦截超过 250 行的 `.go` 文件。
4. 文档更新到新结构，开发者可按目录快速定位能力边界。
5. wait 逻辑不再分散复制，新增 action/workflow 可复用统一等待框架。

## 5. 实施计划（建议分 7 个里程碑）

### M0: 基线冻结与映射清单

1. 冻结当前可工作的 commit 作为重构基线。
2. 建立“旧 phase 模块 -> 新能力域”映射表。
3. 输出迁移顺序（先低耦合域，后高耦合编排域）。

产出：`旧->新` 映射文档 + 风险点列表。

M0 执行结果：已完成，见 `applications/proxmox-cli/todo/proxmox-cli/refactor-m0-baseline-mapping.md`。

### M1: 目标结构定稿（目录与包边界）

1. 设计并评审新目录结构（以能力域为主，横切能力单独分层）。
2. 定义包职责：action handler、service、adapter/client、workflow orchestration。
3. 定义迁移规则：import 方向、禁止反向依赖、公共模型放置位置。

产出：结构草图 + 包职责说明 + 迁移守则。

M1 执行状态：completed，见 `applications/proxmox-cli/todo/proxmox-cli/refactor-m1-target-structure.md`。

### M2: 无行为变化的“搬迁优先”重构

1. 按能力域逐批搬迁代码（优先读类/低风险 action）。
2. 每批迁移后立即编译与回归，确保行为等价。
3. 保留最薄兼容层（如果需要），避免一次改动过大。

产出：分批 commit（每批可独立回滚）。

M2 执行状态：completed，见 `applications/proxmox-cli/todo/proxmox-cli/refactor-m2-capability-migration.md`。

### M3: workflow 编排层收敛

1. 将 workflow 入口与 orchestration 逻辑从 phase 语义中解耦。
2. 统一 workflow 共享步骤（输入校验、诊断、重试/恢复语义）。
3. 清理 phase/action 残留命名与桥接代码。

产出：稳定的 workflow 编排层与共享执行骨架。

M3 执行状态：completed，见 `applications/proxmox-cli/todo/proxmox-cli/refactor-m3-workflow-orchestration.md`。

### M4: wait 框架统一

1. 盘点所有 action/workflow 中的 wait 逻辑与参数（interval、timeout、最大轮询次数、恢复点）。
2. 抽象统一等待组件（建议含：condition、poller、deadline、backoff、diagnostics hook）。
3. 先迁移高频链路（如任务状态轮询、安装等待、串口等待），再覆盖其他等待场景。
4. 对外保持兼容：默认值、输出字段、失败语义不变；仅收敛实现。

产出：统一 wait 框架 + 首批迁移清单 + 兼容性验证结果。

M4 执行状态：completed，见 `applications/proxmox-cli/todo/proxmox-cli/refactor-m4-wait-unification.md`。

### M5: lint 强化（Go 文件行数 <= 250）

1. 在 `lint.yaml` 增加行数检查命令，覆盖 `applications/proxmox-cli/src/**/*.go`。
2. 明确统计口径（建议按物理行统计，自动忽略 `vendor/build/dist`）。
3. 对超限文件进行拆分，优先按职责拆为小函数/小文件。

产出：可执行 lint 规则 + 首轮超限文件治理。

### M6: 文档与收尾

1. 更新模块 README（新结构导航、开发入口）。
2. 更新 todo/issue 中与 phase 强绑定的陈述（保留归档历史，不保留误导入口）。
3. 形成重构回顾：改动范围、收益、后续优化清单。
4. 评估当前 wait 接口参数形态（capability/workflow/taskwait），判断是否需要进一步功能/接口统一重构，并记录决策。
5. 执行一轮完整回归测试，确认重构后功能正确性并沉淀验证结论。

产出：文档闭环 + 重构总结。

## 6. 当前目录现状与后续目标

当前已落地：

1. `internal/capability/...`：能力域实现与分发注册（含 `vm/`、`console/`、`ssh/`、`storage/` 子目录）。
2. `internal/taskwait/...`：统一任务等待能力（初版）。
3. `internal/policy/...`：策略与边界校验（如 VMID 范围）。
4. `internal/cli/...`：CLI 解析与命令分发入口。

M3 目标新增：

1. `internal/workflow/...`：workflow 编排层与共享步骤骨架。

## 7. 风险与缓解

1. 风险：重命名/搬迁导致隐式行为变化。
   - 缓解：先搬迁后重构，批次小、每批即回归。
2. 风险：行数限制一次性触发大量告警。
   - 缓解：先引入规则 + 允许短期待清单，再分批拆分。
3. 风险：旧 phase 语义残留导致双轨维护。
   - 缓解：设定兼容层截止里程碑（M3 结束前清理）。
4. 风险：wait 统一后出现行为漂移（超时边界、诊断字段差异）。
   - 缓解：建立 wait 兼容性回归用例，按 action/workflow 逐条对比旧行为。

## 8. 立即可执行的下一步

1. 启动 M5：落地 Go 文件行数 lint 规则与首轮超限治理。
2. 进入 M6：收敛文档与重构总结，回填 wait 接口参数形态评估。
3. 在整体重构完成后统一执行回归验证，并回填 M3/M4 验收结论。
