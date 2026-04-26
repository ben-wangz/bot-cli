# proxmox-cli 重构 M6：文档与收尾

- 状态: active
- 日期: 2026-04-26
- 依赖: `refactor-m5-lint-line-limit.md`

## 目标

完成重构收尾文档：更新模块导航、去除误导性 phase 入口描述、沉淀 wait 接口评估结论，并为最终回归验证提供执行清单。

## 范围

1. 更新模块 README，反映 capability/workflow 主结构与开发入口。
2. 更新 issue/todo 导航说明，明确 phase 命名文件仅作历史归档。
3. 形成重构回顾（范围、收益、后续优化项）。
4. 评估当前 wait 接口参数形态（cli/capability/workflow/taskwait），记录是否继续统一。
5. 规划并执行一轮完整回归验证，回填验证结论。

## M6-0 盘点

1. `README.md` 已有基础信息，但未给出重构后 internal 包职责导航。
2. `todo/proxmox-cli/issues/README.md` 仍以 phase 文件名列为常规 issue，缺少“归档/历史”定位说明。
3. wait 参数当前分层：
   - CLI：`--wait` 为布尔开关，异步 capability 通过 `taskwait.WaitTask` 等待。
   - capability：通过 `capability.Meta{Async, WaitSkipReason}` 表示是否可/应由 CLI 统一等待。
   - workflow：step 级 `wait bool` 由编排显式传递。
   - taskwait：`WaitOptions{Timeout, Interval}` + `PollOptions` 提供轮询骨架。

## M6-1 已落地（文档导航）

1. `applications/proxmox-cli/README.md` 增加 internal 包职责导航。
2. `applications/proxmox-cli/todo/proxmox-cli/issues/README.md` 标注 phase 命名 issue 为历史归档，并指向当前 `refactor-*.md` 作为执行入口。

## wait 接口评估结论（M6-2）

1. 现状接口已满足“分层清晰 + 契约稳定”：
   - CLI 负责是否触发等待；
   - capability 元数据负责等待适配/跳过原因；
   - taskwait 负责统一轮询技术实现；
   - workflow 负责编排层步骤级等待控制。
2. 当前不建议在 M6 再做接口收敛重构：
   - 若将 workflow/capability 全部收敛到单一参数对象，会引入命令契约和调用层改造，超出“文档与收尾”范围；
   - M4 已完成核心实现统一，继续改接口收益有限、行为漂移风险更高。
3. 后续可选优化（非 M6 阻断项）：
   - 引入内部 `WaitProfile`（仅内部类型），统一超时/间隔默认值来源；
   - 保持 CLI 参数与输出语义不变，仅减少内部默认值分散定义。

## 重构回顾（M6-3）

1. 主要改动范围：
   - 目录组织：从 phase 历史组织迁移到 capability/workflow 主结构；
   - 执行路径：CLI 以 capability/workflow 为主入口，phase 语义退出主路径；
   - 等待能力：抽出 `taskwait.Poll` 并收敛高频 wait 实现；
   - 代码体量治理：引入并落地 Go 单文件 <=250 行强约束。
2. 已确认收益：
   - 可维护性提升：主路径命名与能力域一致，定位成本下降；
   - 复用性提升：wait 轮询骨架复用，减少重复实现；
   - 质量门槛提升：line-limit 进入 lint 严格模式，无基线豁免。
3. 后续优化清单（M6 之后）：
   - 统一梳理 wait 默认值定义位置，减少跨模块散落；
   - 在 prompts/docs 中继续清理历史 action/phase 叙述（保留归档语义）；
   - 结合回归结果决定是否推进内部 `WaitProfile` 轻量收敛。

## 任务清单

1. [x] 更新模块 README 的新结构导航与开发入口说明。
2. [x] 更新 issue/todo 导航，避免 phase 命名造成误导入口。
3. [x] 完成重构回顾文档（范围、收益、后续优化清单）。
4. [x] 回填 wait 接口参数评估与是否继续统一的决策。
5. [ ] 执行完整回归并回填 M3/M4/M6 验证结论。

## 验收

1. 文档入口与目录导航与现状实现一致。
2. 历史 phase 描述仅保留归档语义，不再作为当前执行入口。
3. wait 接口是否继续统一有明确记录与理由。
4. 完整回归结果有结论可追踪。
