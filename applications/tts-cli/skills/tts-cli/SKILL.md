---
name: tts-cli
description: |
  Generate speech with tts-cli using MiMo TTS models (built-in voice, voice design,
  voice clone, and streaming) with deterministic JSON outputs for agent workflows.
license: MIT
compatibility: opencode
metadata:
  audience: coding-agents
  tool: tts-cli
---

# TTS CLI Skill

Use this skill when an agent needs reliable, script-friendly speech synthesis from terminal workflows.

## What tts-cli Provides

`tts-cli` is capability-first:

1. `capability generate_speech`: synthesize speech and save audio to disk.
2. `capability suggest_voices`: inspect model/voice guidance and compatibility rules.
3. `capability file_to_data_uri`: convert local audio files to RFC 2397 Data URI for voice clone input.
4. `capability describe [<name>]`: inspect capability contract and args.

Key behaviors:

- Supports non-streaming and streaming (`stream=true`) execution.
- Emits deterministic JSON envelopes for machine parsing.
- Supports MiMo model families: built-in voice, voice design, and voice clone.

## Operating Principles for Agents

1. Always use `--output json` for deterministic parsing.
2. Set an explicit timeout budget with `--timeout` (default 300s).
3. For voice clone, use `file_to_data_uri` first and pass `--clone-voice-data-uri`.
4. In streaming mode, prefer `--audio-format pcm16` for chunk stitching compatibility.

## Base Command Template

Prefer resolved binary path for repeatable runs:

```bash
TTS_CLI_BIN="${TTS_CLI_BIN:-$(command -v tts-cli 2>/dev/null || true)}"

if [ -z "${TTS_CLI_BIN}" ]; then
  echo "tts-cli not found. Resolve/download binary first." >&2
  echo "See: references/binary-bootstrap-and-release-download.md" >&2
  exit 1
fi

"${TTS_CLI_BIN}" \
  --api-base-url "$TTS_API_BASE_URL" \
  --api-key "$TTS_API_KEY" \
  --timeout 300 \
  --output json \
  --output-dir "./build/regression-tts" \
  capability generate_speech \
  --model mimo-v2.5-tts \
  --assistant-text "这是一次技能快速验证。" \
  --user-text "语速自然，语气平稳。" \
  --builtin-voice Chloe \
  --audio-format wav
```

Fallback for source-only environments:

```bash
cd "<tts-cli-source-root>/src"
go run ./cmd/tts-cli ...
```

## Quick Task Routing

- Need binary bootstrap + release download pattern? -> [Binary bootstrap and release download](references/binary-bootstrap-and-release-download.md)
- Need minimal first-run command chain? -> [Quickstart](references/quickstart.md)
- Need capability and argument catalog? -> [Capability catalog](references/capability-catalog.md)
- Need failure handling for auth/SSE/data-uri/argument issues? -> [Troubleshooting](references/troubleshooting.md)

## Safety Checklist

Before executing:

1. `TTS_API_BASE_URL` and `TTS_API_KEY` are set and non-empty.
2. Output path is writable and deterministic.
3. Every step checks `ok == true` in JSON.
4. Do not print full API keys or huge Data URI values in logs.
