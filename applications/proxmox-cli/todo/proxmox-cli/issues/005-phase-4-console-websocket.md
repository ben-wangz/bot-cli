# ISSUE-005 Phase 4 SSH Control Plane (QGA Bootstrap)

- status: completed
- priority: high
- phase: 4
- depends_on: ISSUE-004

## Dependency Note

- 为解除 ISSUE-004 的 QGA 环境阻塞，需提前实现 Phase 4 最小子集：
  - A29 `open_vm_termproxy`
  - A34 `serial_ws_session_control`（脚本模式最小能力）
- 该最小子集用于在来宾内执行 `qemu-guest-agent` 安装与启用流程，不改变 ISSUE-005 的完整验收范围。

## Scope Update

- ISSUE-005 当前范围收敛为 SSH 控制面交付：通过 QGA 完成 SSH 通路引导，后续稳定交互使用 SSH/SCP/Tunnel。
- A18 `start_vnc_proxy` 与 A19 `connect_vnc_websocket` 从 ISSUE-005 范围中移除。

## Current Validation Note

- QGA 主链路（A15/A17）已可用于来宾内执行命令。
- 下一步聚焦将“可执行命令”升级为“稳定 SSH 控制面能力”。
- 2026-04-23 实测：`build/phase4-suite/summary.json` 显示 A44-A51 全部通过（8/8）。

## Goal

交付 SSH 控制面能力：QGA 负责引导 SSH 通路，SSH 负责稳定命令执行与文件传输。

## Actions

- A44 ssh_check_service
- A45 ssh_inject_pubkey_qga
- A46 ssh_exec
- A47 ssh_scp_transfer
- A48 ssh_print_connect_command
- A49 ssh_tunnel_start
- A50 ssh_tunnel_status
- A51 ssh_tunnel_stop

## Action Contracts

- A44 `ssh_check_service`
  - 输入：`host` `port` `user` `identity-file`(optional) `connect-timeout-seconds`(default 5)
  - 行为：使用本机 ssh client 执行 batch 探活，不进入交互模式。
  - 输出：`reachable` `auth_ok` `latency_ms` `stderr_tail`。
- A45 `ssh_inject_pubkey_qga`
  - 输入：`node` `vmid` `username`，以及 `pub-key-file` 或 `pub-key`（二选一）。
  - 行为：通过 QGA 创建/修正 `~/.ssh/authorized_keys` 与权限（700/600）。
  - 输出：`username` `authorized_keys_path` `fingerprint`。
- A46 `ssh_exec`
  - 输入：`host` `port` `user` `identity-file` `command` `timeout-seconds`。
  - 行为：通过 ssh client 执行远程命令，支持参数透传与超时控制。
  - 输出：`exit_code` `stdout` `stderr` `duration_ms`。
- A47 `ssh_scp_transfer`
  - 输入：`direction(upload|download)` `host` `port` `user` `identity-file` `local-path` `remote-path` `recursive`。
  - 行为：通过 scp 双向传输文件或目录。
  - 输出：`direction` `bytes` `duration_ms` `verified_exists`。
- A48 `ssh_print_connect_command`
  - 输入：`host` `port` `user` `identity-file` `extra-args`。
  - 行为：仅生成可复制执行的 ssh 命令，供用户手动进入交互。
  - 输出：`command`。
- A49 `ssh_tunnel_start`
  - 输入：`host` `port` `user` `identity-file` `local-port` `remote-host` `remote-port` `pid-file` `log-file`。
  - 行为：后台创建 SSH 隧道并写入 pid/log 文件。
  - 输出：`pid` `pid_file` `log_file` `local_endpoint`。
- A50 `ssh_tunnel_status`
  - 输入：`pid-file`。
  - 行为：读取 pid 并检查进程与本地监听状态。
  - 输出：`running` `pid` `pid_file` `local_endpoint` `last_error`。
- A51 `ssh_tunnel_stop`
  - 输入：`pid-file`。
  - 行为：按 pid 停止 tunnel 并清理状态文件（尽量幂等）。
  - 输出：`stopped` `pid` `pid_file` `cleanup`。

## Tasks

- [x] 实现 A44：使用本机 ssh client 检查目标 SSH 可达与认证结果。
- [x] 实现 A45：通过 QGA 注入登录公钥，支持 `--pub-key-file` 与 `--pub-key`。
- [x] 实现 A46：通过 ssh client 执行命令（支持 timeout/identity/port 参数）。
- [x] 实现 A47：通过 scp 执行双向文件传输（upload/download）。
- [x] 实现 A48：生成可复制执行的 ssh 命令，供用户手动进入交互会话。
- [x] 实现 A49：创建 SSH tunnel 并写入 pid/log 文件用于监控与回溯。
- [x] 实现 A50：读取 pid 文件并检查 tunnel 进程/端口存活状态。
- [x] 实现 A51：基于 pid 文件停止 tunnel 并输出清理结果。
- [x] 为 A44-A51 新增独立正向 prompt。

## Suggested Validation Flow

- `A15 agent_network_get_interfaces` 获取 guest IP。
- `A45 ssh_inject_pubkey_qga` 注入登录公钥。
- `A44 ssh_check_service` 验证 SSH 服务可达。
- `A46 ssh_exec` 执行 `hostname` 作为命令链路验证。
- `A47 ssh_scp_transfer` 执行 upload/download 双向验证。
- `A48 ssh_print_connect_command` 输出用户手动接管命令。
- `A49 -> A50 -> A51` 验证 tunnel 生命周期闭环。

## Prompt Execution Classes

- Independent-VM：A45、A47（会修改 guest 状态，需每条 prompt 自建自清理）。
- Shared-VM：A44、A46、A49（低副作用或同链路验证，复用 suite 级 shared VM）。
- No-VM：A48、A50、A51（不创建/销毁 VM；A50/A51 依赖 A49 产出的 pid 文件）。

## Acceptance

- [x] A44-A51 可执行。
- [x] QGA 注入公钥后，SSH 命令执行链路可闭环。
- [x] SCP 双向传输可闭环并可验证文件一致性。
- [x] 可输出用户手动接管用的 SSH 连接命令。
- [x] Tunnel 生命周期（start/status/stop）可闭环，含 pid 文件回溯。
- [x] A44-A51 prompt 全部通过。
