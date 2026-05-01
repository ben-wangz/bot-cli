# image-gen-cli 简略实现方案（v0）

本方案参考 `applications/aria2-cli` 与 `applications/proxmox-cli` 的组织方式，保持“agent-first + capability-first + 稳定结构化输出”的一致风格。

## 1. 目录与模块组织

建议采用与现有应用一致的分层：

- `applications/image-gen-cli/`
  - `README.md`（使用说明与最小示例）
  - `VERSION`（版本号）
  - `design.md`（功能设计）
  - `implementation.md`（本文件）
- `tests/prompts/`（回归提示词用例）
- `src/`（Go 实现）

`src/internal/` 建议按职责拆分：

- `cli`：命令入口、全局参数、子命令分发、help 文本。
- `capability`：能力注册、参数解析、请求分发、能力元数据。
- `imageapi`：与 IMAGE2/gpt-image-2 协议交互的客户端封装（streaming + non-streaming）。
- `output`：统一输出渲染（v0 先固定 JSON）。
- `apperr`：统一错误码与错误返回契约。

## 2. 命令面与交互模型

与 aria2-cli/proxmox-cli 保持一致：

- 根命令：负责全局参数（鉴权、输出、请求模式、超时、续链相关参数等）。
- `capability` 子命令：承载原子能力（生成、描述、参数校验后的执行）。
- `capability describe`：返回能力说明与参数描述，供 agent 自检与自动化调用。

## 3. Capability 设计（v0）

保持“少而稳”的能力集：

- `generate_image`：文生图主能力。
  - 支持 streaming/non-streaming。
  - 支持常用生成参数（size/quality/format/compression/background）。
  - 支持 `store` 与 `previous_response_id`。
- `describe`：列出能力及参数约束。

能力结果统一返回 envelope：

- `ok`
- `request`
- `result`
- `diagnostics`

其中 `result` 至少包含：最终文件路径、response id、关键生成参数回显。

## 4. IMAGE2 对接策略

固定单实现，不做多 provider 抽象。

- 请求层面：按 `todo/API_IMAGE2_接口说明.md` 的 Responses + `image_generation` 规范对齐。
- streaming：处理事件流，预览仅用于过程展示，不参与最终落盘。
- non-streaming：直接从完整响应提取最终结果。
- 最终图来源：统一只认 `output[].result` 的最终图片字段。
- 续链策略：仅在后端支持时使用 `previous_response_id`，不支持时明确降级提示。

## 5. 参数与配置落地

遵循已确认策略：

- 不引入配置文件。
- 优先级：命令行参数 > 环境变量 > 内置默认值。
- 本地参数校验优先：枚举值、尺寸、压缩范围与联动关系先校验再发请求。

## 6. 输出与错误契约

为 agent/脚本稳定消费，v0 输出策略如下：

- 默认 JSON 输出（与 aria2-cli 保持一致）。
- 成功：返回最终文件路径与 response id，便于下一步续链。
- 失败：返回统一错误码与可读错误消息，不返回“假成功”。
- diagnostics：记录模式信息（streaming/non-streaming）、续链信息、关键执行阶段。

## 7. 回归测试组织

参考现有应用 `tests/prompts` 方式，v0 仅保留一个最小链路用例：

- 单用例：同一提示词链路中，先 streaming 生成第一张，再 non-streaming + `previous_response_id` 生成第二张并落盘。

验收重点：

- 两次调用都成功；
- 两张图可打开；
- 第二次明确走续链；
- 最终图均来自最终 `output[].result`。

## 8. 分阶段实施顺序

1. 搭建骨架：目录、CLI 分发、输出与错误契约。
2. 接入主能力：`generate_image`（先 non-streaming，再补 streaming 事件处理）。
3. 接入续链：`store/previous_response_id` 参数与返回 id 串联。
4. 补齐 `capability describe` 与 help 文本。
5. 添加最小回归提示词并完成端到端验收。

## 9. 与现有应用对齐点

- 与 `aria2-cli` 对齐：
  - capability-first 入口
  - JSON 统一输出
  - `capability describe` 元信息
- 与 `proxmox-cli` 对齐：
  - internal 分层清晰（cli/capability/output/apperr）

结论：`image-gen-cli` v0 采用“轻量但结构完整”的实现策略，不做过度抽象，优先把最小两次调用验收链路做稳。
