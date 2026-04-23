# ISSUE-004 Phase 3 Cloud-init and QGA Actions

- status: completed
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

- [x] 实现 qga 网络查询与 IPv4 过滤。
- [x] 实现 qga exec + status 轮询。
- [x] 实现 cloudinit dump 三种类型。
- [x] 实现 storage upload guard 并处理 snippets 限制提示。
- [x] 实现 seed 渲染与本地托管。
- [x] 为 6 个 action 新增独立正向 prompt（A17 拆分 exec/status）。
- [x] 补充 QGA 安装路径依赖说明：当无 qga-ready 基础镜像时，需提前实现 Phase 4 串口路径用于来宾内安装 qemu-guest-agent。

## Dependency Note

- A22 当前定位为 guard（能力探测与限制提示），不承担上传职责。
- 已新增独立动作 `storage_upload_snippet`（不耦合 A22）；但 live 验证显示 snippets 不在 upload API 支持枚举，仍需 root 落盘路径。
- QGA 正向回归在无 qga-ready 模板时会被环境阻塞，需依赖 Phase 4 的控制台能力（至少 A29/A34 子集）完成来宾内安装与启用。
- Live 验证补充：`/nodes/{node}/storage/{storage}/upload` 当前返回可上传类型 `iso|vztmpl|import`，`snippets` 不能经该 API 上传；snippet 落盘需依赖后续 root 路径（Phase 5）或预置模板。
- 已通过 A39 root 会话完成节点 snippet 写入与读取校验，但当前模板侧未形成稳定 cloud-init→qga 生效链路，QGA 正向仍受环境阻塞。
- 方案记录：当前“单盘自包含 ISO”可用于稳定诊断，但产物体积过大；后续回到“小 shim + HTTP 安装源”方案以满足原始目标。

## Acceptance

- [x] 6 个 action 均通过正向主流程。
- [x] A22 对受限类型可正确提示。
- [x] A27 输出 seed 路径与可访问地址。
- [x] 6 条 prompt 通过（A17 拆分 exec/status）。
