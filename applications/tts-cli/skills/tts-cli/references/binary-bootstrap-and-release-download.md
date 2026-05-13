# Binary Bootstrap and Release Download

Use this pattern when the environment does not already provide `tts-cli`.

## Resolve local binary first

```bash
TTS_CLI_BIN="${TTS_CLI_BIN:-$(command -v tts-cli 2>/dev/null || true)}"
```

If found, reuse it for all calls in one run.

## Download from release when missing

```bash
GH_REPO="ben-wangz/bot-cli"
VERSION="<target-version>"
TAG="tts-cli-v${VERSION}"
OS="linux"
ARCH="amd64"
ASSET="tts-cli_${OS}_${ARCH}"

mkdir -p ./build/bin
curl -fsSL -o ./build/bin/tts-cli "https://github.com/${GH_REPO}/releases/download/${TAG}/${ASSET}"
chmod +x ./build/bin/tts-cli
TTS_CLI_BIN="$(pwd)/build/bin/tts-cli"
```

Note: the CLI binary asset is a raw executable, not a `.tar.gz` archive. The `.tar.gz` asset in the release is the skills bundle.

Optional integrity check:

1. Download `checksums.txt` from the same release.
2. Verify `sha256sum` for the selected asset before execution.

Example:

```bash
curl -fsSL -o ./build/bin/checksums.txt "https://github.com/${GH_REPO}/releases/download/${TAG}/checksums.txt"
(cd ./build/bin && sha256sum --check --ignore-missing checksums.txt)
```
