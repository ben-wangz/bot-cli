# proxmox-cli 重构 M2：按能力域搬迁

- 状态: completed
- 日期: 2026-04-26
- 依赖: `refactor-m1-target-structure.md`

## 本轮迁移（第一批 + 第二批 + 第三批 + 第四批）

目标：先搬迁低风险域，保持行为不变。

已完成：

1. `inventory/task` 相关 action 从 `phase1.go` 拆到 `domain_inventory.go`。
2. `access` 相关 action 从 `phase5_acl.go` 拆到 `domain_access.go`。
3. `vm` 相关 action 从 `phase2.go` 拆到 `domain_vm.go`。
4. `guest` 相关 action 从 `phase3.go` 拆到 `domain_guest.go`。
5. `storage` 相关 action 从 `phase3.go` 拆到 `domain_storage.go`。
6. `console` 相关 action 从 `phase4.go` 拆到 `domain_console.go`。
7. `ssh` 相关 action 从 `phase4_ssh.go` 拆到 `domain_ssh.go`。
8. registry 分发与 action 名称不变，输出契约保持不变。
9. phase 兼容层历史辅助函数已抽离到稳定 helper 文件（`result_helpers.go`、`action_helpers_common.go`、`guest_storage_helpers.go`、`console_helpers.go`、`ssh_helpers.go`、`access_helpers.go`）。

## 变更说明

1. 当前仅做代码位置迁移，不改 API 语义、不改参数。
2. `phase1.go`、`phase2.go`、`phase3.go`、`phase4.go`、`phase4_ssh.go` 与 `phase5_acl.go` 仍保留必要桥接/辅助逻辑，作为过渡层。
3. 编译验证已通过（`go build ./...`）。

## 下一批建议（M2 收尾）

1. 进入 M3：workflow 编排层收敛。
2. 在 M3 中移除剩余 phase dispatcher 兼容层（若确认无历史入口依赖）。
