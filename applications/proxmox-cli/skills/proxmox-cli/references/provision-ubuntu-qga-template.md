# Provision Ubuntu QGA Template

Use this flow to build an Ubuntu autoinstall ISO artifact, install a VM, and convert it into a reusable template with QGA support.

## Preconditions

1. source ISO available locally
2. `PVE_POOL` set for scoped operations
3. user scope credentials with required pool/storage rights

## Step Chain

Use resolved binary path (`PROXMOX_CLI_BIN`):

1. Choose online node (`list_nodes`).
2. Guard upload target (`storage_upload_guard`).
3. Build autoinstall ISO (`build_ubuntu_autoinstall_iso`).
4. Upload artifact ISO (`storage_upload_iso`).
5. Run provisioning workflow (`provision-template-from-artifact`).

Example workflow call:

```bash
"${PROXMOX_CLI_BIN}" ... workflow provision-template-from-artifact \
  --node "$NODE" \
  --target-vmid "$TEMPLATE_VMID" \
  --artifact-iso "$ARTIFACT_ISO" \
  --install-timeout-seconds 1800 \
  --resume-from none \
  --pool "$PVE_POOL"
```

## Timeout/Resume Pattern

If install wait exceeds timeout and there is no explicit installer failure:

1. inspect install tasks
2. resume from `serial_wait`

```bash
"${PROXMOX_CLI_BIN}" ... capability review_install_tasks --node "$NODE" --vmid "$TEMPLATE_VMID"
"${PROXMOX_CLI_BIN}" ... workflow provision-template-from-artifact --node "$NODE" --target-vmid "$TEMPLATE_VMID" --artifact-iso "$ARTIFACT_ISO" --install-timeout-seconds 600 --resume-from serial_wait --pool "$PVE_POOL"
```

## Post-Validation

Require all of the following:

1. `get_vm_config` reports template flag enabled
2. cloned disposable VM boots from template
3. `agent_network_get_interfaces` succeeds (QGA available)
