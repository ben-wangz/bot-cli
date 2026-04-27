# Binary Bootstrap and Release Download

This guide describes a portable binary bootstrap pattern for coding agents.

## Goal

Guarantee a usable `proxmox-cli` binary before running VM operations.

## Recommended Resolution Order

1. Use `PROXMOX_CLI_BIN` if explicitly provided.
2. Else try `command -v proxmox-cli`.
3. Else download from GitHub Releases.
4. Verify checksum.
5. Reuse the resolved path for all commands in the run.

## Required Inputs

- `PROXMOX_CLI_GH_REPO` in `owner/repo` format (example: `my-org/proxmox-cli`)
- `PROXMOX_CLI_VERSION` semver without `v` (example: `0.1.1`)

Optional:

- `PROXMOX_CLI_BIN_DIR` (default: `./build/bin`)
- `PROXMOX_CLI_DOWNLOAD_BASE` for mirror/proxy base URL

Guardrails:

1. Prefer a path already covered by repository `.gitignore` (recommended: `./build/bin` in this repo).
2. If using a custom path, ensure it is ignored before download.

## Release Asset Convention

This guide assumes release artifacts are published as:

- tag: `proxmox-cli-v<semver>`
- assets:
  - `proxmox-cli_linux_amd64`
  - `proxmox-cli_linux_arm64`
  - `proxmox-cli_darwin_amd64`
  - `proxmox-cli_darwin_arm64`
  - `checksums.txt`
  - `proxmox-cli_skills_<semver>.tar.gz`

If your project uses different names, adapt the `tag`/`asset` variables.

## Optional: Download Skill Bundle

If you also need the OpenCode skill package from the same release:

```bash
tag="proxmox-cli-v${PROXMOX_CLI_VERSION}"
skill_asset="proxmox-cli_skills_${PROXMOX_CLI_VERSION}.tar.gz"
base="https://github.com/${PROXMOX_CLI_GH_REPO}/releases/download/${tag}"

mkdir -p ./build ./.opencode/skills
curl -fsSL -o ./build/proxmox-cli-skills.tar.gz "${base}/${skill_asset}"
tar -xzf ./build/proxmox-cli-skills.tar.gz -C ./.opencode/skills
```

The archive contains `skills/proxmox-cli/...`, so extraction creates `./.opencode/skills/proxmox-cli`.

## Portable Bootstrap Snippet

```bash
#!/usr/bin/env bash
set -euo pipefail

PROXMOX_CLI_BIN="${PROXMOX_CLI_BIN:-$(command -v proxmox-cli 2>/dev/null || true)}"
PROXMOX_CLI_GH_REPO="${PROXMOX_CLI_GH_REPO:-}"
PROXMOX_CLI_VERSION="${PROXMOX_CLI_VERSION:-}"
PROXMOX_CLI_BIN_DIR="${PROXMOX_CLI_BIN_DIR:-./build/bin}"

if [ -n "${PROXMOX_CLI_BIN}" ] && [ -x "${PROXMOX_CLI_BIN}" ]; then
  printf '%s\n' "${PROXMOX_CLI_BIN}"
  exit 0
fi

if [ -z "${PROXMOX_CLI_GH_REPO}" ] || [ -z "${PROXMOX_CLI_VERSION}" ]; then
  echo "Set PROXMOX_CLI_GH_REPO and PROXMOX_CLI_VERSION, or provide PROXMOX_CLI_BIN" >&2
  exit 1
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

tag="proxmox-cli-v${PROXMOX_CLI_VERSION}"
asset="proxmox-cli_${os}_${arch}"

if [ -n "${PROXMOX_CLI_DOWNLOAD_BASE:-}" ]; then
  base="${PROXMOX_CLI_DOWNLOAD_BASE%/}/${PROXMOX_CLI_GH_REPO}/releases/download/${tag}"
else
  base="https://github.com/${PROXMOX_CLI_GH_REPO}/releases/download/${tag}"
fi

mkdir -p "${PROXMOX_CLI_BIN_DIR}"
curl -fsSL -o "${PROXMOX_CLI_BIN_DIR}/proxmox-cli" "${base}/${asset}"
curl -fsSL -o "${PROXMOX_CLI_BIN_DIR}/checksums.txt" "${base}/checksums.txt"

(cd "${PROXMOX_CLI_BIN_DIR}" && sha256sum --check --ignore-missing checksums.txt)
chmod +x "${PROXMOX_CLI_BIN_DIR}/proxmox-cli"

printf '%s\n' "$(cd "${PROXMOX_CLI_BIN_DIR}" && pwd)/proxmox-cli"
```

## Execution Pattern

After resolving the path:

```bash
"${PROXMOX_CLI_BIN}" --help
"${PROXMOX_CLI_BIN}" --api-base "${PVE_API_BASE_URL%/}/api2/json" --insecure-tls --output json capability list_nodes
```

## Security Notes

1. Always verify checksum before execution.
2. Keep credentials out of command history when possible.
3. Pin version in CI to avoid unreviewed upgrades.
4. Prefer local pinned binary over mutable global PATH in automation.
