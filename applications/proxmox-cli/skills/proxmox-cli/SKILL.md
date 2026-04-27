---
name: proxmox-cli
description: |
  Operate Proxmox VE with proxmox-cli: bootstrap user/pool ACL, provision Ubuntu
  templates with QGA, and manage VM lifecycle through capability/workflow commands.
license: MIT
compatibility: opencode
metadata:
  audience: coding-agents
  tool: proxmox-cli
---

# Proxmox CLI Skill

Use this skill when an agent needs to operate PVE virtual machines safely and repeatably from the CLI.

## What proxmox-cli Provides

`proxmox-cli` has two execution layers:

1. `capability`: focused operations (inventory, VM, task, guest/QGA, SSH, storage, access).
2. `workflow`: multi-step orchestration for common end-to-end jobs.

Core workflows:

- `bootstrap-bot-user-pool-acl`
- `provision-template-from-artifact`

Representative capability families:

- Inventory/task: `list_nodes`, `list_cluster_resources`, `get_vm_config`, `get_task_status`, `list_tasks_by_vmid`
- VM lifecycle: `clone_template`, `vm_power`, `update_vm_config`, `migrate_vm`, `convert_vm_to_template`
- Guest/QGA: `agent_network_get_interfaces`, `agent_exec`, `agent_exec_status`
- Access control: `create_pve_user_with_root`, `create_pool_with_root`, `get_user_acl_binding`, `grant_user_acl`, `revoke_user_acl`
- SSH control plane: `ssh_check_service`, `ssh_inject_pubkey_qga`, `ssh_exec`, `ssh_scp_transfer`, `ssh_tunnel_start`
- Storage/ISO: `storage_upload_guard`, `build_ubuntu_autoinstall_iso`, `storage_upload_iso`

## Operating Principles for Agents

1. Always use `--output json` for deterministic parsing.
2. Use least privilege by default (`--auth-scope user` for routine operations).
3. Use `--auth-scope root` only for bootstrap/ACL administrative tasks.
4. Treat async operations as task-based: capture `upid`, then poll `get_task_status` if strict completion is required.
5. Keep disposable VM lifecycle self-contained: create, validate, then clean up.

## Base Command Template

Prefer a resolved binary path for repeatable runs:

```bash
PROXMOX_CLI_BIN="${PROXMOX_CLI_BIN:-$(command -v proxmox-cli 2>/dev/null || true)}"

if [ -z "${PROXMOX_CLI_BIN}" ]; then
  echo "proxmox-cli not found. Resolve/download binary first." >&2
  echo "See: references/binary-bootstrap-and-release-download.md" >&2
  exit 1
fi

"${PROXMOX_CLI_BIN}" \
  --api-base "${PVE_API_BASE_URL%/}/api2/json" \
  --insecure-tls \
  --output json \
  --auth-scope user \
  --auth-user "$PVE_USER" \
  --auth-password "$PVE_PASSWORD" \
  capability list_nodes
```

Fallback for source-only environments:

```bash
cd "<proxmox-cli-source-root>/src"
go run ./cmd/proxmox-cli ...
```

For root bootstrap, switch scope and credentials to root equivalents.

## Quick Task Routing

- Need bot identity and pool ACL? -> [Bootstrap user/pool ACL](references/bootstrap-user-pool-acl.md)
- Need Ubuntu template with QGA? -> [Provision Ubuntu QGA template](references/provision-ubuntu-qga-template.md)
- Need clone/start/migrate/cleanup chains? -> [VM lifecycle playbook](references/vm-lifecycle.md)
- Need deterministic binary bootstrap/download strategy? -> [Binary bootstrap and release download](references/binary-bootstrap-and-release-download.md)
- Need command catalog and semantics? -> [Capability/workflow catalog](references/capability-workflow-catalog.md)
- Need common failures and recovery? -> [Troubleshooting](references/troubleshooting.md)
- Need minimal run recipe? -> [Quickstart](references/quickstart.md)

## Binary-First Recommendation

For repeatable agent runs, prefer prebuilt release binaries over `go run` when possible.

Recommended order:

1. Resolve/download a trusted local binary (with version gate + checksum verification).
2. Reuse the local binary path for all subsequent calls in one run.
3. Fall back to `go run ./cmd/proxmox-cli` only when local binary is unavailable.
4. Keep command flags identical across binary and `go run` execution modes.

Release note: the same release also publishes `proxmox-cli_skills_<version>.tar.gz` for skill installation.

See full pattern and script template in [Binary bootstrap and release download](references/binary-bootstrap-and-release-download.md).

## Safety Checklist

Before mutation operations, confirm:

1. Auth scope is correct (`root` only when required).
2. Target VMID and pool are expected and in policy range.
3. Every step checks `ok == true` in JSON.
4. Async task completion is validated for critical steps.
5. Disposable artifacts (VMs, temp keys, temp files) are removed at end.
