# ISSUE-005 Phase 4 Console and WebSocket

- status: in_progress
- priority: high
- phase: 4
- depends_on: ISSUE-004

## Dependency Note

- 为解除 ISSUE-004 的 QGA 环境阻塞，需提前实现 Phase 4 最小子集：
  - A29 `open_vm_termproxy`
  - A34 `serial_ws_session_control`（脚本模式最小能力）
- 该最小子集用于在来宾内执行 `qemu-guest-agent` 安装与启用流程，不改变 ISSUE-005 的完整验收范围。

## Current Validation Note

- A18/A19/A29/A32/A34/A40 代码路径已接入 CLI 并完成本地编译验证。
- Live PVE 验证中，websocket 握手已打通（`open_vm_termproxy` + `vncwebsocket` 可建立会话）。
- 当前验证结果：`serial_ws_session_control` 协议层已打通，支持脚本输入与回显匹配。
- `start_vnc_proxy` 与 `connect_vnc_websocket` 已补齐，支持 vncproxy 启动与 websocket 探测连通性。
- `validate_k1_serial_readable` 与 `validate_serial_output_criterion2` 已补齐：前者验证串口可读，后者对“仅 termproxy 启动横幅”场景给出失败诊断。
- 当前阻塞转为环境侧：现有模板在 VM 串口仍未稳定出现可交互 login 提示，暂无法通过串口完成来宾内安装 QGA。
- 同期验证发现 snippet 无法经 upload API 直接写入（`snippets` 不在可上传枚举），如需 cloud-init 自举仍需 Phase 5 root 文件落盘路径。
- 已转入 Phase 5 并实现 `node_termproxy_shell_exec` 初版，root-token 会话可建立；下一步需完善命令执行回显能力后回灌 ISSUE-004。

## Goal

交付串口控制面能力：脚本模式为主，交互模式基础版。

## Actions

- A18 start_vnc_proxy
- A19 connect_vnc_websocket
- A29 open_vm_termproxy
- A32 validate_k1_serial_readable
- A34 serial_ws_session_control
- A40 validate_serial_output_criterion2

## Tasks

- [x] 提前落地最小子集：A29 `open_vm_termproxy` + A34 `serial_ws_session_control`（用于 QGA 安装前置）。
- [x] 实现 termproxy 握手与 ticket 提取。
- [x] 实现 websocket 认证行发送。
- [x] 实现 xterm 帧协议编解码与 keepalive。
- [x] 实现脚本模式（expect/timeout/summary）。
- [ ] 实现交互模式基础版（stdin/stdout 透传）。
- [x] 为 6 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 6 个 action 可执行。
- [ ] 脚本模式可用于自动化闭环步骤。
- [ ] 交互模式可人工接管排障。
- [ ] 6 条 prompt 通过。
