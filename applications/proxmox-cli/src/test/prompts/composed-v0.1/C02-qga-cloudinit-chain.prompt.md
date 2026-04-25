# C02 qga cloud-init chain

## Goal

Run one chain to cover:

- `render_and_serve_seed`
- `dump_cloudinit`
- `agent_network_get_interfaces`
- `agent_exec`
- `agent_exec_status`

## Prompt

```text
You are a test execution agent. Execute qga/cloud-init chain.

Setup:
1) Load `build/pve-user.env`, switch to `applications/proxmox-cli/src`.
2) Resolve and start one disposable VM cloned from template.
3) Wait for QGA readiness by polling `agent_network_get_interfaces`.

Chain:
1) `dump_cloudinit`.
2) `render_and_serve_seed` with deterministic local output under `build/`.
3) `agent_network_get_interfaces` and extract non-loopback IPv4.
4) `agent_exec --command hostname`.
5) `agent_exec_status` for returned pid until completion.

Validation:
- All actions return `ok == true`.
- `agent_exec_status` ends with completed/exit code 0.

Cleanup:
- Stop and destroy disposable VM.
- Remove temporary seed artifacts created in this prompt.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
