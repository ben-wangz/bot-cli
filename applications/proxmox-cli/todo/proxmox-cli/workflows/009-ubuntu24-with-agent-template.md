# WORKFLOW-009 ubuntu24-with-agent-template

- status: in_progress
- priority: high
- owner: ben.wangz

## Goal

从零创建一个 Ubuntu24 VM 模板（含 `qemu-guest-agent`），并把模板 VMID 写入固定文件：

- `build/ubuntu-24-with-agent.vm-template.id`

该文件作为后续 action prompt 的统一输入，确保每个测试都可从同一基线模板克隆独立 VM。

## Scope

- command: `proxmox-cli workflow ubuntu24-with-agent-template`
- module: `applications/proxmox-cli/src/internal/cli`
- output artifact: `build/ubuntu-24-with-agent.vm-template.id`

## Inputs

- required:
  - `--node <node>`
  - `--target-vmid <id>`（目标模板 VMID，必须在允许范围内）

## Constraints

- 必须遵守 VMID 安全范围（`PVE_ALLOWED_VMID_MIN..PVE_ALLOWED_VMID_MAX`，默认 `1001..2000`）。
- workflow 输入参数仅允许 `--node` 和 `--target-vmid`。
- workflow 内所有写操作必须等待任务完成（UPID wait）。
- `target-vmid` 已存在且已是模板时，应走幂等复用路径并返回 `reused_template=true`。
- `target-vmid` 已存在但不是模板时，workflow 必须失败并给出可读错误。
- 构建 ISO 依赖本地文件与系统命令（`mkisofs`、loop mount），执行环境需满足 `build_ubuntu_autoinstall_iso` 前置条件。

## Workflow Steps

1. 优先复用预构建 installer ISO（`build/ubuntu-24.04.2-shim-nocloud-serial-poweroff.iso`）；不存在时再构建 autoinstall ISO（嵌入 NoCloud seed，含 qga 包安装与启用）。
2. 上传 ISO 到节点存储（默认 `local`）。
3. 创建目标 VM（默认 2C/4G/32G，`net0=virtio,bridge=vmbr0`）。
4. 启用串口控制台与 VM agent 配置。
5. 挂载安装 ISO，设置安装启动顺序（`ide2` 优先）。
6. 启动安装并执行 `serial_ws_capture_to_file`，等待 `poweroff` 信号。
7. 安装完成后卸载 ISO，切换磁盘启动（`scsi0`）。
8. 启动来宾并轮询 QGA readiness。
9. 停机并转换为模板。
10. 将模板 VMID 写入模板 ID 文件。

## Success Signals

- 命令返回 `ok=true`。
- `result.template_vmid` 为目标模板 VMID（等于输入 `target-vmid`）。
- `result.template_id_path` 文件存在且内容等于 `target-vmid`。
- `result.steps` 包含完整步骤记录。
- 若目标模板已存在并被复用，返回 `result.reused_template=true`。

## Failure Diagnostics

- 参数错误：缺少必选参数、VMID 越界、出现不允许参数。
- 资源错误：`target-vmid` 已存在且不是模板（不可安全复用）。
- API 错误：构建/上传 installer、安装监控、QGA 探测、模板化任一步骤失败。
- 文件错误：写 template id 文件失败。

## Follow-up Integration

后续所有需要“可变更 VM 状态”的测试 prompt，应统一执行：

1. 读取 `build/ubuntu-24-with-agent.vm-template.id`。
2. 分配 fresh `TEST_VMID`。
3. 从模板克隆 `TEST_VMID`。
4. 测试结束后（成功或失败）执行自检与自销毁。
