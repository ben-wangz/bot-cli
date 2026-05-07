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
  --timeout 300 \
  --output json \
  capability generate_image \
  --prompt "A cozy cabin in snowy forest" \
  --stream true \
  --output-dir "./build/regression-image-gen" \
  --output-name "quickstart"
```

Expected success shape:

- Check `ok=true`, `result.output_file`, and `result.response_id`.
