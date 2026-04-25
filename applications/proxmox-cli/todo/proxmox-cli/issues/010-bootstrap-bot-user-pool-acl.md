# ISSUE-010 bootstrap-bot-user-pool-acl

- status: completed
- priority: high
- owner: ben.wangz

## Goal

通过 Phase 5 action 一次性完成“测试账号 + 测试 pool + 最小补充读权限”的引导，输出可被 agent 直接消费的标准 workflow JSON。

## Scope

- command: `proxmox-cli workflow bootstrap-bot-user-pool-acl`
- module: `applications/proxmox-cli/src/internal/cli`
- output: CLI workflow standard JSON payload (`workflow/ok/scope/request/result/diagnostics`)

## Inputs

- required:
  - `--userid <user@realm>`
  - `--poolid <poolid>`
- optional:
  - `--password <plain>` (unset => auto-generate)
  - `--pool-comment <text>`
  - `--user-comment <text>`
  - `--if-exists <fail|reuse>` (default: `reuse`)

## Workflow Plan (Action-first)

1. 创建指定用户
   - action: `create_pve_user_with_root`
   - request: `--userid <user@realm> --password <...> --if-exists <...>`
2. 创建指定 pool
   - action: `create_pool_with_root`
   - request: `--poolid <poolid> --if-exists <...>`
3. 赋予该用户对 pool 的 `PVEAdmin`
   - action: `grant_user_acl`
   - request: `--userid <user@realm> --path /pool/<poolid> --role PVEAdmin --propagate 1`
4. 赋予可完成 A01/A02 的最小补充权限
   - action: `grant_user_acl`
   - request: `--userid <user@realm> --path / --role PVEAuditor --propagate 1`
   - rationale: A01/A02 依赖节点/集群读能力，当前内置角色下使用 `PVEAuditor@/` 作为最小可行补充。
5. 赋予可完成 A22/A27 与 ISO 上传链路的最小补充权限
   - action: `grant_user_acl`
   - request: `--userid <user@realm> --path /storage --role PVEDatastoreAdmin --propagate 1`
   - rationale: `PVEAuditor` 仅满足只读探测，不足以执行 `storage_upload_iso` 上传；`PVEDatastoreAdmin@/storage` 可覆盖上传所需 Datastore 写权限。A27 为本地 seed 渲染，不额外消耗 PVE API 权限。

## Output Contract

遵循现有 workflow 输出规范，不直接写入 `build/pve-user.env`。返回 JSON：

- `workflow`: `bootstrap-bot-user-pool-acl`
- `ok`: `true|false`
- `scope`: 执行时 auth scope
- `request`: 输入参数回显
- `result`:
  - `userid`
  - `poolid`
  - `password`
  - `password_generated`
  - `grants[]`（每次授权的 path/role/propagate）
  - `actions[]`（每个步骤对应 action、ok、关键结果）
- `diagnostics`: 执行步骤与校验摘要

预期 `result.grants[]` 至少包含：

- `/pool/<poolid>` + `PVEAdmin` + `propagate=1`
- `/` + `PVEAuditor` + `propagate=1`
- `/storage` + `PVEDatastoreAdmin` + `propagate=1`

## Verification

至少执行以下校验 action：

1. `get_user_acl_binding --userid <user@realm>`：应包含三类授权
   - `/pool/<poolid>` + `PVEAdmin`
   - `/` + `PVEAuditor`
   - `/storage` + `PVEDatastoreAdmin`
2. `get_effective_permissions --path /pool/<poolid>`：应具备 VM 管理权限。
3. `get_effective_permissions --path /`：应至少包含只读审计能力（支持 A01/A02）。
4. `storage_upload_guard`（A22）正向执行应返回可读能力结论。

## Notes

- 该 workflow 为 root-assisted bootstrap；执行完成后应回到 user 凭据执行常规动作。
- e2e prompt: `applications/proxmox-cli/tests/prompts/workflows/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md`
