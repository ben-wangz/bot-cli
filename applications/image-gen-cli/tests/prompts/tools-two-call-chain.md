# tools method chain test

Purpose:

- Validate tools backend chaining via `previous_response_id`.

Prerequisites:

1. `IMAGE_API_BASE_URL` and `IMAGE_API_KEY` are exported.
2. `jq` is available.
3. Run commands from repository root.

Execution commands:

```bash
mkdir -p build/image-gen-regression/tools

go run ./applications/image-gen-cli/src/cmd/image-gen-cli \
  --api-base-url "$IMAGE_API_BASE_URL" \
  --api-key "$IMAGE_API_KEY" \
  --method tools \
  --timeout 300 \
  --output json \
  capability generate_image \
  --prompt "A cozy cabin in snowy forest" \
  --stream true \
  --store true \
  --output-dir "build/image-gen-regression/tools" \
  --output-name "tools-call-1" | tee build/image-gen-regression/tools/call1.json

jq -e '.ok == true' build/image-gen-regression/tools/call1.json
jq -e '.request.method == "tools"' build/image-gen-regression/tools/call1.json
FIRST_RESPONSE_ID=$(jq -r '.result.response_id' build/image-gen-regression/tools/call1.json)
test -n "$FIRST_RESPONSE_ID"
FIRST_FILE=$(jq -r '.result.output_file' build/image-gen-regression/tools/call1.json)
test -f "$FIRST_FILE"

go run ./applications/image-gen-cli/src/cmd/image-gen-cli \
  --api-base-url "$IMAGE_API_BASE_URL" \
  --api-key "$IMAGE_API_KEY" \
  --method tools \
  --timeout 300 \
  --output json \
  capability generate_image \
  --prompt "Same cabin at night with northern lights" \
  --stream false \
  --store true \
  --previous_response_id "$FIRST_RESPONSE_ID" \
  --output-dir "build/image-gen-regression/tools" \
  --output-name "tools-call-2" | tee build/image-gen-regression/tools/call2.json

jq -e '.ok == true' build/image-gen-regression/tools/call2.json
jq -e '.request.method == "tools"' build/image-gen-regression/tools/call2.json
jq -e '.request.args.previous_response_id == env.FIRST_RESPONSE_ID' build/image-gen-regression/tools/call2.json
SECOND_FILE=$(jq -r '.result.output_file' build/image-gen-regression/tools/call2.json)
test -f "$SECOND_FILE"
```

Pass criteria:

- Both calls return `ok=true`.
- Both responses include `request.method=tools`.
- First call returns non-empty `result.response_id`.
- Second call sends `previous_response_id` and succeeds.
- `result.output_file` from both calls exists on disk.
