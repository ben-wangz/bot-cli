# Troubleshooting

## Timeout errors

Symptoms:

- `context deadline exceeded`
- `Client.Timeout exceeded while awaiting headers`

Actions:

1. Increase `--timeout` (default 300s, try 600s for slow backends).
2. Confirm gateway/network health to `IMAGE_API_BASE_URL`.

## Streaming completed without final image

Symptoms:

- `streaming completed without final image result`

Actions:

1. Retry with non-streaming (`--stream false`) as fallback.
2. Check backend Responses/SSE compliance for `response.completed.response.output[].result`.

## Chaining unsupported or unstable

Symptoms:

- errors mentioning `previous_response_id` / `store` unsupported

Actions:

1. Retry without chaining flags.
2. Confirm backend supports persisted responses with stable response IDs.
