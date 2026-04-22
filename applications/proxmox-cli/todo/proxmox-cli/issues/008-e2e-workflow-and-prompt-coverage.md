# ISSUE-008 E2E Workflow and Prompt Coverage

- status: open
- priority: high
- phase: cross
- depends_on: ISSUE-001, ISSUE-002, ISSUE-003, ISSUE-004, ISSUE-005, ISSUE-006, ISSUE-007

## Goal

完成 Ubuntu24 串口自动安装闭环 workflow 验证，并收敛 43 条 action prompt 覆盖。

## Scope

- workflow: `ubuntu24-serial-autoinstall`
- action prompts: A01-A43（每 action 1 条独立正向用例）
- cleanup: 每次回归自动清理 VM 与临时镜像/seed

## Tasks

- [ ] 生成 workflow e2e 测试 prompt。
- [ ] 汇总 A01-A43 prompt 覆盖清单。
- [ ] 执行全链路：ISO -> autoinstall -> qga -> ssh hostname。
- [ ] 验证回归资源约束（最多 1 台 VM，2C/4G/32G）。
- [ ] 回归后清理测试资产并产出报告。
- [ ] 将安装介质方案从“整盘重打包 ISO”切回“小 shim + HTTP 安装源”，并纳入 workflow 可重复执行。

## Small Shim + HTTP Source Breakdown

- [ ] 定义 shim 产物验收标准（体积显著小于官方 ISO，且保留串口启动与 autoinstall 参数注入能力）。
- [ ] 设计并固化安装源 HTTP 服务路径（目录结构、访问地址、生命周期管理）。
- [ ] 在 workflow 中拆分“shim 引导介质”和“安装源 URL”两个输入，并明确失败诊断输出。
- [ ] 增加对内核启动参数的显式校验（`autoinstall`、NoCloud、serial console、HTTP source）。
- [ ] 增加 live 回归步骤：后台 `serial_ws_capture_to_file` 持续采集并按日志信号提前结束。
- [ ] 为“小 shim + HTTP 安装源”新增独立 e2e prompt，与“大 ISO 单盘自包含”回归路径分离。
- [ ] 完成一次端到端验收（安装成功、QGA 可用、SSH hostname 成功、资产清理完成）。

## Pending Design Decision

- 当前可稳定路径为“大 ISO 单盘自包含”；该路径用于阶段性回归，但不满足最初 shim 体积目标。
- 已确认后续目标：回到“小 shim + HTTP 安装源”实现方式，并以此作为 E2E 主路径验收标准。

## Acceptance

- [ ] 43 条 action prompt 全覆盖。
- [ ] E2E workflow 成功返回 vmid/ip/hostname。
- [ ] 回归结束后无遗留测试 VM 与临时文件。
- [ ] 小 shim 路径成为主验收路径；大 ISO 路径仅保留为临时诊断 fallback。
