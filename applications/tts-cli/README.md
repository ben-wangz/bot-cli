# tts-cli

Agent-first CLI for speech synthesis via MiMo-compatible `/v1/chat/completions` APIs.

## Versioning

Version is managed via the root `version-control.yaml` `forgekit` app mapping.

- Binary name: `tts-cli`
- Version file: `applications/tts-cli/VERSION`
- Mapping file: `version-control.yaml`

Common commands:

- `FORGEKIT_BIN=$(bash "$PROJECT_ROOT/setup/forgekit.sh")`
- `${FORGEKIT_BIN} --project-root "$PROJECT_ROOT" version get tts-cli`
- `${FORGEKIT_BIN} --project-root "$PROJECT_ROOT" version bump binary tts-cli patch`

## Scope

- capability-first command surface (`capability ...`)
- deterministic JSON output for agent/script consumption
- supports built-in voice, voice design, voice clone, and streaming compatibility mode

## Quick Start

1. Build binary:

```bash
cd applications/tts-cli/src
go build ./cmd/tts-cli
```

2. Set API environment:

```bash
export TTS_API_BASE_URL="https://token-plan-cn.xiaomimimo.com/v1"
export TTS_API_KEY="<your-key>"
```

3. Run a capability call:

```bash
./tts-cli \
  --api-base-url "$TTS_API_BASE_URL" \
  --api-key "$TTS_API_KEY" \
  --output-dir build/tts \
  capability generate_speech \
  --model mimo-v2.5-tts \
  --assistant-text "这是一次快速验证。" \
  --user-text "语速自然，语气平稳。" \
  --builtin-voice Chloe \
  --audio-format wav
```

## Core Commands

- `capability generate_speech`
- `capability suggest_voices`
- `capability file_to_data_uri`
- `capability list`
- `capability describe [<name>]`

## Global Options

- `--api-base-url <url>` (or env `TTS_API_BASE_URL`)
- `--api-key <token>` (or env `TTS_API_KEY`)
- `--timeout <seconds>` (default: 300)
- `--output-dir <path>`
- `--output-name <name>`
- `--output json`

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

## Regression Prompt

- `applications/tts-cli/tests/opencode-test-prompt.md`

This prompt covers build check, capability discovery, Data URI conversion, and end-to-end synthesis across built-in, voicedesign, voiceclone, and streaming modes.

## OpenCode Skill Install

OpenCode docs confirm skill discovery locations include:

- Project-level (default): `$PROJECT_ROOT/.opencode/skills/<name>/SKILL.md`
- User-level (global): `~/.config/opencode/skills/<name>/SKILL.md`

Reference: `https://opencode.ai/docs/skills` (`Place files` section).

Install `tts-cli` skill to project-level path via release asset (`.tar.gz`):

```bash
PROJECT_ROOT="/path/to/your/project"
GH_REPO="ben-wangz/bot-cli"
VERSION="0.1.0"
TAG="tts-cli-v${VERSION}"
ASSET="tts-cli_skills_${VERSION}.tar.gz"

mkdir -p "$PROJECT_ROOT/build" "$PROJECT_ROOT/.opencode/skills"
curl -fsSL -o "$PROJECT_ROOT/build/tts-cli-skills.tar.gz" \
  "https://github.com/${GH_REPO}/releases/download/${TAG}/${ASSET}"
rm -rf "$PROJECT_ROOT/.opencode/skills/tts-cli"
tar -xzf "$PROJECT_ROOT/build/tts-cli-skills.tar.gz" -C "$PROJECT_ROOT/.opencode/skills"
```

Install to user-level path (shared across projects):

```bash
mkdir -p "$HOME/.config/opencode/skills"
rm -rf "$HOME/.config/opencode/skills/tts-cli"
cp -R "$PROJECT_ROOT/.opencode/skills/tts-cli" "$HOME/.config/opencode/skills/tts-cli"
```

## Release

- Tag format: `tts-cli-v<semver>`
- Release assets should include platform binaries, `checksums.txt`, and skills bundle `tts-cli_skills_<version>.tar.gz`
