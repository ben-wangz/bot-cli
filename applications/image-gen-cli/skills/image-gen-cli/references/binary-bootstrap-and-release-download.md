# Binary Bootstrap and Release Download

Use this pattern when the environment does not already provide `image-gen-cli`.

## Goal

Guarantee a usable `image-gen-cli` binary before running image generation calls.

## Recommended Resolution Order

1. Use `IMAGE_GEN_CLI_BIN` if explicitly provided.
2. Else try `command -v image-gen-cli`.
3. Else download from GitHub Releases.
4. Verify checksum when `checksums.txt` is available.
5. Reuse the resolved path for all commands in the run.

## Required Inputs

- `IMAGE_GEN_CLI_GH_REPO` in `owner/repo` format (example: `ben-wangz/bot-cli`)
- `IMAGE_GEN_CLI_VERSION` semver without `v` (example: `0.1.2`)

Optional:

- `IMAGE_GEN_CLI_BIN_DIR` (default: `./build/bin`)
- `IMAGE_GEN_CLI_DOWNLOAD_BASE` for mirror/proxy base URL

Guardrails:

1. Prefer a path already covered by repository `.gitignore` (recommended: `./build/bin` in this repo).
2. If using a custom path, ensure it is ignored before download.

## Release Asset Convention

Current `image-gen-cli` releases publish:

- tag: `image-gen-cli-v<semver>`
- binary assets:
  - `image-gen-cli_linux_amd64`
  - `image-gen-cli_linux_arm64`
  - `image-gen-cli_darwin_amd64`
  - `image-gen-cli_darwin_arm64`
- optional checksum file:
  - `checksums.txt`
- skills bundle archive:
  - `image-gen-cli_skills_<semver>.tar.gz`

Important:

1. The CLI binary assets are raw executables, not `.tar.gz` archives.
2. The skills bundle is a `.tar.gz` archive and should not be confused with the CLI binary asset.
3. If a future release adds both raw binaries and archives, prefer the raw binary asset that matches `os` and `arch` unless your environment explicitly requires an archive workflow.

## Optional: Download Skill Bundle

If you also need the OpenCode skill package from the same release:

```bash
tag="image-gen-cli-v${IMAGE_GEN_CLI_VERSION}"
skill_asset="image-gen-cli_skills_${IMAGE_GEN_CLI_VERSION}.tar.gz"
base="https://github.com/${IMAGE_GEN_CLI_GH_REPO}/releases/download/${tag}"

mkdir -p ./build ./.opencode/skills
curl -fsSL -o ./build/image-gen-cli-skills.tar.gz "${base}/${skill_asset}"
tar -xzf ./build/image-gen-cli-skills.tar.gz -C ./.opencode/skills
```

The archive contains `skills/image-gen-cli/...`, so extraction creates `./.opencode/skills/image-gen-cli`.

## Portable Bootstrap Snippet

This snippet resolves `os`/`arch`, downloads the matching raw binary asset, verifies checksum when available, and writes the final executable to `build/bin/image-gen-cli`.

```bash
#!/usr/bin/env bash
set -euo pipefail

IMAGE_GEN_CLI_BIN="${IMAGE_GEN_CLI_BIN:-$(command -v image-gen-cli 2>/dev/null || true)}"
IMAGE_GEN_CLI_GH_REPO="${IMAGE_GEN_CLI_GH_REPO:-ben-wangz/bot-cli}"
IMAGE_GEN_CLI_VERSION="${IMAGE_GEN_CLI_VERSION:-}"
IMAGE_GEN_CLI_BIN_DIR="${IMAGE_GEN_CLI_BIN_DIR:-./build/bin}"

if [ -n "${IMAGE_GEN_CLI_BIN}" ] && [ -x "${IMAGE_GEN_CLI_BIN}" ]; then
  printf '%s\n' "${IMAGE_GEN_CLI_BIN}"
  exit 0
fi

if [ -z "${IMAGE_GEN_CLI_VERSION}" ]; then
  echo "Set IMAGE_GEN_CLI_VERSION, or provide IMAGE_GEN_CLI_BIN" >&2
  exit 1
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

tag="image-gen-cli-v${IMAGE_GEN_CLI_VERSION}"
asset="image-gen-cli_${os}_${arch}"

if [ -n "${IMAGE_GEN_CLI_DOWNLOAD_BASE:-}" ]; then
  base="${IMAGE_GEN_CLI_DOWNLOAD_BASE%/}/${IMAGE_GEN_CLI_GH_REPO}/releases/download/${tag}"
else
  base="https://github.com/${IMAGE_GEN_CLI_GH_REPO}/releases/download/${tag}"
fi

mkdir -p "${IMAGE_GEN_CLI_BIN_DIR}"
curl -fsSL -o "${IMAGE_GEN_CLI_BIN_DIR}/${asset}" "${base}/${asset}"

if curl -fsSL -o "${IMAGE_GEN_CLI_BIN_DIR}/checksums.txt" "${base}/checksums.txt"; then
  (cd "${IMAGE_GEN_CLI_BIN_DIR}" && sha256sum --check --ignore-missing checksums.txt)
fi

mv -f "${IMAGE_GEN_CLI_BIN_DIR}/${asset}" "${IMAGE_GEN_CLI_BIN_DIR}/image-gen-cli"
chmod +x "${IMAGE_GEN_CLI_BIN_DIR}/image-gen-cli"

printf '%s\n' "$(cd "${IMAGE_GEN_CLI_BIN_DIR}" && pwd)/image-gen-cli"
```

## Execution Pattern

After resolving the path:

```bash
"${IMAGE_GEN_CLI_BIN}" --help
"${IMAGE_GEN_CLI_BIN}" --output json capability list
```

## Security Notes

1. Verify checksum before execution when `checksums.txt` is published.
2. Pin version in CI to avoid unreviewed upgrades.
3. Prefer local pinned binary over mutable global PATH in automation.
