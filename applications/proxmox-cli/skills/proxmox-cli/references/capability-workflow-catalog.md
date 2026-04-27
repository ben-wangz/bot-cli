# Capability and Workflow Catalog

This reference summarizes high-value command groups for agents.

## Workflows

1. `bootstrap-bot-user-pool-acl`
   - purpose: create/reuse bot user and pool, then grant ACL
   - scope: root
2. `provision-template-from-artifact`
   - purpose: provision Ubuntu from artifact ISO and convert to template
   - scope: usually user (with required pool/storage rights)

## Capability Groups

### Inventory and Task

- `list_nodes`
- `list_cluster_resources`
- `list_vms_by_node`
- `get_vm_config`
- `get_vm_status`
- `get_task_status`
- `get_next_vmid` (returns VMID inside allowed operation range)
- `list_tasks_by_vmid`

### VM Lifecycle

- `clone_template`
- `create_vm`
- `update_vm_config`
- `vm_power`
- `destroy_vm`
- `migrate_vm`
- `convert_vm_to_template`

### Guest/QGA

- `agent_network_get_interfaces`
- `agent_exec`
- `agent_exec_status`

### Storage and ISO

- `storage_upload_guard`
- `build_ubuntu_autoinstall_iso`
- `storage_upload_iso`

### Access Control

- `create_pve_user_with_root`
- `create_pool_with_root`
- `get_user_acl_binding`
- `grant_user_acl`
- `revoke_user_acl`

### SSH Control Plane

SSH capabilities run in batch/key mode (`BatchMode=yes`). Interactive password prompt mode is not supported in capability path.

- `ssh_check_service`
- `ssh_inject_pubkey_qga`
- `ssh_exec`
- `ssh_scp_transfer`
- `ssh_print_connect_command`
- `ssh_tunnel_start`
- `ssh_tunnel_status`
- `ssh_tunnel_stop`

SSH usage examples:

- `ssh_check_service`
  - `proxmox-cli capability ssh_check_service --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519`
- `ssh_exec`
  - `proxmox-cli capability ssh_exec --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519 --timeout-seconds 30 --command "hostname"`
- `ssh_scp_transfer` upload
  - `proxmox-cli capability ssh_scp_transfer --direction upload --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519 --local-path ./artifact --remote-path /tmp/artifact`
- `ssh_scp_transfer` download
  - `proxmox-cli capability ssh_scp_transfer --direction download --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519 --local-path ./build/log.txt --remote-path /var/log/cloud-init.log`

## Execution Notes

1. Prefer `workflow` when objective is end-to-end and standardized.
2. Prefer `capability` for composable custom chains.
3. Parse JSON responses and enforce `ok == true` at each step.
4. For task-producing capabilities, track `upid` and poll status when strict completion matters.
