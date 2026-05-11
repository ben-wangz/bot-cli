# direct two-call smoke

Purpose:

- Validate default direct backend behavior with non-stream first, then stream.

Prerequisites:

1. `IMAGE_API_BASE_URL` and `IMAGE_API_KEY` are exported.
2. `jq` is available.
3. Run commands from repository root.

Execution commands:

```bash
set -euo pipefail

mkdir -p build/image-gen-regression/direct

go run ./applications/image-gen-cli/src/cmd/image-gen-cli \
  --api-base-url "$IMAGE_API_BASE_URL" \
  --api-key "$IMAGE_API_KEY" \
  --method direct \
  --timeout 300 \
  --output json \
  capability generate_image \
  --prompt "A cozy cabin in snowy forest" \
  --stream false \
  --output-dir "build/image-gen-regression/direct" \
  --output-name "direct-call-1" | tee build/image-gen-regression/direct/call1.json

jq -e '.ok == true' build/image-gen-regression/direct/call1.json
jq -e '.request.method == "direct"' build/image-gen-regression/direct/call1.json
FIRST_FILE=$(jq -r '.result.output_file' build/image-gen-regression/direct/call1.json)
test -f "$FIRST_FILE"

go run ./applications/image-gen-cli/src/cmd/image-gen-cli \
  --api-base-url "$IMAGE_API_BASE_URL" \
  --api-key "$IMAGE_API_KEY" \
  --method direct \
  --timeout 300 \
  --output json \
  capability generate_image \
  --prompt "Same cabin at night with northern lights" \
  --stream true \
  --output-dir "build/image-gen-regression/direct" \
  --output-name "direct-call-2" | tee build/image-gen-regression/direct/call2.json

jq -e '.ok == true' build/image-gen-regression/direct/call2.json
jq -e '.request.method == "direct"' build/image-gen-regression/direct/call2.json
SECOND_FILE=$(jq -r '.result.output_file' build/image-gen-regression/direct/call2.json)
test -f "$SECOND_FILE"
```

Pass criteria:

- Both calls return `ok=true`.
- First non-stream call succeeds before stream call starts.
- Both responses include `request.method=direct`.
- `result.output_file` from both calls exists on disk.
