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

## Acceptance

- [ ] 43 条 action prompt 全覆盖。
- [ ] E2E workflow 成功返回 vmid/ip/hostname。
- [ ] 回归结束后无遗留测试 VM 与临时文件。
