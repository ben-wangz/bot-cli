# proxmox-cli v0.1 清理候选清单

- 状态: completed
- 优先级: medium
- 关联计划: `development-plan-v0.1.md`（M4）

## 1. 候选判定标准

1. 不在当前需求主路径（Phase 1-5 + 两条 e2e workflow）内。
2. 未被回归入口引用，或仅为历史兼容说明。
3. 删除后可由现有能力替代，且不会破坏当前回归。

## 2. 首批候选（M4 执行结果）

### C1 `node_termproxy_shell_exec`

- 证据
  - ISSUE-006 明确 root shell 路径不在主链路验收范围内，并在 v0.1 M4 完成实现删除。
  - 现有主路径由 `provision-template-from-artifact` + SSH/QGA action 覆盖。
- 风险
  - 可能仍被少量手工排障脚本使用。
- 结论
  - 已删除（v0.1 M4）。

### C2 `storage_upload_snippet`

- 证据
  - Live 环境中 snippets 不在 upload API 可上传类型集合内；主路径已改为 `storage_upload_guard` + ISO 上传链路。
  - 目前两条 workflow 均不依赖该 action。
- 风险
  - 若后续引入 root 落盘路径，可能再次需要。
- 结论
  - 已删除（v0.1 M4）。

## 3. 审计记录

1. 已完成调用点、help、prompt 引用审计。
2. 已完成代码/文档/prompt 同步删除。
3. 替代路径：
   - `node_termproxy_shell_exec` -> 使用 Phase 4 串口动作与 SSH 控制面动作。
   - `storage_upload_snippet` -> 使用 `storage_upload_guard` + `storage_upload_iso` 主链路。

## 4. 退出标准

1. 每个候选均有明确结论：保留/弃用/删除。
2. 删除项已完成代码、help、文档、prompt 同步。
3. 回归入口执行成功，无新增断档。
