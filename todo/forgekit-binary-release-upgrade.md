# TODO: 先增强 forgekit 的 binary release 支持，再升级本项目版本管理

- 状态: completed
- 优先级: high
- 创建日期: 2026-04-17

## 背景

当前项目使用 forgekit 的兼容模式管理版本（通过 `container/VERSION` 路径约定），带来了冗余结构。

## 根因

`forgekit` 的 `version-control.yaml` 目前不单独支持 binary release。

补充说明：
- 从通用能力看，也应支持“独立 container”（不被 Helm chart 直接管理）。
- 但本项目当前阶段只关注 binary release 管理能力。

## 目标

1. 先在 forgekit 中支持独立 binary release 的版本声明与版本操作。
2. 再把本项目迁移到新的版本管理方案，移除兼容模式造成的冗余。

## 计划任务

1. 在 forgekit 设计并实现 `version-control.yaml` 的 binary 配置结构。
2. 为 forgekit `version get/bump` 增加 binary 目标支持（不依赖 chart/container 约定）。
3. 为新能力补齐测试与文档（含迁移说明）。
4. 本项目切换到 binary release 版本管理新方案。
5. 清理兼容模式遗留目录与配置，保持最小必要结构。

## 验收标准

- 可通过 forgekit 直接对本项目的 binary 模块执行 `version get/bump`。
- 本项目不再依赖 `container/VERSION` 兼容路径。
- 迁移后版本管理路径清晰、无冗余目录约束。

## 完成记录（2026-04-25）

1. 仓库根 `version-control.yaml` 已声明 `proxmox-cli` binary 映射。
2. 版本文件已迁移到 `applications/proxmox-cli/VERSION`。
3. 兼容目录 `applications/proxmox-cli/container/` 已移除。
4. 已验证：`go run ./cmd/forgekit --project-root /root/code/github/bot-cli version get proxmox-cli` 返回 `0.1.0`。
