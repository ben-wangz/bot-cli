# ISSUE-005 Phase 4 Console and WebSocket

- status: open
- priority: high
- phase: 4
- depends_on: ISSUE-004

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

- [ ] 实现 termproxy 握手与 ticket 提取。
- [ ] 实现 websocket 认证行发送。
- [ ] 实现 xterm 帧协议编解码与 keepalive。
- [ ] 实现脚本模式（expect/timeout/summary）。
- [ ] 实现交互模式基础版（stdin/stdout 透传）。
- [ ] 为 6 个 action 各新增 1 条独立正向 prompt。

## Acceptance

- [ ] 6 个 action 可执行。
- [ ] 脚本模式可用于自动化闭环步骤。
- [ ] 交互模式可人工接管排障。
- [ ] 6 条 prompt 通过。
