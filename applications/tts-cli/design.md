# tts-cli 设计方案（面向 bot-cli Agent，MiMo-V2.5-TTS，Go 直连 API）

## 1. 背景与目标

`tts-cli` 是 `bot-cli` 体系下的一个 Agent-facing 应用能力组件：

- 对 AI agent 提供稳定、可发现、可组合的 TTS capability 接口。
- 通过 Go 实现的本地执行入口承载调用与结果标准化。
- 底层对接 MiMo-V2.5-TTS 系列模型完成文本转语音（TTS）。

目标：

- 仅使用 Go 实现，不依赖 Python 运行时。
- 直接调用 HTTP API（OpenAI 兼容 `/v1/chat/completions`）。
- 支持三类能力：内置音色、音色设计、音色克隆。
- 支持非流式与流式（兼容模式）两种调用方式。
- 输出可直接落地为音频文件，便于 agent 编排、脚本集成与自动化。
- 对齐 bot-cli 既有应用（如 proxmox-cli / image-gen-cli）的 capability/describe/help 交互契约。

非目标（首版不做）：

- 图形界面。
- 多提供商抽象层（首版只做 MiMo）。
- 复杂音频后处理（降噪、混响、均衡等）。

## 2. 需求范围

### 2.1 功能需求

1. 支持模型：
   - `mimo-v2.5-tts`：内置音色合成模型；适合常规 TTS 生产，支持直接使用平台预置音色（如 `Chloe`、`Mia` 等）。
   - `mimo-v2.5-tts-voicedesign`：文本驱动音色设计模型；通过 `user` 描述生成目标声线，适合需要定制人物音色但无参考音频的场景。
   - `mimo-v2.5-tts-voiceclone`：参考音频克隆模型；通过 Data URI 传入样本音频复刻说话人音色，适合角色复用、品牌声线统一等场景。

2. 请求模式：
   - 非流式：一次性返回音频 Base64。
   - 流式：SSE 兼容格式返回分片（文档说明为“推理完成后统一返回流式结果”）。

3. 输入参数：
   - API 地址、API Key、超时。
   - 模型、音频格式、音色。
   - `user` 指令文本（可选/按模型要求）。
   - `assistant` 合成文本（必填）。
   - 输出目录、输出文件名。

4. 输出行为：
   - 解析音频 Base64。
   - 写入目标文件（`wav/mp3/pcm16` 等）。
   - 输出结构化 JSON（成功/失败、输出路径、响应 ID、诊断信息）。

### 2.2 约束需求

- 语言：Go。
- 调用方式：直接 HTTP API。
- 不引入 Python SDK 或 Python 工具链。
- 交互与输出契约需面向 AI agent 消费，而非仅面向人工命令行操作。
- CLI 设计与现有 `image-gen-cli` / `proxmox-cli` 风格保持一致（可观测、可自动化、可被 agent 稳定解析）。

## 3. 总体架构

采用分层结构，降低耦合，便于后续扩展。

1. `cmd/tts-cli`
   - 程序入口，参数接入。

2. `internal/cli`
   - 命令路由、help 文案、全局参数解析。

3. `internal/capability`
   - 能力注册与参数校验。
   - `generate_speech` 能力实现。

4. `internal/ttsapi`
   - HTTP 请求封装。
   - 非流式 JSON 解析。
   - 流式 SSE 解析与分片收集。

5. `internal/output`
   - 标准输出格式（json）。

6. `internal/apperr`
   - 统一错误码与可重试标识。

## 4. 命令与参数设计

### 4.1 命令结构

首版沿用统一模式：

- `tts-cli capability generate_speech ...`
- `tts-cli capability describe <name>`（输出 capability 的用途、参数、示例）

后续可扩展：

- `tts-cli capability list`
- `tts-cli capability suggest_voices`

说明（对齐 proxmox-cli 交互习惯）：

- `tts-cli --help`：显示全局帮助与命令入口。
- `tts-cli capability --help`：显示 capability 总览与实现列表。
- `tts-cli capability <name> --help`：显示单 capability 的参数说明与示例。
- `tts-cli capability describe <name>`：输出结构化描述，便于 agent/脚本消费。

说明（agent 优先）：

- 命令入口虽然是 CLI，但主要消费方是 AI agent。
- 因此 `describe`、稳定字段名、参数可发现性属于一等能力，而不是附加文档能力。
- 不提供 `tts-cli help` 子命令，仅保留 `tts-cli --help` / `-h` 作为根帮助入口。
- `help` 与 `describe` 的输出内容统一使用英文。

### 4.2 全局参数

- `--api-base-url`（默认 `https://api.xiaomimimo.com/v1`）
- `--api-key`（可由环境变量注入）
- `--timeout`（秒）
- `--output`（目前只支持 `json`）
- `--output-dir`
- `--output-name`

help 文案要求：

- 对 `--clone-voice-data-uri` 必须注明“Data URI / Data URL（RFC 2397）”。
- 对 `--timeout` 说明默认值和单位（秒）。
- 对 `--output` 说明支持格式（首版 `json`）。

### 4.3 generate_speech 参数

- `--model`：默认 `mimo-v2.5-tts`
- `--assistant-text`：必填，待合成文本
- `--user-text`：可选，风格说明或上下文
- `--audio-format`：默认 `wav`，流式建议 `pcm16`
- `--builtin-voice`：内置音色 ID（用于 `mimo-v2.5-tts`）
- `--clone-voice-data-uri`：克隆音色数据（用于 `mimo-v2.5-tts-voiceclone`），格式 `data:{mime};base64,{audio}`
  - help/describe 中需明确说明：该格式为 Data URI / Data URL，规范来自 RFC 2397。
- `--stream`：`true/false`

help/describe 文案要求（`generate_speech`）：

- 明确区分两个文本参数语义：
  - `--assistant-text`："说什么"，即最终被合成为语音并朗读的正文。
  - `--user-text`："怎么说"，用于语气、情绪、节奏、情节/场景等上下文控制。
- 对 `mimo-v2.5-tts` 与 `mimo-v2.5-tts-voiceclone`：`--user-text` 为可选风格控制参数。
- 对 `mimo-v2.5-tts-voicedesign`：`--user-text` 还承担音色描述职责，应视为必需输入。
- 上述语义在 `help`/`describe` 中必须以英文描述表达。

### 4.4 suggest_voices 参数与行为

- `--model <model-id>`：可选。
  - 传入时：仅返回该模型的 `--voice` 建议映射。
  - 不传时：返回全部已记录的 model id 与对应建议映射。

该能力定位为“信息提示”，不是“参数约束”：

- 仅用于向用户展示建议值和兼容关系。
- 不阻止用户在 `generate_speech` 中传入新的 `--builtin-voice` 值。

`suggest_voices` 需要同时返回两类参数建议：

- `builtin_voice`：内置音色可选值与适用模型。
- `clone_voice_data_uri`：Data URI 前缀规范与 MIME 白名单。

建议同时支持：

- `tts-cli capability suggest_voices --help`
- `tts-cli capability describe suggest_voices`（结构化说明，含示例）

实现优先级：

- `suggest_voices` 为本版（MVP）必须实现的独立 capability，不延期到后续阶段。

可选扩展参数（第二阶段）：

- `--save-raw-pcm`
- `--sample-rate`（默认 24000，仅用于封装 WAV 时）

## 5. 核心流程设计

### 5.1 非流式流程

1. 参数校验。
2. 构建 `/chat/completions` 请求体。
3. 发起 HTTP POST。
4. 解析 `choices[0].message.audio.data`（Base64）。
5. Base64 解码并写文件。
6. 输出 JSON 结果。

### 5.2 流式流程

1. 参数校验（流式建议 `pcm16`）。
2. 发起 `stream=true` 请求。
3. 读取 SSE 事件。
4. 从每个 chunk 提取 `delta.audio.data`，解码后按序拼接。
5. 完成后写文件并输出 JSON 诊断（chunk 数、总字节数、耗时）。

## 6. 参数校验策略

1. 通用校验：
   - `assistant-text` 必填。
   - `timeout > 0`。
   - `model` 必须在允许集合内。

2. 模型特定校验：
   - `mimo-v2.5-tts-voicedesign`：要求 `user-text` 非空（用于音色描述）。
   - `mimo-v2.5-tts-voiceclone`：优先使用 `clone-voice-data-uri`；格式建议 `data:{mime};base64,...`。
   - `mimo-v2.5-tts`：使用 `builtin-voice`；为空时可由服务端默认音色处理。

3. Data URI MIME 固定建议（白名单，不强制拦截）：
   - `data:audio/mpeg;base64,...`
   - `data:audio/mp3;base64,...`
   - `data:audio/wav;base64,...`
   - 用于向用户给出明确格式说明与示例；未知 MIME 透传上游，由上游最终判定。

4. 兼容策略（建议映射不强制）：
   - 内置 model/voice mapping 只用于 `suggest_voices` 展示与 warning 提示。
   - 对未知 `--builtin-voice` 值不做本地拦截，直接透传上游。
   - 对 `clone-voice-data-uri` 的 MIME 若不在建议白名单，仅 warning，不强制报错。
   - 若值不在建议表，可在 `diagnostics.warnings` 给出提示，但不返回错误。

3. 流式校验：
   - 若 `stream=true` 且 `audio-format!=pcm16`，允许但给出 warning（按文档建议 `pcm16`）。

## 7. 错误模型设计

统一错误结构：

- `code`：
  - `config_error`
  - `invalid_args`
  - `network_error`
  - `rpc_error`
  - `parse_error`
  - `internal_error`
- `message`：可读错误信息。
- `retryable`：是否建议重试。

判定原则：

- 网络超时、连接中断：`retryable=true`
- 参数错误、鉴权失败：`retryable=false`

## 8. 输出与可观测性

统一 JSON 输出：

- `ok`
- `request`（关键参数回显）
- `result`（`output_file`、`response_id`、`audio_format`、`stream`）
- `diagnostics`（chunk 数、音频字节数、耗时毫秒）
- `error`（失败时）

agent 消费约束：

- 字段命名保持稳定，避免破坏提示词/流程编排的兼容性。
- 错误结构固定（`code/message/retryable`），便于 agent 做重试与降级决策。
- `diagnostics` 提供最小必要上下文（耗时、chunk、warning），避免 agent 额外猜测。

`suggest_voices` 输出建议包含：

- `models`: 已记录 model id 列表
- `voice_rules`: 每个模型的 `--builtin-voice` / `--clone-voice-data-uri` 使用规则
- `builtin_voices`: `mimo-v2.5-tts` 可选内置音色表
- `clone_voice_data_uri`: Data URI 结构说明与 MIME 白名单（`audio/mpeg`、`audio/mp3`、`audio/wav`）

`describe` 输出建议结构（对齐 proxmox-cli 习惯）：

- `command`: `capability`
- `action`: `describe`
- `capability`: capability 名称
- `summary`: capability 摘要
- `args.required`: 必填参数列表
- `args.optional`: 可选参数列表
- `examples`: 调用示例数组
- `help`: 人类可读帮助元信息

日志策略：

- 默认只输出结果 JSON。
- 不默认打印敏感信息（API Key）。

## 9. 目录与文件规划（建议）

```text
applications/tts-cli/
  cmd/tts-cli/main.go
  src/internal/cli/app.go
  src/internal/cli/help.go
  src/internal/capability/registry.go
  src/internal/capability/ops_generate.go
  src/internal/ttsapi/client.go
  src/internal/ttsapi/sse.go
  src/internal/output/render.go
  src/internal/apperr/errors.go
  api.md
  design.md
```

## 10. 迭代计划

### Phase 1（MVP）

- CLI 基础框架。
- `capability list` 实现（结构化列出已实现 capability，便于 agent 发现）。
- `capability suggest_voices` 实现（返回模型与 voice 建议映射）。
- `generate_speech` 非流式打通。
- 音频文件输出。
- JSON 结果输出与基础错误码。

### Phase 2

- 流式 SSE 支持与拼接。
- 三模型参数规则补齐（voicedesign/voiceclone）。
- 更完整诊断字段。

### Phase 3

- 增加回归脚本与文档样例。
- 增加兼容性增强（更多返回格式处理、失败回退策略）。

## 11. 风险与应对

1. 文档与实际返回字段存在差异。
   - 应对：解析逻辑保留多路径兼容，错误信息带上关键上下文。

2. 流式能力为兼容模式，实时性不稳定。
   - 应对：stream 模式按“分片拼接”实现，不承诺低延迟。

3. 音频格式与扩展名不一致。
   - 应对：由 `audio-format` 决定默认扩展名，并允许用户显式指定文件名。

4. 语音克隆输入体积较大。
   - 应对：本地预检查 Base64 长度，超限提前报错。

## 12. 验收标准

满足以下条件即视为首版可用：

1. agent 通过 `capability generate_speech` 可稳定调用 MiMo TTS 并生成音频文件。
2. 支持至少一个内置音色模型的非流式调用。
3. JSON 输出含成功与失败两类结构。
4. 参数错误、网络错误、上游错误可区分。
5. 不依赖 Python 或第三方 Python SDK。
6. `capability describe` / `--help` 可完整暴露参数语义，满足 agent 自发现需求。
