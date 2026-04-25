# proxmox-cli v0.1 测试收敛方案

- 状态: draft
- 优先级: high
- 关联计划: `development-plan-v0.1.md`（M2/M3）

## 1. 收敛原则

1. 已被 e2e workflow 稳定覆盖的 action，不再保留重复单 action prompt。
2. 未被 e2e workflow 覆盖的 action，优先收敛到“虚拟 workflow”组合 prompt。
3. 保留少量必要单 action prompt，仅用于无法稳定组合验证的能力。

## 2. e2e workflow 覆盖清单（可删除重复 action prompt）

### 2.1 `bootstrap-bot-user-pool-acl`

- 覆盖 action
  - `create_pve_user_with_root`
  - `create_pool_with_root`
  - `grant_user_acl`
  - `get_user_acl_binding`
- 说明
  - `revoke_user_acl` 不在该 workflow 正向链路中，建议保留独立校验（或并入 ACL lifecycle 组合 prompt）。

### 2.2 `provision-template-from-artifact`

- 覆盖 action（workflow 内 + e2e 脚本前置）
  - `list_nodes`
  - `list_cluster_resources`
  - `storage_upload_guard`
  - `build_ubuntu_autoinstall_iso`
  - `storage_upload_iso`
  - `create_vm`
  - `enable_serial_console`
  - `set_vm_agent`
  - `attach_cdrom_iso`
  - `set_net_boot_config`
  - `vm_power`
  - `serial_ws_capture_to_file`
  - `get_vm_status`
  - `convert_vm_to_template`
  - `get_vm_config`

## 3. 虚拟 workflow 组合建议（不新增产品 workflow）

### 3.1 VM lifecycle chain

- 目标覆盖
  - `clone_template`
  - `update_vm_config`
  - `review_install_tasks`
  - `sendkey`
  - `migrate_vm`
- 说明
  - 独立于 e2e 模板生产流程，使用临时 VM，链路内自清理。

### 3.2 QGA and cloud-init chain

- 目标覆盖
  - `agent_network_get_interfaces`
  - `agent_exec`
  - `agent_exec_status`
  - `storage_upload_snippet`（若 v0.1 范围仍保留）

### 3.3 Serial and websocket chain

- 目标覆盖
  - `start_vnc_proxy`
  - `connect_vnc_websocket`
  - `open_vm_termproxy`
  - `validate_k1_serial_readable`
  - `serial_ws_session_control`
  - `validate_serial_output_criterion2`

### 3.4 SSH control-plane chain

- 目标覆盖
  - `ssh_check_service`
  - `ssh_inject_pubkey_qga`
  - `ssh_exec`
  - `ssh_scp_transfer`
  - `ssh_print_connect_command`
  - `ssh_tunnel_start`
  - `ssh_tunnel_status`
  - `ssh_tunnel_stop`

### 3.5 ACL lifecycle chain

- 目标覆盖
  - `revoke_user_acl`
- 说明
  - 与 bootstrap workflow 组合验证授权增删闭环。

## 4. 待删除候选（按 M2 执行）

1. 与 2.1 重复的 Phase 5 单 action prompt。
2. 与 2.2 重复的 Phase 1/2/3 单 action prompt。
3. 删除前先确认 README 顺序与 suite runner 不再引用。

## 5. 验收标准

1. Prompt 资产数量下降，且覆盖声明无缺口。
2. 回归入口仍为 README 的 setup -> suites -> e2e 顺序。
3. 任一 action 至少有一种有效覆盖路径（workflow 或虚拟 workflow）。
