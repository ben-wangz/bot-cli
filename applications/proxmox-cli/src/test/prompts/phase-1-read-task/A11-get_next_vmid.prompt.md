# A11 get_next_vmid

## Preconditions

- `build/pve-user.env` is loaded.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A11 `get_next_vmid` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src

执行命令:
go run ./cmd/proxmox-cli --output json action get_next_vmid

成功判定:
- exit code = 0
- JSON 中 `action == "get_next_vmid"`
- JSON 中 `ok == true`
- `result.next_vmid` 为正整数

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
