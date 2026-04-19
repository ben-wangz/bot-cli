# A05 get_effective_permissions

## Preconditions

- `build/pve-user.env` is loaded.
- `<PATH>` is a valid ACL path (default `/`).

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A05 `get_effective_permissions` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src

执行命令:
go run ./cmd/proxmox-cli --output json action get_effective_permissions --path <PATH>

成功判定:
- exit code = 0
- JSON 中 `action == "get_effective_permissions"`
- JSON 中 `ok == true`
- `request.path` 存在

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
