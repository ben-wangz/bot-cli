---
name: image-gen-cli
description: |
  Generate images with image-gen-cli using direct image generation or Responses
  tools mode with deterministic JSON output and optional chaining in tools mode.
license: MIT
compatibility: opencode
metadata:
  audience: coding-agents
  tool: image-gen-cli
---

# Image Gen CLI Skill

Use this skill when an agent needs reliable, script-friendly text-to-image generation from terminal workflows.

## What image-gen-cli Provides

`image-gen-cli` is capability-first (no workflow layer in v0):

1. `capability generate_image`: generate and save a final image to disk.
2. `capability describe [<name>]`: inspect capability contract and args.

Key behaviors:

- Supports `--method direct` (default) and `--method tools`.
- Supports streaming and non-streaming execution.
- Extracts final image from direct Images API responses or Responses tool output.
- Emits deterministic JSON envelopes for machine parsing.
- Supports `store` + `previous_response_id` chaining only in `--method tools`.

## Operating Principles for Agents

1. Always use `--output json` for deterministic parsing.
2. Set an explicit timeout budget with `--timeout`; for image generation prefer `> 300s` and use `600` when backend latency is unknown.
3. Treat missing final image as failure and save artifacts to a deterministic path.
4. For direct smoke checks, start with `--stream false` before trying streaming.

## Base Command Template

Prefer resolved binary path for repeatable runs:

```bash
IMAGE_GEN_CLI_BIN="${IMAGE_GEN_CLI_BIN:-$(command -v image-gen-cli 2>/dev/null || true)}"

if [ -z "${IMAGE_GEN_CLI_BIN}" ]; then
  echo "image-gen-cli not found. Resolve/download binary first." >&2
  echo "See: references/binary-bootstrap-and-release-download.md" >&2
  exit 1
fi

"${IMAGE_GEN_CLI_BIN}" \
  --api-base-url "$IMAGE_API_BASE_URL" \
  --api-key "$IMAGE_API_KEY" \
  --method direct \
  --timeout 600 \
  --output json \
  capability generate_image \
  --prompt "A cinematic mountain village at sunrise" \
  --stream false \
  --output-dir "./build/regression-image-gen" \
  --output-name "example"
```

Fallback for source-only environments:

```bash
cd "<image-gen-cli-source-root>/src"
go run ./cmd/image-gen-cli ...
```

## Quick Task Routing

- Need binary bootstrap + release download pattern? -> [Binary bootstrap and release download](references/binary-bootstrap-and-release-download.md)
- Need minimal first-run command chain? -> [Quickstart](references/quickstart.md)
- Need capability and argument catalog? -> [Capability catalog](references/capability-catalog.md)
- Need failure handling for timeout/direct-vs-tools/SSE/chaining? -> [Troubleshooting](references/troubleshooting.md)

## Safety Checklist

Before executing:

1. `IMAGE_API_BASE_URL` and `IMAGE_API_KEY` are set and non-empty.
2. Output path is writable and deterministic.
3. Every step checks `ok == true` in JSON.
