# ISSUE-006 Phase 5 Root-Assisted User Bootstrap

- status: in_progress
- priority: high
- phase: 5
- depends_on: ISSUE-005

## Goal

项目主路径不依赖 root 完成 VM 管理；root 仅用于一次性辅助创建/赋权最小权限 PVE user，后续日常执行全部走 user 路径。

## Actions

- P5-01 create_pve_user_with_root
- P5-02 get_user_acl_binding
- P5-03 grant_user_acl
- P5-04 revoke_user_acl

## Scope Change

- 废弃（cancelled）：旧方案中的 A36/A37/A38。
- 废弃（cancelled）：`node_termproxy_shell_exec` 作为常规能力路径；不再作为 VM 管理主链路依赖。
- 保留唯一 root 目标：用 root 凭据一次性完成 user 创建、角色绑定、ACL 授权，随后回到 user 账号执行。
- 权限能力扩展为 user+ACL 管理：创建用户、查询授权、授权新增、授权撤销（变更通过 delete+add 组合实现）。

## Tasks

- [x] 实现 P5-01 `create_pve_user_with_root`（创建 user，支持幂等重跑）。
- [x] 实现 P5-02 `get_user_acl_binding`（读取 user 现有 ACL/角色绑定）。
- [x] 实现 P5-03 `grant_user_acl`（新增授权）。
- [x] 实现 P5-04 `revoke_user_acl`（撤销授权，幂等）。
- [x] 文档化授权变更策略：不提供 update action，统一使用 revoke+grant 组合。
- [x] 输出最小权限授予清单（路径、角色、动词）与审计结果。
- [x] 为 4 个 action 各新增 1 条独立正向 prompt。
- [x] 文档化切换策略：bootstrap 完成后，所有 workflow/action 统一使用 user 凭据。

## Acceptance

- [x] P5-01..P5-04 可执行并形成完整 user+ACL 管理闭环。
- [ ] bootstrap 后，Phase 1-4 与 workflow 在 user 凭据下可执行，不依赖 root action。
- [x] 可输出可审计结论（授予对象、角色、路径、校验状态）。
- [ ] 4 条 prompt 通过。

## Validation Note

- 重新审查结论：VM 管理主路径不再需要 root 会话 shell 能力；root 仅用于用户授权 bootstrap。
- 既有 `node_termproxy_shell_exec` 能力作为历史实现保留，不纳入后续验收目标。
- 2026-04-23 root 实测（`build/phase5-suite/`）：
  - P5-01：`botcli-phase5@pve` 创建/复用成功（幂等）。
  - P5-03：对 `/vms` 授予 `PVEAuditor` 成功。
  - P5-02：授权后可读到 binding，撤销后 count 返回 0。
  - P5-04：授权撤销成功（幂等可重复）。

## Least-Privilege Profile (Initial)

- subject: `botcli-phase5@pve`
- binding_path: `/vms`
- role: `PVEAuditor`
- propagate: `1`
- change_policy: `revoke + grant`（不提供 update action）
