# C03 serial websocket chain

## Goal

Run one chain to cover:

- `open_vm_termproxy`
- `validate_k1_serial_readable`
- `serial_ws_session_control`
- `validate_serial_output_criterion2`

## Prompt

```text
You are a test execution agent. Execute serial/websocket control-plane chain.

Setup:
1) Load `build/pve-user.env`, switch to `applications/proxmox-cli/src`.
2) Clone and boot one disposable VM from template.

Chain:
1) `open_vm_termproxy`.
2) `validate_k1_serial_readable`.
3) `serial_ws_session_control` with bounded timeout.
4) `validate_serial_output_criterion2` against captured output.

Validation:
- All capabilities return `ok == true`.
- Session and validation capabilities produce expected readable output markers.

Cleanup:
- Stop and destroy disposable VM.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
