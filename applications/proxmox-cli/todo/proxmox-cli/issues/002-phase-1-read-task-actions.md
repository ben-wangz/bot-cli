# ISSUE-002 Phase 1 Read and Task Actions

- status: open
- priority: high
- phase: 1
- depends_on: ISSUE-001

## Goal

实现只读与任务查询动作，建立异步任务观测能力。

## Actions

- A01 list_nodes
- A02 list_cluster_resources
- A03 list_vms_by_node
- A04 get_vm_config
- A05 get_effective_permissions
- A06 get_task_status
- A11 get_next_vmid
- A12 get_vm_status
- A21 list_tasks_by_vmid

## Tasks

- [ ] 完成 9 个 action 的参数与执行实现。
- [ ] 统一 node/vmid/upid 参数校验。
- [ ] 补齐 action help 参数说明与示例。
- [ ] 为 9 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 9 个 action 命令均可执行并返回规范结果。
- [ ] task 类 action 可输出关键诊断字段。
- [ ] 9 条 prompt 通过。
