# A06 get_task_status

## Preconditions

- `build/pve-user.env` is loaded.
- `<NODE>` and `<UPID>` are valid and accessible.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A06 `get_task_status` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src
3) 准备一个可查询的 UPID

执行命令:
go run ./cmd/proxmox-cli --output json action get_task_status --node <NODE> --upid <UPID>

成功判定:
- exit code = 0
- JSON 中 `action == "get_task_status"`
- JSON 中 `ok == true`
- `diagnostics` 至少包含 `status` 或 `exitstatus`

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
