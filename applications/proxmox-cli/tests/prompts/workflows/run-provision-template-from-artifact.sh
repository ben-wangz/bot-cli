#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${REPO_ROOT}" ]]; then
  REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../../../" && pwd)"
fi
SRC_DIR="${REPO_ROOT}/applications/proxmox-cli/src"
ENV_FILE="${REPO_ROOT}/build/pve-user.env"
LOG_DIR="${REPO_ROOT}/build/logs"
RUN_ID="$(date +%Y%m%d-%H%M%S)"

NODE_OVERRIDE=""
LOG_FILE=""
INSTALL_TIMEOUT_SECONDS="3600"
RESUME_FROM="none"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --node)
      NODE_OVERRIDE="${2:-}"
      shift 2
      ;;
    --log-file)
      LOG_FILE="${2:-}"
      shift 2
      ;;
    --install-timeout-seconds)
      INSTALL_TIMEOUT_SECONDS="${2:-}"
      shift 2
      ;;
    --resume-from)
      RESUME_FROM="${2:-}"
      shift 2
      ;;
    *)
      echo "unsupported arg: $1" >&2
      exit 2
      ;;
  esac
done

if [[ ! "${INSTALL_TIMEOUT_SECONDS}" =~ ^[0-9]+$ ]] || [[ "${INSTALL_TIMEOUT_SECONDS}" -le 0 ]]; then
  echo "install-timeout-seconds must be a positive integer" >&2
  exit 2
fi
if [[ "${RESUME_FROM}" != "none" && "${RESUME_FROM}" != "serial_wait" ]]; then
  echo "resume-from must be one of none|serial_wait" >&2
  exit 2
fi

mkdir -p "${LOG_DIR}"
if [[ -z "${LOG_FILE}" ]]; then
  LOG_FILE="${LOG_DIR}/provision-template-from-artifact-${RUN_ID}.log"
fi

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

echo "log_file=${LOG_FILE}"
exec >>"${LOG_FILE}" 2>&1

if [[ ! -f "${ENV_FILE}" ]]; then
  log "missing env file: ${ENV_FILE}"
  exit 2
fi

set -a
source "${ENV_FILE}"
set +a

export PVE_ALLOWED_VMID_MIN=1001
export PVE_ALLOWED_VMID_MAX=2000

if [[ -z "${PVE_API_BASE_URL:-}" ]]; then
  log "PVE_API_BASE_URL is empty"
  exit 2
fi

cli_json() {
  (
    cd "${SRC_DIR}"
    go run ./cmd/proxmox-cli \
      --api-base "${PVE_API_BASE_URL%/}/api2/json" \
      --insecure-tls \
      --timeout 20m \
      --output json \
      "$@"
  )
}

log "[STEP A0] resolving node"
if [[ -n "${NODE_OVERRIDE}" ]]; then
  NODE="${NODE_OVERRIDE}"
else
  NODE="$(cli_json capability list_nodes | python3 -c 'import sys,json; d=json.load(sys.stdin); r=d.get("result",[]); print(next((x.get("node") for x in r if x.get("status")=="online" and x.get("node")), ""))')"
fi
if [[ -z "${NODE}" ]]; then
  log "no online node found"
  exit 3
fi
log "resolved NODE=${NODE}"

log "[STEP A1] resolving target vmid in range 1001..2000"
TARGET_VMID="$(cli_json capability list_cluster_resources --type vm | python3 -c 'import sys,json; d=json.load(sys.stdin); used=set();
for x in d.get("result",[]):
  try: used.add(int(x.get("vmid")))
  except Exception: pass
print(next((str(i) for i in range(1001,2001) if i not in used), ""))')"
if [[ -z "${TARGET_VMID}" ]]; then
  log "vmid_range_exhausted"
  exit 3
fi
log "resolved TARGET_VMID=${TARGET_VMID}"

LOCAL_SOURCE_ISO="${REPO_ROOT}/build/ubuntu-24.04.4-live-server-amd64.iso"
LOCAL_OUTPUT_ISO="${REPO_ROOT}/build/e2e-provision-artifact.iso"
LOCAL_WORK_DIR="${REPO_ROOT}/build/autoinstall-iso-work/e2e-provision-artifact"
UPLOAD_FILENAME="e2e-provision-artifact.iso"

if [[ ! -f "${LOCAL_SOURCE_ISO}" ]]; then
  log "source iso not found: ${LOCAL_SOURCE_ISO}"
  exit 2
fi

log "[STEP A2] storage_upload_guard"
cli_json capability storage_upload_guard --node "${NODE}" --storage local --content-type iso >/dev/null

log "[STEP A3] build_ubuntu_autoinstall_iso"
cli_json capability build_ubuntu_autoinstall_iso --source-iso "${LOCAL_SOURCE_ISO}" --output-iso "${LOCAL_OUTPUT_ISO}" --work-dir "${LOCAL_WORK_DIR}" >/dev/null
if [[ ! -s "${LOCAL_OUTPUT_ISO}" ]]; then
  log "built iso missing or empty: ${LOCAL_OUTPUT_ISO}"
  exit 3
fi

log "[STEP B1] about to upload ISO"
UPLOAD_JSON="$(cli_json capability storage_upload_iso --node "${NODE}" --storage local --source-path "${LOCAL_OUTPUT_ISO}" --filename "${UPLOAD_FILENAME}" --if-exists replace)"
printf '%s\n' "${UPLOAD_JSON}" >> "${LOG_FILE}"

log "[STEP B2] upload completed, extracting volid"
ARTIFACT_ISO="$(printf '%s' "${UPLOAD_JSON}" | python3 -c 'import sys,json; d=json.load(sys.stdin); print((d.get("result") or {}).get("volid") or "")')"
if [[ -z "${ARTIFACT_ISO}" ]]; then
  ARTIFACT_ISO="local:iso/${UPLOAD_FILENAME}"
fi
log "resolved ARTIFACT_ISO=${ARTIFACT_ISO}"

log "[STEP C1] workflow provision-template-from-artifact"
workflow_args=(workflow provision-template-from-artifact --node "${NODE}" --target-vmid "${TARGET_VMID}" --artifact-iso "${ARTIFACT_ISO}" --install-timeout-seconds "${INSTALL_TIMEOUT_SECONDS}" --resume-from "${RESUME_FROM}")
if [[ -n "${PVE_POOL:-}" ]]; then
  workflow_args+=(--pool "${PVE_POOL}")
  log "[STEP C1] using pool=${PVE_POOL}"
fi
WORKFLOW_JSON="$(cli_json "${workflow_args[@]}")"
printf '%s\n' "${WORKFLOW_JSON}" >> "${LOG_FILE}"

SERIAL_LOG_PATH="$(printf '%s' "${WORKFLOW_JSON}" | python3 -c 'import sys,json; d=json.load(sys.stdin); print((d.get("result") or {}).get("serial_log_path") or "")')"

log "[DONE] workflow completed"
printf '%s\n' "{\"workflow\":\"provision-template-from-artifact\",\"success\":true,\"node\":\"${NODE}\",\"target_vmid\":${TARGET_VMID},\"artifact_iso\":\"${ARTIFACT_ISO}\",\"serial_log_path\":\"${SERIAL_LOG_PATH}\",\"log_file\":\"${LOG_FILE}\"}"
