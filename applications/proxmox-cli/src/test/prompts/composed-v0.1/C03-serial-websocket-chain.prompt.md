# C03 serial websocket chain

## Goal

Run one chain to cover:

- `start_vnc_proxy`
- `connect_vnc_websocket`
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
1) `start_vnc_proxy`.
2) `connect_vnc_websocket`.
3) `open_vm_termproxy`.
4) `validate_k1_serial_readable`.
5) `serial_ws_session_control` with bounded timeout.
6) `validate_serial_output_criterion2` against captured output.

Validation:
- All actions return `ok == true`.
- Session and validation actions produce expected readable output markers.

Cleanup:
- Stop and destroy disposable VM.

Return:
- `chain`, `command`, `success`, `key_result`, `diagnostics`.
```
