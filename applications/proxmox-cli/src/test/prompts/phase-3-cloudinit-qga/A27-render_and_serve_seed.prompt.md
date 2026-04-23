# A27 render_and_serve_seed

## Preconditions

- `build/pve-user.env` is loaded.
- `build/phase3-cloudinit-qga.shared-vmid` exists.

## Prompt

```text
You are a test execution agent. Run the A27 `render_and_serve_seed` positive-path test.

Setup:
1) Load env vars and switch to `applications/proxmox-cli/src`.
2) Read `SHARED_VMID` from `build/phase3-cloudinit-qga.shared-vmid`.
3) Prepare `build/regression-phase3/seed` for output.

Command:
go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json action render_and_serve_seed --vmid "$SHARED_VMID" --seed-dir build/regression-phase3/seed --seed-name "p3-$SHARED_VMID" --host 127.0.0.1 --port 18088

Success criteria:
- exit code = 0
- JSON field `action == "render_and_serve_seed"`
- JSON field `ok == true`
- JSON contains `result.seed_path`
- JSON contains `result.seed_url`
- JSON contains `result.serve_address`

Independence rule:
- This test must be self-contained and order-independent.
- Never depend on files created by other prompt runs.
- Do not create, stop, or destroy VMs inside this prompt run.
```
