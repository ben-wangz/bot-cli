# Provision Template From Artifact

## Purpose

Run workflow `provision-template-from-artifact` with explicit prerequisites in the same prompt:

1) build local Ubuntu autoinstall ISO (`build_ubuntu_autoinstall_iso`)
2) upload ISO to PVE storage (`storage_upload_iso`, with `storage_upload_guard` first)
3) provision template from uploaded artifact (`provision-template-from-artifact`)

## Prompt

```text
You are a coding/test execution agent. Execute end-to-end template provisioning from uploaded ISO artifact.

Execution requirements:
1) Work from repository root.
2) Load env vars from `build/pve-user.env`.
3) Use source dir `applications/proxmox-cli/src`.
4) Persist all progress logs to a file under `build/logs/`.
5) Use API base `${PVE_API_BASE_URL%/}/api2/json` and always include `--insecure-tls --output json`.
6) Set VMID policy bounds explicitly:
   - `PVE_ALLOWED_VMID_MIN=1001`
   - `PVE_ALLOWED_VMID_MAX=2000`
7) For ISO upload command, set CLI timeout to `--timeout 20m`.
   - rationale: `3 GiB / 3 MiB/s ~= 1024s ~= 17m04s`, so 20m leaves safety margin.

Input resolution:
7) Resolve `NODE` from first online node in `capability list_nodes` output.
8) Resolve a fresh `TARGET_VMID` without using `capability get_next_vmid`.
   - query existing VMIDs via `capability list_cluster_resources --type vm`
   - select the first free VMID in range `1001..2000`
   - if no free VMID exists in this range, stop and report `vmid_range_exhausted`.
9) Set deterministic local artifact paths:
   - `LOCAL_SOURCE_ISO=build/ubuntu-24.04.4-live-server-amd64.iso`
   - `LOCAL_OUTPUT_ISO=build/e2e-provision-artifact.iso`
   - `LOCAL_WORK_DIR=build/autoinstall-iso-work/e2e-provision-artifact`
   - `UPLOAD_FILENAME=e2e-provision-artifact.iso`

Pre-checks:
10) Ensure `LOCAL_SOURCE_ISO` exists.
11) Run storage guard:
    - `go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json capability storage_upload_guard --node "$NODE" --storage local --content-type iso`

Step A - Build installer ISO (A52):
12) Execute:
    - `go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json capability build_ubuntu_autoinstall_iso --source-iso "$LOCAL_SOURCE_ISO" --output-iso "$LOCAL_OUTPUT_ISO" --work-dir "$LOCAL_WORK_DIR"`
13) Validate:
    - `ok == true`
    - output file `$LOCAL_OUTPUT_ISO` exists and non-empty

Step B - Upload installer ISO (A53), split with progress markers:
14) Print marker before upload, then execute upload:
    - marker: `[STEP B1] about to upload ISO`
    - `go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --timeout 20m --output json capability storage_upload_iso --node "$NODE" --storage local --source-path "$LOCAL_OUTPUT_ISO" --filename "$UPLOAD_FILENAME" --if-exists replace`
15) Print marker after upload returns:
    - marker: `[STEP B2] upload completed, extracting volid`
16) Extract `ARTIFACT_ISO` from result `volid` (fallback to `local:iso/$UPLOAD_FILENAME` if empty).
17) Validate:
    - `ok == true`
    - `ARTIFACT_ISO` matches `<storage>:iso/<file>.iso` pattern

Step C - Provision template from uploaded artifact:
18) Execute workflow:
    - `go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json workflow provision-template-from-artifact --node "$NODE" --target-vmid "$TARGET_VMID" --artifact-iso "$ARTIFACT_ISO" --install-timeout-seconds 3600 --resume-from none`
19) If step 18 times out in serial wait, switch to resume windows (`--install-timeout-seconds 600 --resume-from serial_wait`) until success or non-timeout failure.
    Before each resume attempt, do fail-fast inspection first:
    - resolve `SERIAL_LOG_PATH` from latest workflow result (`result.serial_log_path`, fallback `build/serial-provision-template-$TARGET_VMID.log`)
    - run `go run ./cmd/proxmox-cli --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json capability review_install_tasks --node "$NODE" --vmid "$TARGET_VMID"`
    - if any install task has `state` in `{error,failed}` or contains non-empty `error`, stop and report `installer_error_detected` (do not keep resuming)
    - scan latest serial log segment; if fatal markers appear (`subiquity.*ERROR`, `curtin command .* failed`, `installation failed`, `Traceback`, `No space left on device`), stop and report `installer_error_detected`
    - only run next resume attempt when checks above are clean
20) Validate workflow result:
    - `workflow == "provision-template-from-artifact"`
    - `ok == true`
    - `result.template_vmid == TARGET_VMID`
    - `result.serial_log_path` exists
    - `result.steps` includes at least:
      - `attach_cdrom_iso` (install)
      - `serial_ws_capture_to_file`
      - `detach_cdrom_iso`
      - `convert_vm_to_template`

Return only structured output:
- workflow
- command
- success
- node
- target_vmid
- artifact_iso
- serial_log_path
- log_file
- diagnostics

Safety:
- Do not modify or delete non-test VMs.
- If provisioning fails after VM creation, report vmid and log path for manual cleanup.
```
