# image-gen-cli

Agent-first CLI for text-to-image generation via IMAGE2 / `gpt-image-2` compatible Responses API.

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
./image-gen-cli capability generate_image --prompt "A cinematic mountain village at sunrise"
```

## Core Commands

- `capability generate_image`
- `capability describe [<name>]`

## Global Options

- `--api-base-url <url>` (or env `IMAGE_API_BASE_URL`)
- `--api-key <token>` (or env `IMAGE_API_KEY`)
- `--timeout <seconds>`
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

## Prompt Regressions

- `applications/image-gen-cli/tests/prompts/minimal-two-call-chain.md`
This prompt covers the agreed minimal validation path:

1. first call uses streaming mode and saves final image
2. second call uses non-streaming mode with `previous_response_id` and saves final image
