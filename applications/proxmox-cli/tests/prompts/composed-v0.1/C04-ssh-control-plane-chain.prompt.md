# C04 ssh control-plane chain

## Goal

Run one chain to cover:

- `ssh_check_service`
- `ssh_inject_pubkey_qga`
- `ssh_exec`
- `ssh_scp_transfer`
- `ssh_print_connect_command`
- `ssh_tunnel_start`
- `ssh_tunnel_status`
- `ssh_tunnel_stop`

## Prompt

```text
You are a test execution agent. Execute SSH control-plane chain.

Setup:
1) Load `build/pve-user.env`, switch to `applications/proxmox-cli/src`.
2) Clone and boot one disposable VM from template (ensure clone uses `--pool "$PVE_POOL"` and a deterministic disposable name, for example `botcli-c04-<vmid>`).
3) Generate temporary ed25519 keypair under `build/`.
4) Inject public key via `ssh_inject_pubkey_qga` (use `--pub-key-file` or `--pub-key`).

Chain:
1) `ssh_check_service`.
2) `ssh_exec --command hostname`.
3) `ssh_scp_transfer` upload then download, compare file content.
4) `ssh_print_connect_command` and validate command structure.
5) `ssh_tunnel_start` + `ssh_tunnel_status` + `ssh_tunnel_stop`.

Validation:
- All capabilities return `ok == true`.
- SCP roundtrip content matches.
- Tunnel status reports running before stop.

Cleanup:
- Cleanup must run in `finally` even if SSH checks fail.
- Stop VM best-effort, then destroy disposable VM via Proxmox API delete (`purge=1`, `destroy-unreferenced-disks=1`).
- If VM not found at cleanup time, treat as already cleaned.
- Delete temporary key files and transfer artifacts.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
