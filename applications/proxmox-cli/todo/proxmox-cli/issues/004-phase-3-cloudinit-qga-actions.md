# ISSUE-004 Phase 3 Cloud-init and QGA Actions

- status: open
- priority: high
- phase: 3
- depends_on: ISSUE-003

## Goal

实现 cloud-init 与 qemu guest agent 相关动作，为安装后自动化提供能力。

## Actions

- A15 agent_network_get_interfaces
- A17 agent_exec / agent_exec_status
- A20 dump_cloudinit
- A22 storage_upload_guard
- A27 render_and_serve_seed

## Tasks

- [ ] 实现 qga 网络查询与 IPv4 过滤。
- [ ] 实现 qga exec + status 轮询。
- [ ] 实现 cloudinit dump 三种类型。
- [ ] 实现 storage upload guard 并处理 snippets 限制提示。
- [ ] 实现 seed 渲染与本地托管。
- [ ] 为 5 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 5 个 action 均通过正向主流程。
- [ ] A22 对受限类型可正确提示。
- [ ] A27 输出 seed 路径与可访问地址。
- [ ] 5 条 prompt 通过。
