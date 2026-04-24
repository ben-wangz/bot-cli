# ISSUE-012 provision-template-from-artifact

- status: draft
- priority: high
- owner: ben.wangz

## Goal

消费“已制备、已上传”的 installer artifact，创建并验证模板 VM。流程强调可复用、可恢复，并优先保证失败可诊断。

## Prerequisite Actions

在执行本 workflow 前，artifact 需通过以下 action 预先准备：

1. `build_ubuntu_autoinstall_iso`（本地构建 installer ISO）
2. `storage_upload_iso`（上传 ISO 到目标 node/storage，得到 `<storage>:iso/<file>.iso`）

建议在上传前先执行：`storage_upload_guard`（用于存储能力探测与限制提示）。

## Scope

- command (planned): `proxmox-cli workflow provision-template-from-artifact`
- module: `applications/proxmox-cli/src/internal/cli`
- output: workflow standard JSON (`workflow/ok/scope/request/result/diagnostics`)

## Inputs

- required:
  - `--node <node>`
  - `--target-vmid <id>`
  - `--artifact-iso <storage:iso/file.iso>`
  - `--install-timeout-seconds <n>`
- optional:
  - `--template-name <name>`
  - `--memory <mb>`
  - `--cores <n>`
  - `--disk-size-gb <n>`
  - `--net0 <value>`
  - `--install-boot-order <order>`
  - `--runtime-boot-order <order>`
  - `--bootdisk <diskid>` (default `scsi0`)
  - `--resume-from <none|serial_wait>` (default `none`)

## Mandatory Behavior Rules

1. `list_cluster_resources --type vm` 若发现 `target-vmid` 已存在，直接报错。
2. `artifact-iso` 必须是已上传到 PVE 的 volid；workflow 内不执行上传，不存在则直接报错。
3. `attach_cdrom_iso` 为必选步骤，且当前仅支持 ISO 安装路径。
4. `set_net_boot_config` 必须使用输入的 boot hints（示例见下文）。
5. `serial_ws_capture_to_file` 仅使用 `timeout`，不配置 `expect`。
6. 若串口等待超时，且 `get_vm_status` 不是预期 `stopped`，必须直接失败，并提示检查 serial 日志与 VM 状态。
7. 支持 `--resume-from serial_wait`：从 `serial_ws_capture_to_file` 步骤继续等待（标记“未充分测试，慎用”）。
8. `attach_cdrom_iso --iso none,media=cdrom` 为必选（安装后必须卸载介质）。
9. 不包含 SSH 探测步骤（简化优先）。

## Action Steps

1. `list_nodes`
   - 校验目标 node 可用。
2. `list_cluster_resources --type vm`
   - 若 `target-vmid` 已存在，直接报错退出。
3. `create_vm`
   - 创建目标 VM。
4. `enable_serial_console`
   - 开启串口会话。
5. `set_vm_agent --enabled 1`
   - 预置 agent 标志（不作为安装完成判定）。
6. `attach_cdrom_iso --iso <artifact-iso> --slot ide2`
   - 必选。若 ISO volid 不存在，此步报错并终止。
7. `set_net_boot_config`
   - 使用 `install_boot_order` 进入安装阶段。
8. `vm_power --mode start --desired-state running`
   - 启动安装。
9. `serial_ws_capture_to_file --timeout-seconds <install-timeout-seconds>`
   - 仅按超时等待并记录日志，不配 `expect`。
10. `get_vm_status`
    - 读取当前状态。
    - 若第 9 步超时且此处状态不是 `stopped`，直接失败，错误信息必须包含：
      - `serial_log_path`
      - `vm_status`
      - `next_action_hint`（检查日志与 VM 当前状态）
11. `attach_cdrom_iso --iso none,media=cdrom --slot ide2`
    - 必选。卸载安装介质。
12. `set_net_boot_config`
    - 切换为 `runtime_boot_order`。
13. `vm_power --mode stop --desired-state stopped`
    - 若 VM 仍在运行，先确保停机。
14. `convert_vm_to_template`
    - 转换为模板。
15. `get_vm_config`
    - 校验模板状态与关键配置。

## Resume-from Design

- `--resume-from none`：执行完整流程。
- `--resume-from serial_wait`：跳过 1-8 步，直接从第 9 步继续等待安装完成。
- 风险提示：`serial_wait` 仍缺少充分回归，默认不建议生产常态使用。

## Boot Hints Example

```text
--install-boot-order "order=ide2"
--runtime-boot-order "order=scsi0"
--bootdisk scsi0
```

含义：安装阶段仅从挂载 ISO 启动；安装后切换为系统盘启动。

## Long Wait Observability (Serial)

当安装过程在 `subiquity/Install/install/configure_apt/cmd-in-target` 等步骤停留较久时，默认串口日志可见性不足。

建议在 `applications/proxmox-cli/assets/ubuntu-24.04/user-data` 增加 `early-commands`，将 installer 内部日志实时转发到 `ttyS0`，便于被 `serial_ws_capture_to_file` 采集：

```yaml
autoinstall:
  early-commands:
    - sh -c 'for f in /var/log/installer/subiquity-server-debug.log /var/log/installer/curtin-install.log /var/log/cloud-init.log; do [ -f "$f" ] && tail -F "$f" >/dev/ttyS0 2>&1 & done'
```

运行策略建议：

1. 首次执行 `--resume-from none` 可使用较短超时（如 `600s`）。
2. 超时后检查 `serial_log_path` 定位当前安装阶段。
3. 继续执行 `--resume-from serial_wait --install-timeout-seconds 600`，分段观察进度，避免一次长等待不可见。

## Output Contract

- `result.template_vmid`
- `result.template_name`
- `result.serial_log_path`
- `result.steps[]`
- `diagnostics.step_count`
- `diagnostics.resumed_from`
