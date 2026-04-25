# proxmox-cli 开发方案（基于 A01-A43）

- 状态: draft
- 优先级: high
- 创建日期: 2026-04-18
- 需求来源: `/root/code/geekcity-k8s/pve/todo/proxmox-cli-requirements.md`

## 1. 目标与交付范围

本方案用于把需求文档中的结论落地为可执行开发计划，目标是交付一个可独立运行的 `proxmox-cli`，覆盖 A01-A43 全量 action，并可跑通 Ubuntu 24.04 串口自动安装闭环。

### 1.1 核心目标

1. 提供统一 CLI：`action`、`workflow`、`console`、`auth` 四类命令。
2. 完整实现并可验证 A01-A43（约束类 action 也要有明确输出）。
3. 内建三挡凭据能力：`PVE_USER < PVE_ROOT_API_TOKEN < PVE_ROOT`。
4. 跑通完整 workflow：上传 ISO -> 串口 autoinstall -> 串口 post-install -> qga 取 IP -> SSH 验证 hostname。

### 1.2 非目标（当前阶段不做）

1. 不做 Web UI 或 VNC 图形自动化。
2. 不做多云抽象层，范围仅 Proxmox。
3. 不做复杂 ACL 配置管理，只做权限校验与错误引导。

## 2. 需求剖析与关键约束

## 2.1 不可妥协的约束

1. Action-first：所有能力先抽象为 action，workflow 只是编排。
2. 串口优先：安装与运维默认走 `termproxy + vncwebsocket`（cookie 模式）。
3. 异步任务闭环：写操作必须跟踪 `UPID`，支持 status/log/tasks 诊断。
4. 最小权限分层：默认 `PVE_USER`，仅在必要步骤升级到 `PVE_ROOT`。
5. failure-first：先诊断后重试，禁止盲等。
6. 可复现：请求参数、会话方式、seed 生成和串口协议都可重放。

## 2.2 需求中最关键的技术点

1. API 双认证：既要支持 Token，也要支持 Session Ticket/Cookie。
2. `args` 写入权限特殊：需求已验证需 `root@pam` 会话。
3. `sendkey` 必须使用 `PUT`。
4. storage upload 需识别 `snippets` 限制并给出可理解报错。
5. 串口 websocket 要实现 xterm 帧协议（输入/窗口/keepalive）。
6. 初始环境依赖必须可代码化创建：测试环境不依赖人工预置镜像/服务。

## 2.3 初始环境依赖开发要求（新增）

1. 对任何“先决环境能力”（如 qga-ready 基础镜像、seed 上传能力、串口会话入口），必须提供 action/workflow 级自动化创建路径。
2. 测试用例默认假设“环境可由当前仓库自举”；若暂不可自举，必须在 issue 中显式标记 blocker 与解除路径。
3. 禁止将长期依赖固化为“手工前置步骤”；手工步骤仅允许短期过渡，并需在后续 issue 中被替换为可执行动作。
4. 回归入口应优先验证“环境创建动作”再验证业务动作，确保在新环境可重复恢复。

## 3. 总体技术设计

实现语言确定为 Go（单二进制、并发与网络库成熟、CLI 生态稳定），按“命令层-应用层-基础设施层”分层。

## 3.1 分层架构

1. 命令层（CLI）
   - 解析全局参数和子命令。
   - 统一 `--help`、输出格式、exit code。
2. 应用层（Use Cases）
   - Action Registry（A01-A43）。
   - Workflow Engine（步骤编排、条件分支、错误收敛）。
3. 基础设施层（Infra）
   - Proxmox HTTP Client（API + 重试 + 脱敏日志）。
   - Task Client（`wait_task`、`task_log`、`list_tasks`）。
   - Console Client（termproxy + websocket + xterm 编解码）。
4. 安全与配置层
   - Credential Loader（flag/env/auth-file 合并与优先级）。
   - Secret Redactor（日志与错误统一脱敏）。

## 3.2 建议目录结构

```text
applications/proxmox-cli/
  cmd/proxmox-cli/
  internal/
    auth/
    cli/
    action/
    workflow/
    pveapi/
    console/
    task/
    output/
    diagnose/
  test/
    fixtures/
    integration/
  todo/
    proxmox-cli-development-plan.md
```

## 3.3 统一执行模型

每个 action 统一返回以下结构，保证可观测和可编排：

```json
{
  "action": "create_vm",
  "ok": true,
  "scope": "user",
  "request": {"node": "eva003", "vmid": 120},
  "result": {"upid": "UPID:..."},
  "diagnostics": {
    "task_status": "running",
    "task_log_hint": "...",
    "next_suggested_actions": ["get_task_status", "list_tasks_by_vmid"]
  }
}
```

## 4. CLI 设计与帮助文档方案

## 4.1 顶层命令

1. `proxmox-cli action <name>`：执行单 action。
2. `proxmox-cli workflow <name>`：执行组合流程。
3. `proxmox-cli console <subcommand>`：串口会话与交互。
4. `proxmox-cli auth <subcommand>`：认证、权限、凭据连通性检查。

## 4.2 全局参数（与需求对齐）

1. `--api-base`
2. `--auth-scope <user|root-token|root>`
3. `--auth-file`
4. `--insecure-tls`
5. `--timeout`
6. `--wait`
7. `--dry-run`
8. `--output <json|yaml|table>`

## 4.3 Help 交付标准

1. `action --help`：必须列出 A01-A43。
2. `action <name> --help`：参数、默认值、权限挡位、至少两个示例。
3. `workflow ubuntu24-serial-autoinstall --help`：输入、前置条件、诊断点、输出契约。

## 4.4 Console 模式范围（已确认）

1. 先实现脚本模式为主，用于自动化闭环执行与可复现回归。
2. 同步实现交互模式基础版，用于人工接管和现场排障。
3. 首期不做高级终端能力（会话录制回放、多会话共享控制、复杂宏脚本）。
4. 脚本模式必须内置超时、关键字等待、失败摘要输出。

## 5. A01-A43 实现拆解（分阶段）

## 5.1 Phase 0：工程底座（先行）

交付：CLI 框架、配置加载、HTTP 基础 client、输出模块、日志脱敏。

## 5.2 Phase 1：读类与任务基础能力

覆盖：A01 A02 A03 A04 A05 A06 A11 A12 A21

目标：先打通只读与任务观测，为后续写操作提供诊断底座。

## 5.3 Phase 2：VM 生命周期与配置写操作

覆盖：A07 A08 A09 A10 A13 A14 A23 A24 A25 A26 A28 A30 A31

关键点：

1. 所有写操作统一返回 `UPID` 并可 `--wait`。
2. `sendkey` 固化为 `PUT`，错误方法给明确提示。
3. 对 URL 编码参数（`net0`,`boot`）做统一编码器。

## 5.4 Phase 3：Cloud-init 与 guest agent

覆盖：A15 A17 A20 A22 A27

关键点：

1. `render_and_serve_seed` 封装 seed 生成和本地托管。
2. `storage_upload_guard` 前置检查 `snippets` 受限场景。
3. qga 相关 action 要有“未安装/未启动”可读错误。
4. `storage_upload_guard` 仅做守卫与提示，不耦合上传写入；上传建议独立 action。
5. 若无 qga-ready 模板，QGA 回归依赖 Phase 4 串口最小能力先行（A29/A34）。
6. Live 环境补充：PVE upload API 对 `snippets` 存在限制（仅 `iso|vztmpl|import` 可传），snippet 自动落盘需后续 root 路径支持。

## 5.5 Phase 4：串口与 websocket 控制面

覆盖：A18 A19 A29 A32 A34 A40

关键点：

1. 实现 termproxy 握手与 websocket 认证行发送。
2. 实现 xterm 帧收发与 keepalive。
3. 支持 expect 风格等待关键字（含超时和日志片段回显）。

## 5.6 Phase 5：root 辅助用户授权（收敛后）

覆盖：P5-00 P5-01 P5-02 P5-03 P5-04

关键点：

1. 项目主路径不依赖 root；常规 VM 管理统一走 user 凭据。
2. root 仅用于一次性 bootstrap：创建 user、绑定角色、下发 ACL。
3. 提供 pool+user+ACL 管理闭环：创建 pool、查询授权、授权新增、授权撤销（变更通过 revoke+grant 组合，幂等）。
4. 历史 root shell 能力不再作为验收依赖路径。

## 6. 工作流方案（ubuntu24-serial-autoinstall）

将需求中的 12 步闭环落地为可重放 workflow：

1. ISO 获取与上传（含镜像源 fallback）。
2. seed 生成与 HTTP 托管。
3. VM 创建与硬件初始化。
4. 串口配置与启动顺序设置。
5. root 节点 shell 提取 kernel/initrd。
6. root 会话写入 `args`（autoinstall + nocloud-net）。
7. 启动 VM 并附着串口 websocket。
8. 监控安装执行（diskwrite + tasks + serial）。
9. 安装后串口执行 qga 安装启用。
10. 注入 SSH 资产与权限修正。
11. 使用 qga 获取业务 IP。
12. SSH `hostname` 验证并输出结果。

每一步都要求产出：`inputs`、`action list`、`success signal`、`failure diagnostics`、`rollback/next step`。

## 7. 认证与安全方案

## 7.1 凭据读取优先级

`CLI flags > process env > --auth-file`

## 7.2 安全约束

1. `PVE_ROOT_PASSWORD` 支持仅环境变量注入（不落盘）。
2. 日志和错误输出统一脱敏（token/password/cookie/ticket）。
3. `--debug` 也不允许输出 secret 原文。

## 7.3 权限守卫

1. action 元数据声明 `required_scope`。
2. 执行前校验 scope，避免“执行后才报权限错误”。
3. 对可升级路径给自动建议（例如从 user 切换到 root）。

## 8. 测试与验收方案

## 8.1 测试层次

1. 单元测试
   - URL 编码器、xterm 帧编解码、脱敏器、参数校验器。
2. 集成测试（Mock）
   - API 响应映射、UPID 轮询、失败场景分支。
3. 环境测试（Live PVE）
   - A01-A43 覆盖清单执行。
   - ubuntu24 workflow 端到端执行。

## 8.2 验收门槛

1. A01-A43 全部可执行或可验证输出明确。
2. 帮助文档满足“参数解释 + 实例 + 权限挡位”。
3. 能稳定跑通 hostname 最终校验。
4. 失败路径诊断可在一次输出中定位问题方向。

## 8.3 Action 级正向主流程测试（OpenCode Prompt）

要求：每个 action 都提供一条独立的“正向主流程”测试 prompt，由 OpenCode 执行并回传结果。当前阶段只覆盖成功路径，不覆盖异常分支。

### 8.3.1 Prompt 模板

```text
你是测试执行代理。请在 proxmox-cli 项目中执行 action 正向测试。

目标 action: <ACTION_NAME>
命令: proxmox-cli action <ACTION_NAME> <ARGS> --auth-scope <SCOPE> --output json

执行要求:
1) 先打印将执行的完整命令（脱敏后）。
2) 执行命令并捕获原始输出。
3) 判断是否成功（exit code=0 且 JSON 中 ok=true）。
4) 如果 result 中有 upid，继续调用:
   - proxmox-cli action get_task_status --node <NODE> --upid <UPID> --output json
   并附上状态摘要。
5) 仅输出以下结构:
   - action
   - command
   - success
   - key_result
   - diagnostics
```

### 8.3.2 交付要求

1. 在 `applications/proxmox-cli/tests/prompts/` 下为 A01-A43 生成 43 条 prompt 文件。
2. 文件命名建议：`A01-list_nodes.prompt.md` ... `A43-cleanup_obsolete_artifact.prompt.md`。
3. 每个 action 仅保留 1 条独立正向测试用例，不做多套参数预设。
4. 每条 prompt 固定包含：前置条件、命令、成功判定、输出模板。
5. 若 action 有前置依赖，可在前置条件中声明“依赖 action 已通过”，无需在同一 prompt 内重复编排全链路。
6. workflow 另有独立 e2e prompt，不替代 action 级测试。
7. 为支持 sub-agent 并发执行，action prompt 必须严格独立，不得依赖执行顺序或其他 prompt 的输出/副作用；每条 prompt 需在本次执行内自行解析所需输入（如 node/vmid/upid）。

## 9. 里程碑计划（建议）

1. M1（2-3 天）: Phase 0 + Phase 1。
2. M2（3-5 天）: Phase 2 + Phase 3。
3. M3（3-4 天）: Phase 4。
4. M4（2-3 天）: Phase 5。
5. M5（2 天）: 全量联调、文档补齐、验收回归。

## 10. 风险与缓解

1. 串口行为在不同 PVE 版本可能存在差异。
   - 缓解：协议解析容错 + 原始帧可选落盘。
2. `args` 权限限制导致 workflow 中断。
   - 缓解：前置权限探测 + 自动切换建议。
3. seed 可达性导致 autoinstall 不触发。
   - 缓解：内置 seed 探活和访问日志联动诊断。
4. qga 未正确安装影响 IP 获取。
   - 缓解：串口执行 post-install 模块默认启用。

## 11. 已确认实现策略

1. Action 实现优先级按 Phase 顺序推进（Phase 0 -> Phase 5）。
2. A01-A43 测试 prompt 不做多套模板，每个 action 仅 1 条独立正向用例。
3. 对存在依赖关系的 action，测试前置假设“依赖 action 已通过”，并在 prompt 中显式声明。

## 12. Live PVE 使用约束（已确认）

1. 连接信息来源：`~/code/github/bot-cli/build/pve-user.env`。
2. 可在任意时段使用 Live PVE 环境。
3. 严格约束：只允许操作本次开发/测试中由我创建的 VM，不操作他人 VM。
4. 资源使用原则：单节点在叠加本次测试需求后，CPU/Memory/Storage 任一项不超过节点总资源 70% 即可使用。
5. 每次回归最多使用 1 台测试 VM。
6. 回归 VM 规格固定：2C / 4G / 32G（如无特殊说明按此默认）。
7. 每次回归结束后自动清理测试 VM 与临时镜像/seed 文件。
8. 执行策略：每次 workflow 或批量 action 测试前先做节点容量快照评估，超阈值则降载或延后执行。

## 13. 当前讨论结论

1. 语言：Go。
2. Console：脚本模式为主，交互模式基础版。
3. Workflow：不支持 `--resume-from`。
4. 测试：每个 active action 1 条独立正向 prompt；可声明依赖前置 action 已通过。
5. Live PVE：可任意时段使用，但仅操作自建 VM；节点资源占用不超过 70%。
6. 回归约束：每次最多 1 台 VM，规格 2C/4G/32G，结束自动清理 VM 与临时镜像/seed。
7. 并发测试原则：若采用 sub-agent 并发，单条 action prompt 必须满足顺序无关和数据自洽。
8. 对 VM 生命周期类写操作（创建/迁移/模板转换/配置变更），即使启用 sub-agent，也应限制为单写并发（有效并发=1），并要求每条用例使用本地新分配 VMID 与独立清理。
9. 安全范围控制：PVE 写操作使用环境变量 `PVE_ALLOWED_VMID_MIN` 与 `PVE_ALLOWED_VMID_MAX` 限定可操作 VMID 区间，默认范围 `1001-2000`。

---

在以上约束下，方案讨论项已收敛完毕。下一步进入“Phase 0 任务拆解”，输出首批 issue 列表（按 action 分组并包含验收标准）。
