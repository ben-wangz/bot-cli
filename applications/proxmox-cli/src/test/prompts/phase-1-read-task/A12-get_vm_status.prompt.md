# A12 get_vm_status

## Preconditions

- `build/pve-user.env` is loaded.
- `<NODE>` and `<VMID>` point to an existing VM.

## Prompt

```text
你是测试执行代理。请在 bot-cli 仓库执行 A12 `get_vm_status` 正向测试。

执行前置:
1) source build/pve-user.env
2) 进入 applications/proxmox-cli/src
3) 设置 NODE 与 VMID

执行命令:
go run ./cmd/proxmox-cli --output json action get_vm_status --node <NODE> --vmid <VMID>

成功判定:
- exit code = 0
- JSON 中 `action == "get_vm_status"`
- JSON 中 `ok == true`
- `request.node` 与 `request.vmid` 与输入一致

仅输出以下结构:
- action
- command
- success
- key_result
- diagnostics
```
