# A01 list_nodes

## Preconditions

- `applications/proxmox-cli/src` can build successfully.
- Live env is available in `build/pve-user.env`.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A01 `list_nodes` 正向测试。

执行前置:
1) 加载环境变量: source build/pve-user.env
2) 工作目录切换到 applications/proxmox-cli/src

执行命令:
go run ./cmd/proxmox-cli --output json action list_nodes

成功判定:
- 进程 exit code = 0
- 返回 JSON 中 `action == "list_nodes"`
- 返回 JSON 中 `ok == true`
- `result` 为数组，且至少包含 1 个节点对象

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
