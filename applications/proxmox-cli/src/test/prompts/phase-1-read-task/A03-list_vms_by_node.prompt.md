# A03 list_vms_by_node

## Preconditions

- `build/pve-user.env` is loaded.
- `<NODE>` is a valid PVE node name.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A03 `list_vms_by_node` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src
3) 设置 NODE（例如 eva003）

执行命令:
go run ./cmd/proxmox-cli --output json action list_vms_by_node --node <NODE>

成功判定:
- exit code = 0
- JSON 中 `action == "list_vms_by_node"`
- JSON 中 `ok == true`
- `request.node` 与输入一致

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
