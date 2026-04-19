# A21 list_tasks_by_vmid

## Preconditions

- `build/pve-user.env` is loaded.
- `<NODE>` and `<VMID>` are valid.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A21 `list_tasks_by_vmid` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src
3) 设置 NODE 与 VMID

执行命令:
go run ./cmd/proxmox-cli --output json action list_tasks_by_vmid --node <NODE> --vmid <VMID>

成功判定:
- exit code = 0
- JSON 中 `action == "list_tasks_by_vmid"`
- JSON 中 `ok == true`
- `result` 为数组

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
