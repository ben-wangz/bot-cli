# ISSUE-006 Phase 5 Root-Assisted User Bootstrap

- status: in_progress
- priority: high
- phase: 5
- depends_on: ISSUE-005

## Goal

项目主路径不依赖 root 完成 VM 管理；root 仅用于一次性辅助创建/赋权最小权限 PVE user，后续日常执行全部走 user 路径。

## Actions

- A39 bootstrap_user_acl_with_root

## Scope Change

- 废弃（cancelled）：A36/A37/A38/A41/A42。
- 废弃（cancelled）：`node_termproxy_shell_exec` 作为常规能力路径；不再作为 VM 管理主链路依赖。
- 保留唯一 root 目标：用 root 凭据一次性完成 user 创建、角色绑定、ACL 授权，随后回到 user 账号执行。

## Tasks

- [ ] 新增并实现 A39 `bootstrap_user_acl_with_root`（创建 user/角色/ACL，支持幂等重跑）。
- [ ] 输出最小权限授予清单（路径、角色、动词）与审计结果。
- [ ] 新增 A39 独立正向 prompt（root 凭据仅用于该步骤）。
- [ ] 文档化切换策略：bootstrap 完成后，所有 workflow/action 统一使用 user 凭据。

## Acceptance

- [ ] A39 可在 root 凭据下完成 user 授权 bootstrap，并支持幂等重跑。
- [ ] bootstrap 后，Phase 1-4 与 workflow 在 user 凭据下可执行，不依赖 root action。
- [ ] A39 产出可审计结论（授予对象、角色、路径、校验状态）。
- [ ] A39 prompt 通过。

## Validation Note

- 重新审查结论：VM 管理主路径不再需要 root 会话 shell 能力；root 仅用于用户授权 bootstrap。
- 既有 `node_termproxy_shell_exec` 能力作为历史实现保留，不纳入后续验收目标。
