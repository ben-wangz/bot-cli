# Setup: Prepare `build/pve-user.env` and template baseline

## Purpose

Provide one stable setup prompt for regression runs. This setup uses root bootstrap credentials to create a fresh bot user + pool, grants required ACLs, writes `build/pve-user.env`, and prepares `build/ubuntu-24-with-agent.vm-template.id` via the artifact provisioning chain for subsequent action suites.

## Prompt

```text
You are a coding/test execution agent. Prepare `build/pve-user.env` from `build/pve-root.env` for regression testing, then prepare a reusable template id file for action prompts.

Execution requirements:
1) Work from repository root.
2) Load root credentials from `build/pve-root.env`.
3) Build deterministic-per-run identifiers:
   - `RUN_ID="$(date +%Y%m%d-%H%M%S)"`
   - `BOT_USERID="botcli-reg-${RUN_ID}@pve"`
   - `BOT_POOLID="botcli-reg-${RUN_ID}-pool"`
   - `BOT_PASSWORD="$(python3 -c 'import secrets,string; s=string.ascii_letters+string.digits; print("".join(secrets.choice(s) for _ in range(20)))')"`
4) If `build/pve-user.env` exists, back it up to `build/pve-user.env.bak-${RUN_ID}`.
5) Execute workflow:
   - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope root --output json workflow bootstrap-bot-user-pool-acl --userid "$BOT_USERID" --poolid "$BOT_POOLID" --password "$BOT_PASSWORD" --if-exists fail --user-comment "regression bot user" --pool-comment "regression bot pool" --sdn-acl-path "/sdn/zones/localnetwork"`
6) Validate workflow response:
   - `workflow == "bootstrap-bot-user-pool-acl"`
   - `ok == true`
   - `result.userid == BOT_USERID`
   - `result.poolid == BOT_POOLID`
   - `result.grants` contains:
      - `/pool/${BOT_POOLID}` + `PVEAdmin`
      - `/` + `PVEAuditor`
      - `/storage` + `PVEDatastoreAdmin`
      - `/sdn/zones/localnetwork` + `PVEAdmin`
7) Write `build/pve-user.env` with exactly these keys:
   - `PVE_USER=${BOT_USERID}`
   - `PVE_PASSWORD=${BOT_PASSWORD}`
   - `PVE_POOL=${BOT_POOLID}`
   - `PVE_ACL_PATH=/pool/${BOT_POOLID}`
   - `PVE_API_BASE_URL=${PVE_API_BASE_URL}`
   - `PVE_API_SSL_SELF_SIGNED=1`
8) Verify new user env works:
   - `source build/pve-user.env`
   - run `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action get_effective_permissions --path "/pool/${PVE_POOL}"`
   - ensure `ok == true` and result for `/pool/${PVE_POOL}` is non-empty.
9) Prepare template baseline directly in setup via artifact provisioning chain (do not call any bootstrap prompt file):
   - resolve `NODE` from first online node in `action list_nodes`
   - resolve `TEMPLATE_VMID` as first free VMID in range `1001..2000` from `action list_cluster_resources --type vm` (do not use `get_next_vmid`)
   - set artifact paths:
     - `LOCAL_SOURCE_ISO=build/ubuntu-24.04.4-live-server-amd64.iso`
     - `LOCAL_OUTPUT_ISO=build/e2e-provision-artifact.iso`
     - `LOCAL_WORK_DIR=build/autoinstall-iso-work/e2e-provision-artifact`
     - `UPLOAD_FILENAME=e2e-provision-artifact.iso`
   - ensure `LOCAL_SOURCE_ISO` exists
    - run prerequisite actions in order (auth-scope user):
      - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action storage_upload_guard --node "$NODE" --storage local --content-type iso`
      - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action build_ubuntu_autoinstall_iso --source-iso "$LOCAL_SOURCE_ISO" --output-iso "$LOCAL_OUTPUT_ISO" --work-dir "$LOCAL_WORK_DIR"`
      - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action storage_upload_iso --node "$NODE" --storage local --source-path "$LOCAL_OUTPUT_ISO" --filename "$UPLOAD_FILENAME" --if-exists replace`
    - resolve `ARTIFACT_ISO` from upload result `volid` (fallback `local:iso/$UPLOAD_FILENAME`)
    - immediately delete local temporary ISO artifact after successful upload:
      - `rm -f "$LOCAL_OUTPUT_ISO"`
      - require local file no longer exists; if delete fails, stop and report `local_iso_cleanup_failed`
    - remove local autoinstall work directory as intermediate artifact:
      - `rm -rf "$LOCAL_WORK_DIR"`
      - require directory no longer exists; if delete fails, stop and report `local_workdir_cleanup_failed`
    - execute workflow (first attempt):
      - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json workflow provision-template-from-artifact --node "$NODE" --target-vmid "$TEMPLATE_VMID" --artifact-iso "$ARTIFACT_ISO" --install-timeout-seconds 1800 --resume-from none --pool "$PVE_POOL"`
    - if first attempt times out in serial wait, continue with resume mode using 600s windows until success or non-timeout failure:
      - before each resume attempt, inspect install health and fail fast on known error states:
        - resolve `SERIAL_LOG_PATH` from latest workflow result `result.serial_log_path` (fallback `build/serial-provision-template-$TEMPLATE_VMID.log`)
        - run `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action review_install_tasks --node "$NODE" --vmid "$TEMPLATE_VMID"`
        - if review result reports any `state` in `{error,failed}` or any task with non-empty `error`, stop and report `installer_error_detected`
        - scan the latest serial log segment (or full file if small) and if fatal markers are found, stop and report `installer_error_detected`; suggested markers include: `subiquity.*ERROR`, `curtin command .* failed`, `installation failed`, `Traceback`, `No space left on device`
      - only when checks above do not indicate fatal failure, run resume:
        - `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json workflow provision-template-from-artifact --node "$NODE" --target-vmid "$TEMPLATE_VMID" --artifact-iso "$ARTIFACT_ISO" --install-timeout-seconds 600 --resume-from serial_wait --pool "$PVE_POOL"`
    - require final workflow `ok == true` and `result.template_vmid == TEMPLATE_VMID`
10) Write and verify template-id artifact:
   - write `build/ubuntu-24-with-agent.vm-template.id` with exact content `$TEMPLATE_VMID` (single line)
   - read back the file and require value equals workflow `result.template_vmid`
    - validate target VM is template via `go run ./applications/proxmox-cli/src/cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --auth-scope user --output json action get_vm_config --node "$NODE" --vmid "$TEMPLATE_VMID"` and require `result.template == 1`
   - if any check fails, stop and report `template_id_file_invalid`
11) Return only structured output:
   - setup
   - workflow
   - success
   - backup_file
   - pve_user_env_path
   - template_id_file
   - template_vmid
   - userid
   - poolid
   - diagnostics
   - node

Safety:
- This setup is the only prompt that depends on `build/pve-root.env`.
- Do not modify unrelated users/pools.
```
