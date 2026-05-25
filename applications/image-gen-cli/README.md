# image-gen-cli

Agent-first CLI for text-to-image generation via IMAGE2 / `gpt-image-2` compatible Responses API.

## Versioning

Version is managed via the root `version-control.yaml` `forgekit` app mapping.

- Binary name: `image-gen-cli`
- Version file: `applications/image-gen-cli/VERSION`
- Mapping file: `version-control.yaml`

Common commands:

- `FORGEKIT_BIN=$(bash "$PROJECT_ROOT/setup/forgekit.sh")`
- `${FORGEKIT_BIN} --project-root "$PROJECT_ROOT" version get image-gen-cli`
- `${FORGEKIT_BIN} --project-root "$PROJECT_ROOT" version bump binary image-gen-cli patch`

## Scope (v0)

- capability-first command surface (`capability ...`)
- no workflow layer in v0; keep capability-only surface
- deterministic JSON output for agent/script consumption
- single backend implementation (no multi-provider abstraction)

## Quick Start

1. Build binary:

```bash
cd applications/image-gen-cli/src
go build ./cmd/image-gen-cli
```

2. Set API environment:

```bash
export IMAGE_API_BASE_URL="https://<your-domain>"
export IMAGE_API_KEY="<your-key>"
```

3. Run a capability call (streaming by default):

```bash
./image-gen-cli --method direct capability generate_image --prompt "A cinematic mountain village at sunrise"
```

## Core Commands

- `capability generate_image`
- `capability describe [<name>]`

## Global Options

- `--api-base-url <url>` (or env `IMAGE_API_BASE_URL`)
- `--api-key <token>` (or env `IMAGE_API_KEY`)
- `--method <direct|tools>` (or env `IMAGE_METHOD`, default: `direct`)
- `--timeout <seconds>` (default: 300)
- `--output json`

Notes:

- `--previous-response-id`, `--store`, and `--model` require `--method tools`.
- `diagnostics.preview_count` is deprecated and kept for backward compatibility.

## Output Contract

Default output is JSON envelope:

```json
{
  "ok": true,
  "request": {},
  "result": {},
  "diagnostics": {}
}
```

## Prompt Regressions

- `applications/image-gen-cli/tests/prompts/direct-two-call-smoke.md`
This prompt covers the agreed minimal validation path:

1. first call uses non-streaming mode and saves final image
2. second call uses streaming mode and saves final image

## OpenCode Skill Install

OpenCode docs confirm skill discovery locations include:

- Project-level (default): `$PROJECT_ROOT/.opencode/skills/<name>/SKILL.md`
- User-level (global): `~/.config/opencode/skills/<name>/SKILL.md`

Reference: `https://opencode.ai/docs/skills` (`Place files` section).

Install `image-gen-cli` skill to project-level path via release asset (`.tar.gz`):

```bash
PROJECT_ROOT="/path/to/your/project"
GH_REPO="ben-wangz/bot-cli"
VERSION="0.1.0"
TAG="image-gen-cli-v${VERSION}"
ASSET="image-gen-cli_skills_${VERSION}.tar.gz"

mkdir -p "$PROJECT_ROOT/build" "$PROJECT_ROOT/.opencode/skills"
curl -fsSL -o "$PROJECT_ROOT/build/image-gen-cli-skills.tar.gz" \
  "https://github.com/${GH_REPO}/releases/download/${TAG}/${ASSET}"
rm -rf "$PROJECT_ROOT/.opencode/skills/image-gen-cli"
tar -xzf "$PROJECT_ROOT/build/image-gen-cli-skills.tar.gz" -C "$PROJECT_ROOT/.opencode/skills"
```

Install to user-level path (shared across projects):

```bash
mkdir -p "$HOME/.config/opencode/skills"
rm -rf "$HOME/.config/opencode/skills/image-gen-cli"
cp -R "$PROJECT_ROOT/.opencode/skills/image-gen-cli" "$HOME/.config/opencode/skills/image-gen-cli"
```

## Release

- Tag format: `image-gen-cli-v<semver>`
- Release notes directory: `applications/image-gen-cli/release/`
- Release assets should include platform binaries, `checksums.txt`, and skills bundle `image-gen-cli_skills_<version>.tar.gz`
