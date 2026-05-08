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
ASSET="tts-cli_${OS}_${ARCH}.tar.gz"

mkdir -p ./build/bin
curl -fsSL -o ./build/bin/tts-cli.tgz "https://github.com/${GH_REPO}/releases/download/${TAG}/${ASSET}"
tar -xzf ./build/bin/tts-cli.tgz -C ./build/bin
chmod +x ./build/bin/tts-cli
TTS_CLI_BIN="$(pwd)/build/bin/tts-cli"
```

Optional integrity check:

1. Download `checksums.txt` from the same release.
2. Verify `sha256sum` for the selected asset before extraction.
