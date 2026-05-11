# Troubleshooting

## Timeout errors

Symptoms:

- `context deadline exceeded`
- `Client.Timeout exceeded while awaiting headers`

Actions:

1. Increase `--timeout` beyond `300`; start with `600s` for image generation when backend latency is unknown.
2. Confirm gateway/network health to `IMAGE_API_BASE_URL`.
3. For direct mode, retry `--stream false` first; some providers deliver non-stream reliably before streaming stabilizes.

## Direct vs tools compatibility

Symptoms:

- `previous_response_id requires --method tools`
- `store requires --method tools`
- `model requires --method tools`

Actions:

1. Use `--method direct` for basic image generation.
2. Switch to `--method tools` only when you need chaining or Responses-model selection.

## Streaming completed without final image

Symptoms:

- `streaming completed without final image result`

Actions:

1. Retry with non-streaming (`--stream false`) as fallback.
2. For `--method tools`, check backend Responses/SSE compliance for `response.completed.response.output[].result`.
3. For `--method direct`, inspect whether the provider emits `image_generation.completed` with `b64_json`.

## Provider claims model support but generation still fails

Symptoms:

- `/v1/models` lists `gpt-image-2`
- generation returns `503`, `No available compatible accounts`, `permission_error`, or hangs for a long time

Actions:

1. Verify `/v1/images/generations` with a minimal `curl` request.
2. If listed in `/v1/models` but generation fails, treat it as provider-side availability rather than a local CLI issue.
3. Keep timeout above `300s` before concluding the route is unavailable on slow providers.

## Chaining unsupported or unstable

Symptoms:

- errors mentioning `previous_response_id` / `store` unsupported

Actions:

1. Retry without chaining flags.
2. Confirm backend supports persisted responses with stable response IDs.
