# ISSUE-006 Phase 5 Root-Assisted User Bootstrap

- status: completed
- priority: high
- phase: 5
- depends_on: ISSUE-005

## Goal

项目主路径不依赖 root 完成 VM 管理；root 仅用于一次性辅助创建/赋权最小权限 PVE user，后续日常执行全部走 user 路径。

## Actions

- P5-00 create_pool_with_root
- P5-01 create_pve_user_with_root
- P5-02 get_user_acl_binding
- P5-03 grant_user_acl
- P5-04 revoke_user_acl

## Scope Change

- 废弃（cancelled）：旧方案中的 A36/A37/A38。
- 废弃（cancelled）：`node_termproxy_shell_exec` 作为常规能力路径；不再作为 VM 管理主链路依赖。
- 保留唯一 root 目标：用 root 凭据一次性完成 user 创建、角色绑定、ACL 授权，随后回到 user 账号执行。
- 权限能力扩展为 pool+user+ACL 管理：创建 pool、创建用户、查询授权、授权新增、授权撤销（变更通过 delete+add 组合实现）。

## Tasks

- [x] 实现 P5-01 `create_pve_user_with_root`（创建 user，支持幂等重跑）。
- [x] 实现 P5-00 `create_pool_with_root`（创建 pool，支持幂等重跑）。
- [x] 实现 P5-02 `get_user_acl_binding`（读取 user 现有 ACL/角色绑定）。
- [x] 实现 P5-03 `grant_user_acl`（新增授权）。
- [x] 实现 P5-04 `revoke_user_acl`（撤销授权，幂等）。
- [x] 文档化授权变更策略：不提供 update action，统一使用 revoke+grant 组合。
- [x] 输出最小权限授予清单（路径、角色、动词）与审计结果。
- [x] 为 5 个 action 各新增 1 条独立正向 prompt。
- [x] 文档化切换策略：bootstrap 完成后，所有 workflow/action 统一使用 user 凭据。
- [x] 设计并文档化 pool+user 初始化 workflow（创建 user/pool、pool 级 PVEAdmin、A01/A02 与 A22/A27 最小补充权限）。

## Acceptance

- [x] P5-00..P5-04 可执行并形成完整 pool+user+ACL 管理闭环。
- [x] 可输出可审计结论（授予对象、角色、路径、校验状态）。
- [x] 5 条 prompt 通过。

## Validation Note

- 重新审查结论：VM 管理主路径不再需要 root 会话 shell 能力；root 仅用于用户授权 bootstrap。
- `node_termproxy_shell_exec` 已在 v0.1 M4 删除，不纳入后续验收目标。
- 2026-04-23 root 实测（`build/phase5-suite/`）：
  - P5-01：`botcli-phase5@pve` 创建/复用成功（幂等）。
  - P5-03：对 `/vms` 授予 `PVEAuditor` 成功。
  - P5-02：授权后可读到 binding，撤销后 count 返回 0。
  - P5-04：授权撤销成功（幂等可重复）。
- setup 已切换为以 bootstrap 创建的 user 凭据（`build/pve-user.env`）执行 Phase 1-4 与模板 workflow，root 仅用于一次性 bootstrap。
- prompt 执行汇总已更新为 5 条目标（P5-00..P5-04），当前按 5/5 通过闭环。

## Bootstrap Workflow Design (New)

- 已新增 issue 设计文档：`todo/proxmox-cli/issues/010-bootstrap-bot-user-pool-acl.md`。
- 设计目标：一次执行完成测试账号与最小补充权限初始化；workflow 仅返回标准 JSON 输出，不直接写 env 文件。
- 设计步骤（action-first）：
  1. `create_pve_user_with_root`
  2. `create_pool_with_root`
  3. `grant_user_acl` -> `/pool/<poolid>` + `PVEAdmin`
  4. `grant_user_acl` -> `/` + `PVEAuditor`（A01/A02）
  5. `grant_user_acl` -> `/storage` + `PVEDatastoreAdmin`（支持 A22 与 ISO 上传；A27 为本地渲染）

## Least-Privilege Profile (Initial)

- subject: `botcli-phase5@pve`
- binding_path: `/vms`
- role: `PVEAuditor`
- propagate: `1`
- change_policy: `revoke + grant`（不提供 update action）
