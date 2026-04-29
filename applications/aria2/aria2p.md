# aria2p 参考分析（面向 agent-cli 需求）

本文聚焦你的目标：**不是复刻 aria2p 的完整产品形态**，而是提炼“可被 agent 稳定调用”的基础 CLI 操作集合。

## 1. 先明确定位差异

- aria2p 定位：人类用户友好的 Python CLI + library（含 `top` 交互界面、`listen` 监听、丰富别名、日志等）。
- 我们的 aria2-cli 定位：agent 调用优先，强调**确定性输出、低交互、可组合、可自动化**。

结论：我们应当“借鉴 aria2p 的操作边界”，但要主动裁剪掉交互/展示导向能力。

## 2. aria2p 中值得模仿的基础操作

基于 `build/aria2p/src/aria2p/cli/parser.py`、`build/aria2p/src/aria2p/cli/commands/*.py`、`build/aria2p/src/aria2p/client.py`：

### 2.1 任务创建（必须）

- `add`：自动识别 URI / magnet / torrent / metalink
- `add-magnets`
- `add-torrents`
- `add-metalinks`

为什么要模仿：这是下载系统的入口能力，且 aria2p 已证明这组拆分稳定可用。

### 2.2 队列控制（必须）

- `pause` / `pause --all` / `pause --force`
- `resume` / `resume --all`
- `remove` / `remove --all` / `remove --force`
- `purge`（清理 completed/removed/failed 的历史结果）

为什么要模仿：这是“任务生命周期管理”最小闭环。

### 2.3 状态查询（必须）

- 单任务：`tellStatus`（按 gid）
- 列表：`tellActive`、`tellWaiting`、`tellStopped`
- 全局：`getGlobalStat`

为什么要模仿：agent 需要可轮询状态，才能编排 workflow。

### 2.4 透传调用（建议保留一个）

- `call`（直接调用 RPC method）

为什么要模仿：当我们尚未封装某个 capability 时，`call` 可以作为逃生阀，降低首版覆盖压力。

## 3. aria2p 中不建议直接照搬的部分

### 3.1 交互/界面导向能力

- `top`（TUI）
- `show`（面向人类阅读的表格输出）
- `listen`（长期监听通知、回调模块）

原因：这些能力更适合人工操作，不利于 agent 的短命令、结构化、幂等调用。

### 3.2 过强的“人类友好语法糖”

- 过多别名（`rm/del/delete`）
- 过强容错方法名匹配（`call` 子命令中去掉大小写/下划线/连字符）

原因：agent 场景更适合**明确、稳定、可预测**的命令面。

## 4. 建议我们在 aria2-cli 里保留的最小命令面

建议首版只做以下命令（可映射为 capability）：

1. `capability add_uri`
2. `capability add_magnet`
3. `capability add_torrent`
4. `capability add_metalink`
5. `capability pause`
6. `capability pause_all`
7. `capability resume`
8. `capability resume_all`
9. `capability remove`
10. `capability remove_all`
11. `capability purge_download_result`
12. `capability tell_status`
13. `capability list_active`
14. `capability list_waiting`
15. `capability list_stopped`
16. `capability get_global_stat`
17. `capability rpc_call`（保底）

## 5. 参数与输出规范（比 aria2p 更偏 agent）

### 5.1 参数规范

- 全局连接参数统一：`--rpc-endpoint`、`--rpc-secret`、`--timeout`
- 所有变更类命令支持 `--wait`（可选）
- 列表类命令支持 `--offset`、`--limit`
- 所有命令避免交互输入（不弹提示）

### 5.2 输出规范

- 默认 `--output json`
- 统一响应外壳：`ok`、`request`、`result`、`diagnostics`
- 错误统一：`code`、`message`、`retryable`

## 6. 从 aria2p 可直接借鉴的实现思路

### 6.1 RPC 方法映射表

`build/aria2p/src/aria2p/client.py` 已集中列出方法常量（如 `aria2.addUri`、`aria2.pause`、`aria2.tellActive` 等），可作为我们 Go 客户端方法清单。

### 6.2 secret 注入机制

aria2p 在调用时自动把 `token:<secret>` 插入参数（针对 `aria2.*` 方法）。这个细节必须保留，否则调用会频繁鉴权失败。

### 6.3 错误模型

aria2p 对 JSON-RPC 错误码（`-32601/-32602/...`）有明确转换，建议我们也保留标准错误码映射，便于 agent 做恢复策略。

## 7. 结论（面向当前需求）

你当前需求是“把基本操作变成可被 agent 调用的 CLI”。基于 aria2p，建议只采纳其：

- **任务创建**
- **生命周期控制**
- **状态查询**
- **保底 RPC 透传**

并显式放弃其人机交互导向能力（`top/show/listen`）作为首版范围外。这样可以最快得到一个与 `proxmox-cli` 风格一致、适合 workflow 编排的 `aria2-cli`。
