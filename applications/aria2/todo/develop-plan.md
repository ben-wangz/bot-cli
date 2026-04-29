# aria2-cli 开发计划（可执行版）

## 1. 目标与边界

## 目标

- 在本仓落地一个 **agent-first** 的 `aria2-cli`，用于通过 JSON-RPC 操作 `aria2c`。
- 命令可被 agent 稳定调用：默认结构化输出、低交互、可编排。
- 首版优先覆盖“基本操作闭环”：创建、控制、查询、清理、透传。

## 非目标（首版不做）

- 不做 TUI/交互界面（如 `top`）。
- 不做人类阅读优先表格展示（如 `show`）。
- 不做长期事件监听回调（如 `listen`）。

---

## 2. 参考输入（已确认）

- `applications/aria2/report.md`：定义了可对齐 `proxmox-cli` 的工程形态。
- `applications/aria2/aria2p.md`：提炼了 aria2p 中应模仿与应裁剪的能力边界。

---

## 3. 交付物清单

1. `applications/aria2-cli/src/`
   - Go 模块、可编译二进制入口
   - CLI 命令分发（至少 `capability`、`workflow`、`auth/help`）
   - aria2 JSON-RPC 客户端层
2. `applications/aria2-cli/skills/aria2-cli/SKILL.md`
   - agent 使用规范（JSON 输出、安全、幂等与等待策略）
3. `applications/aria2-cli/tests/prompts/`
   - 最小能力链路提示词测试
4. 文档
   - `README.md`（安装、配置、示例）
   - capability 与 workflow 目录说明

---

## 4. 首版能力范围（MVP）

## Capability（必须）

- 创建类
  - `add_uri`
  - `add_magnet`
  - `add_torrent`
  - `add_metalink`
- 控制类
  - `pause`
  - `pause_all`
  - `resume`
  - `resume_all`
  - `remove`
  - `remove_all`
  - `purge_download_result`
- 查询类
  - `tell_status`
  - `list_active`
  - `list_waiting`
  - `list_stopped`
  - `get_global_stat`
- 保底透传
  - `rpc_call`

## Workflow（建议首版）

- `queue_add_and_wait`
  - 添加任务后按 gid 轮询状态，直到完成/失败/超时
- `pause_resume_chain`
  - 对目标任务执行 pause -> resume 并验证状态变化
- `cleanup_completed`
  - 清理 stopped/completed 结果（必要时调用 purge）

---

## 5. 技术设计原则

## 5.1 命令契约

- 默认输出 `--output json`
- 统一响应结构：`ok`、`request`、`result`、`diagnostics`
- 统一错误结构：`code`、`message`、`retryable`

## 5.2 全局参数

- `--rpc-endpoint`
- `--rpc-secret`
- `--timeout`
- `--wait`（对变更类操作可选启用）

## 5.3 鉴权与协议细节

- 对 `aria2.*` 方法自动注入 `token:<secret>` 参数。
- 映射 JSON-RPC 标准错误码（如 `-32601/-32602/-32603`）到统一错误模型。

## 5.4 与 proxmox-cli 对齐

- 保持 `capability/workflow` 双层调用模型。
- 保持 deterministic 输出语义，便于 agent 脚本化处理。

---

## 6. 实施计划（精简版）

## Step 1：搭骨架并打通 RPC

任务：

- 建立 `applications/aria2-cli/` 基础目录与 Go 入口。
- 实现 CLI 分发（`capability` / `workflow`）。
- 实现 JSON-RPC 客户端、secret 注入、统一错误模型。

完成标准：

- `aria2-cli --help` 可用。
- `capability rpc_call` 能调用 `aria2.getVersion`。
- 错误输出符合统一 JSON 结构。

## Step 2：实现 MVP 基础能力

任务：

- 完成创建/控制/查询/清理能力（MVP 清单全部）。
- 支持 `pause/resume/remove` 的单任务与 all 变体。
- 支持 `offset/limit` 查询参数与 `--wait` 可选等待。

完成标准：

- MVP capability 全部可执行。
- add -> status -> pause -> resume -> remove -> purge 链路可跑通。

## Step 3：补 workflow、skill、文档与最小测试

任务：

- 完成 3 条 workflow：`queue_add_and_wait`、`pause_resume_chain`、`cleanup_completed`。
- 编写 `skills/aria2-cli/SKILL.md` 与 README。
- 增加最小 prompts 测试（主路径 + 关键错误路径）。

完成标准：

- workflow 输出可直接被 agent 消费。
- skill 可被加载。
- 最小提示词回归通过。

---

## 7. 风险与对策

- 风险：RPC 请求体过大（torrent/metalink base64）。
  - 对策：文档提示 `--rpc-max-request-size`，并给出错误诊断信息。
- 风险：远端环境差异导致状态字段不一致。
  - 对策：响应解析做宽容处理，关键字段缺失进入 diagnostics。
- 风险：队列轮询导致 workflow 超时不稳定。
  - 对策：统一 `--wait-timeout` 与 `--wait-interval` 参数，给出可调默认值。

---

## 8. 验收标准（发布前）

- 功能验收
  - MVP capability 全部可用且返回结构一致。
  - 3 条 workflow 可运行并可复现。
- 工程验收
  - 目录结构与 `proxmox-cli` 风格一致。
  - Skill 可加载，README 可独立指导使用。
- 可运维验收
  - 常见故障具备明确错误消息与恢复建议。

---

## 9. 建议执行顺序

1. 先打通 `rpc_call + tell_status + get_global_stat`
2. 再完成 `add_* + pause/resume/remove + purge`
3. 最后补 `workflow + skill + prompts tests`
