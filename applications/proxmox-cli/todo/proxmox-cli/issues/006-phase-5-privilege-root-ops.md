# ISSUE-006 Phase 5 Privilege Ladder and Root Ops

- status: in_progress
- priority: high
- phase: 5
- depends_on: ISSUE-005

## Goal

实现 root 专属能力路径与权限阶梯验证。

## Actions

- A36 enforce_args_permission_guard
- A37 extract_kernel_initrd_workaround
- A38 set_args_with_root_session
- A39 node_termproxy_shell_exec
- A41 revalidate_privilege_ladder
- A42 autoinstall_trigger_retest

## Tasks

- [ ] 实现 `args` 权限错误识别与引导。
- [x] 实现 root session ticket 通道（root-token 登录与会话建立）。
- [x] 实现 node termproxy shell 执行器（会话握手、交互输入与回显已打通）。
- [ ] 实现 kernel/initrd 提取与校验。
- [ ] 实现权限梯度复验报告输出。
- [ ] 为 6 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 6 个 action 可体现 user/root 差异。
- [ ] A38/A39 在 root 路径可成功执行。
- [ ] A41 产出可审计结论。
- [ ] 6 条 prompt 通过。

## Validation Note

- A39 已验证 root-token/root 会话均可创建节点 termproxy 会话并通过 websocket 认证。
- 节点会话交互输入已打通（`echo ...` 可见回显），termproxy 输入协议采用 data/resize 帧并兼容 text+binary 发送。
- API 文档显示 `cmd-opts` 为“null terminated strings”，后续需补充编码策略与输入协议适配。
- 根据上游实现（`PVE/API2/Nodes.pm`）验证：仅 `root@pam` session（非 token 标识）可执行 `upgrade/ceph_install` 命令；`root-token` 实际受限于 `login` 路径。
- 因此 A39 若要进入 root 命令执行闭环，需补充 root password 会话凭据，或先实现完整 VNC/RFB 键盘事件注入以完成交互登录。
