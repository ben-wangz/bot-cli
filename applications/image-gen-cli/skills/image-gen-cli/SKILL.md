---
name: image-gen-cli
description: |
  Generate images with image-gen-cli using IMAGE2 / Responses API with reliable
  final-image extraction, deterministic JSON output, and optional response chaining.
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

- Supports streaming and non-streaming execution.
- Extracts final image from `output[].result` semantics.
- Emits deterministic JSON envelopes for machine parsing.
- Supports `store` + `previous_response_id` chaining when backend supports it.

## Operating Principles for Agents

1. Always use `--output json` for deterministic parsing.
2. Set an explicit timeout budget with `--timeout` (default 300s).
3. Treat missing final image as failure and save artifacts to a deterministic path.

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
  --timeout 300 \
  --output json \
  capability generate_image \
  --prompt "A cinematic mountain village at sunrise" \
  --stream true \
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
- Need failure handling for timeout/SSE/chaining? -> [Troubleshooting](references/troubleshooting.md)

## Safety Checklist

Before executing:

1. `IMAGE_API_BASE_URL` and `IMAGE_API_KEY` are set and non-empty.
2. Output path is writable and deterministic.
3. Every step checks `ok == true` in JSON.
