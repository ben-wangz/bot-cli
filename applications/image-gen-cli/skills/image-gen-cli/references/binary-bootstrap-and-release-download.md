# Binary Bootstrap and Release Download

Use this pattern when the environment does not already provide `image-gen-cli`.

## Resolve local binary first

```bash
IMAGE_GEN_CLI_BIN="${IMAGE_GEN_CLI_BIN:-$(command -v image-gen-cli 2>/dev/null || true)}"
```

If found, reuse it for all calls in one run.

## Download from release when missing

```bash
GH_REPO="ben-wangz/bot-cli"
VERSION="<target-version>"
TAG="image-gen-cli-v${VERSION}"
OS="linux"
ARCH="amd64"
ASSET="image-gen-cli_${OS}_${ARCH}"

mkdir -p ./build/bin
curl -fsSL -o ./build/bin/image-gen-cli "https://github.com/${GH_REPO}/releases/download/${TAG}/${ASSET}"
chmod +x ./build/bin/image-gen-cli
IMAGE_GEN_CLI_BIN="$(pwd)/build/bin/image-gen-cli"
```

Note: the CLI binary asset is a raw executable, not a `.tar.gz` archive. The `.tar.gz` asset in the release is the skills bundle.

Optional integrity check:

1. Download `checksums.txt` from the same release.
2. Verify `sha256sum` for the selected asset before execution.
