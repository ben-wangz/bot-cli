# proxmox-cli v0.1 清理候选清单

- 状态: draft
- 优先级: medium
- 关联计划: `development-plan-v0.1.md`（M4）

## 1. 候选判定标准

1. 不在当前需求主路径（Phase 1-5 + 两条 e2e workflow）内。
2. 未被回归入口引用，或仅为历史兼容说明。
3. 删除后可由现有能力替代，且不会破坏当前回归。

## 2. 首批候选

### C1 `node_termproxy_shell_exec`

- 证据
  - 在 ISSUE-006 中已标记为“历史实现保留，不纳入后续验收目标”。
  - 现有主路径由 `provision-template-from-artifact` + SSH/QGA action 覆盖。
- 风险
  - 可能仍被少量手工排障脚本使用。
- 建议
  - v0.1 先标记 deprecated（help 文案 + 文档），v0.2 再删除实现。

### C2 `storage_upload_snippet`

- 证据
  - Live 环境中 snippets 不在 upload API 可上传类型集合内；主路径已改为 `storage_upload_guard` + ISO 上传链路。
  - 目前两条 workflow 均不依赖该 action。
- 风险
  - 若后续引入 root 落盘路径，可能再次需要。
- 建议
  - M4 做使用面审计（代码调用/文档引用/prompt 覆盖）后决定保留或删除。

## 3. 审计任务

1. 枚举候选 action 的调用点、help 暴露、测试覆盖。
2. 输出“删除影响面”报告（CLI 行为变化、文档变化、测试变化）。
3. 为每个删除动作提供替代路径或迁移说明。

## 4. 退出标准

1. 每个候选均有明确结论：保留/弃用/删除。
2. 删除项已完成代码、help、文档、prompt 同步。
3. 回归入口执行成功，无新增断档。
