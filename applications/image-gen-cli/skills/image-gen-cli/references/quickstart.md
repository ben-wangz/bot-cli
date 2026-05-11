# Quickstart

## Prerequisites

1. `image-gen-cli` binary available in `PATH`, or run via source with `go run`.
2. Environment variables set:
   - `IMAGE_API_BASE_URL`
   - `IMAGE_API_KEY`

## Minimal Run

```bash
image-gen-cli \
  --api-base-url "$IMAGE_API_BASE_URL" \
  --api-key "$IMAGE_API_KEY" \
  --method direct \
  --timeout 600 \
  --output json \
  capability generate_image \
  --prompt "A cozy cabin in snowy forest" \
  --stream false \
  --output-dir "./build/regression-image-gen" \
  --output-name "quickstart"
```

Recommended progression:

1. Start with `--method direct --stream false`.
2. Only try `--stream true` after non-stream succeeds.
3. Use `--method tools` only when you need `--store` or `--previous-response-id`.

Expected success shape:

- Check `ok=true` and `result.output_file`.
- `result.response_id` may be empty in direct mode.
