# aria2-cli 前期调研报告

## 1) `applications/proxmox-cli/` 结构拆解

已确认 `proxmox-cli` 是一个“可执行 CLI + agent skill + 文档与测试提示词”的完整应用骨架，核心结构如下：

- `applications/proxmox-cli/src/`
  - Go 模块（`go.mod`）
  - 入口：`cmd/proxmox-cli/main.go`
  - 业务实现：`internal/`
    - `cli/`：命令分发与全局参数（如 `capability` / `workflow` / `auth`）
    - `capability/`：原子能力（VM、存储、SSH、控制台、ACL 等）
    - `workflow/`：多步骤编排（如 bootstrap、模板制备）
    - `pveapi/`、`taskwait/`、`output/`、`auth/` 等基础层
- `applications/proxmox-cli/skills/proxmox-cli/SKILL.md`
  - 为 OpenCode agent 提供可加载技能定义（frontmatter + 使用指南 + 安全原则）
  - 说明 capability/workflow 目录化能力与调用规范
- `applications/proxmox-cli/tests/prompts/`
  - 面向 agent 的提示词测试资产（setup/workflow/composed chains）
- 其它：`README.md`、`assets/`、`release/`、`VERSION`

结论：`proxmox-cli` 当前模式可抽象为：

1. **CLI 内核**（命令分发 + 能力注册 + 输出协议）
2. **能力层**（原子 RPC/操作）
3. **工作流层**（面向场景的编排）
4. **Agent Skill 层**（让智能体“知道何时、如何、安全地调用 CLI”）
5. **提示词测试层**（回归与链路验证）

这可以直接作为 `aria2-cli` 的脚手架参考。

## 2) aria2 现有 agent/cli/skill 生态调研

### 2.1 官方能力基线（aria2）

- 官方仓库：`https://github.com/aria2/aria2`
- 官方文档明确提供 RPC 接口（JSON-RPC / XML-RPC，含 HTTP 与 WebSocket）：
  - `https://aria2.github.io/manual/en/html/aria2c.html#rpc-interface`

这意味着我们做 `aria2-cli` 时，推荐路线是：

- 使用 `aria2c --enable-rpc` 作为 daemon
- 新 CLI 通过 JSON-RPC 调用（与 `proxmox-cli` 的 capability 思想一致）

### 2.2 已有第三方 CLI 实现（可借鉴）

发现一个成熟的第三方实现：

- `aria2p`（Python，CLI + library）
  - PyPI：`https://pypi.org/project/aria2p/`
  - GitHub：`https://github.com/pawamoy/aria2p`
  - 定位：通过 JSON-RPC 操作 `aria2c`，提供 `add/pause/resume/remove/show/call/listen/top` 等命令

说明：`aria2` 生态中“RPC 客户端型 CLI”是可行且成熟的方向。

### 2.3 是否已有“agent-cli / OpenCode skill”

本地仓库内检查结果：

- 当前仅有 `proxmox-cli` 应用，未发现 `aria2` 相关应用/代码/skill 文件。
- 全仓检索 `aria2`、`SKILL.md` 相关关键词，均无命中。

网络侧检查结果（尽量充分）：

- OpenCode skills 官方文档可确认 skill 机制与目录规范：
  - `https://opencode.ai/docs/skills`
- 在公开检索中，**未发现明确、通用、可直接复用的“aria2 OpenCode skill”标准实现**。
- 发现的是 aria2 通用客户端/前端（如 `aria2p`、`AriaNg`），但它们不是 OpenCode agent skill。

受限说明：

- GitHub 部分搜索接口对未登录/匿名抓取有限流（429）与权限限制，导致“全网穷尽式代码搜索”存在客观不完备。
- 但基于已拿到的高相关来源（官方文档 + 主流第三方客户端 + OpenCode skills 文档 + 本仓检索），结论可信度较高。

## 3) 对 `aria2-cli` 的落地建议（基于现状）

建议先按 `proxmox-cli` 模式做一个最小可用版本（MVP）：

1. `applications/aria2-cli/src/`：Go CLI 内核（命令分发 + JSON 输出）
2. capability 首批：
   - `add_uri`
   - `tell_active`
   - `tell_waiting`
   - `pause`
   - `unpause`
   - `remove`
   - `get_global_stat`
3. workflow 首批：
   - `queue_add_and_wait`
   - `pause_resume_chain`
   - `cleanup_completed`
4. `applications/aria2-cli/skills/aria2-cli/SKILL.md`：定义 agent 使用原则（默认 JSON 输出、secret 处理、幂等操作优先）
5. `tests/prompts/`：补最基础的链路提示词测试

---

以上为本轮“先调研再报告”的结果，文件已落地：`applications/aria2/report.md`。
