# Quickstart

This is a portable quickstart for any coding agent using `proxmox-cli`.

## 1) Load Environment

Set environment variables from your own secret source (env file, vault, CI secret store):

- user scope: `PVE_API_BASE_URL`, `PVE_USER`, `PVE_PASSWORD`
- root scope (only for ACL/bootstrap): `PVE_ROOT_USER`, `PVE_ROOT_PASSWORD`

Do not hardcode credentials in scripts.

## 2) Resolve Binary (Download if Missing)

Use this generic GitHub Releases bootstrap snippet:

```bash
PROXMOX_CLI_BIN="${PROXMOX_CLI_BIN:-$(command -v proxmox-cli 2>/dev/null || true)}"
PROXMOX_CLI_GH_REPO="${PROXMOX_CLI_GH_REPO:-<owner>/<repo>}"
PROXMOX_CLI_VERSION="${PROXMOX_CLI_VERSION:-0.1.1}"

if [ -z "${PROXMOX_CLI_BIN}" ]; then
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo "unsupported arch: $arch" >&2; exit 1 ;;
  esac

  tag="proxmox-cli-v${PROXMOX_CLI_VERSION}"
  asset="proxmox-cli_${os}_${arch}"
  base="https://github.com/${PROXMOX_CLI_GH_REPO}/releases/download/${tag}"

  mkdir -p ./.bin
  curl -fsSL -o ./.bin/proxmox-cli "${base}/${asset}"
  curl -fsSL -o ./.bin/checksums.txt "${base}/checksums.txt"
  (cd ./.bin && sha256sum --check --ignore-missing checksums.txt)
  chmod +x ./.bin/proxmox-cli
  PROXMOX_CLI_BIN="$(pwd)/.bin/proxmox-cli"
fi
```

If your environment does not provide `sha256sum`, install it first or use an equivalent SHA256 verifier.

Optional (install OpenCode skill bundle from the same release):

```bash
tag="proxmox-cli-v${PROXMOX_CLI_VERSION}"
skill_asset="proxmox-cli_skills_${PROXMOX_CLI_VERSION}.tar.gz"
base="https://github.com/${PROXMOX_CLI_GH_REPO}/releases/download/${tag}"

mkdir -p ./build ./.opencode/skills
curl -fsSL -o ./build/proxmox-cli-skills.tar.gz "${base}/${skill_asset}"
tar -xzf ./build/proxmox-cli-skills.tar.gz -C ./.opencode/skills
```

## 3) Verify Connectivity and Identity

```bash
"${PROXMOX_CLI_BIN}" \
  --api-base "${PVE_API_BASE_URL%/}/api2/json" \
  --insecure-tls \
  --output json \
  --auth-scope user \
  --auth-user "$PVE_USER" \
  --auth-password "$PVE_PASSWORD" \
  capability list_nodes
```

Expect:

- `ok == true`
- at least one online node in `result`

## 4) Resolve Baseline Inputs

Before any mutation chain, resolve:

1. source template VMID
2. free target VMID in policy range
3. source and target nodes (for migration scenarios)
4. target pool (`PVE_POOL`)

## 5) Run a Small Validation Chain

Recommended smoke chain:

1. `clone_template`
2. `vm_power` start
3. `agent_network_get_interfaces`
4. `vm_power` stop
5. destroy disposable VM

## 6) Parse Responses Strictly

Require:

- top-level `ok == true`
- no semantic failure fields in `result`
- for task-based calls, verify terminal task state when strict completion is needed
