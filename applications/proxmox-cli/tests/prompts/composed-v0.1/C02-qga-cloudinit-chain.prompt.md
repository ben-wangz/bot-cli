# C02 qga cloud-init chain

## Goal

Run one chain to cover:

- `agent_network_get_interfaces`
- `agent_exec`
- `agent_exec_status`

## Prompt

```text
You are a test execution agent. Execute qga/cloud-init chain.

Setup:
1) Load `build/pve-user.env`, switch to `applications/proxmox-cli/src`.
2) Resolve and start one disposable VM cloned from template (ensure clone uses `--pool "$PVE_POOL"`).
3) Wait for QGA readiness by polling `agent_network_get_interfaces`.

Chain:
1) `agent_network_get_interfaces` and extract non-loopback IPv4.
2) `agent_exec --command hostname`.
3) `agent_exec_status` for returned pid until completion.

Validation:
- All capabilities return `ok == true`.
- `agent_exec_status` ends with completed/exit code 0.

Cleanup:
- Stop and destroy disposable VM.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
