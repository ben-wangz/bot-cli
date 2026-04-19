# A02 list_cluster_resources

## Preconditions

- `build/pve-user.env` is present and valid.
- At least one cluster resource exists.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A02 `list_cluster_resources` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src

执行命令:
go run ./cmd/proxmox-cli --output json action list_cluster_resources --type vm

成功判定:
- exit code = 0
- JSON 中 `action == "list_cluster_resources"`
- JSON 中 `ok == true`
- `result` 为数组

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
