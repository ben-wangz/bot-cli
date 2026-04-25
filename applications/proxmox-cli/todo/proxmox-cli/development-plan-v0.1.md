# proxmox-cli v0.1 收敛重构计划

- 状态: active
- 优先级: high
- 创建日期: 2026-04-25
- 前置基线: `development-plan-mvp-archive.md`（MVP 已完成）

## 1. 目标

在不回退已通过回归能力的前提下，完成 v0.1 收敛重构：

1. 引入并落地 forgekit binary 版本管理。
2. 收敛测试资产，删除被 workflow 覆盖的重复 action prompt。
3. 为未被 workflow 覆盖的 action 设计“虚拟 workflow”级组合回归 prompt（仅测试编排，不新增产品 workflow）。
4. 清理 MVP 阶段不在当前需求范围内的遗留代码。
5. 全过程按 milestone 逐步推进，每个 milestone 独立验收。

## 2. 里程碑（Milestones）

### M0: 冻结基线与范围锁定

- 产出
  - 记录当前可运行基线（Phase 1-5 + 两条 e2e workflow）。
  - 输出“保留/删除/待评估”清单初稿。
- 验收
  - 基线状态可复述且可追溯到日志路径。
  - 清单覆盖 action、prompt、workflow、文档四类对象。

### M1: 版本管理迁移（forgekit binary）

- 需求来源
  - `todo/forgekit-binary-release-upgrade.md`
- 动作
  - 在仓库根 `version-control.yaml` 声明 proxmox-cli binary 版本映射。
  - 将 proxmox-cli 版本真相从兼容路径 `applications/proxmox-cli/container/VERSION` 迁移到 binary 版本文件（路径待 M1 执行时确定并固定）。
  - 更新 README/发布说明中的版本操作命令（`forgekit version get/bump`）。
  - 移除兼容模式冗余目录与配置。
- 验收
  - `forgekit version get <binary-name>` 成功返回版本。
  - `forgekit version bump <binary-name> patch` 可修改版本文件并通过校验。
  - 项目不再依赖 `container/VERSION`。

### M2: 测试去重（workflow 覆盖优先）

- 动作
  - 删除与以下 workflow 重复覆盖的 action prompt：
    - `bootstrap-bot-user-pool-acl` 覆盖：`create_pve_user_with_root`、`create_pool_with_root`、`grant_user_acl`、`get_user_acl_binding`（`revoke_user_acl` 仍保留独立校验）。
    - `provision-template-from-artifact` 覆盖：`create_vm`、`enable_serial_console`、`set_vm_agent`、`attach_cdrom_iso`、`set_net_boot_config`、`vm_power`、`serial_ws_capture_to_file`、`get_vm_status`、`convert_vm_to_template`、`get_vm_config`、`storage_upload_guard`、`storage_upload_iso`、`build_ubuntu_autoinstall_iso`。
  - 更新对应 `RUN-PHASE-*` 套件文档与 README 覆盖说明。
- 验收
  - Prompt 总数下降且无重复覆盖。
  - README 回归入口保持可执行。

### M3: 虚拟 workflow 组合回归

- 动作
  - 用组合 prompt 替代离散 action prompt（不新增产品 workflow 代码）。
  - 首批组合建议：
    - VM lifecycle chain（clone/create/config/power/migrate/template）。
    - QGA & cloud-init chain（seed/upload/agent exec/network）。
    - SSH control-plane chain（key inject/exec/scp/tunnel）。
  - 明确每个组合 prompt 的输入解析、资源隔离、清理策略。
- 验收
  - 覆盖未被 e2e workflow 触达的 action。
  - 每条组合 prompt 可独立重跑。

### M4: 代码清理（需求外遗留）

- 动作
  - 对“非需求内代码”做证据化审计（调用点、文档引用、测试覆盖）。
  - 按“先弃用标记 -> 再删除实现”的顺序推进。
  - 优先候选：历史 root shell 路径、未被当前回归使用的孤立能力。
- 验收
  - 删除后无编译回退。
  - CLI help/文档/测试同步收敛。

### M5: 文档与发布闭环（v0.1）

- 动作
  - 更新模块 README、prompt README、发布步骤、回归记录模板。
  - 形成 v0.1 发布检查单（版本、回归、变更说明）。
- 验收
  - 文档与代码一致。
  - 发布检查单可按步骤直接执行。

## 3. 执行顺序与控制原则

1. 每次只推进一个 milestone，未验收不进入下一个。
2. 每个 milestone 结束后更新状态与变更清单。
3. 重构期间保持“可随时回归”能力：任何删除都需先有替代测试路径。
4. 禁止一次性大删改；必须小步、可验证、可回滚。

## 4. 当前状态

- M0: completed
- M1: completed
- M2: completed
- M3: completed
- M4: pending
- M5: pending
