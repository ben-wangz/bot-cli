# ISSUE-001 Phase 0 Foundation

- status: completed
- priority: high
- phase: 0

## Goal

完成 Go 工程底座，使 `proxmox-cli` 具备可运行入口、配置加载、输出与日志脱敏能力。

## Scope

- CLI 顶层命令骨架：`action` `workflow` `console` `auth`
- 全局参数解析与校验
- 认证配置加载（flag > env > auth-file）
- 输出格式（json/yaml/table）
- 日志脱敏与错误模型

## Tasks

- [x] 初始化 Go module 与标准目录布局。
- [x] 实现顶层命令和 help 文本。
- [x] 实现全局参数定义与默认值。
- [x] 实现 Credential Loader 与 auth-scope 校验。
- [x] 实现基础 HTTP client（timeout/tls/header）。
- [x] 实现统一输出和统一错误码。
- [x] 实现 secret redaction（password/token/cookie/ticket）。
- [x] 建立 `applications/proxmox-cli/src/test/prompts/` 目录结构。

## Acceptance

- [x] `proxmox-cli --help` 可正常运行。
- [x] 配置优先级行为符合约定。
- [x] debug 输出不包含明文 secret。
