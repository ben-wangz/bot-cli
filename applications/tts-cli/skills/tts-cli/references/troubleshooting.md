# Troubleshooting

## Authentication failed

Symptoms:

- `INVALID_API_KEY`
- `authentication failed`
- HTTP `401` / `403`

Actions:

1. Verify `TTS_API_KEY` is valid and not expired.
2. Confirm `TTS_API_BASE_URL` points to the expected gateway.
3. Retry with a known-good key and masked logging.

## Timeout or network errors

Symptoms:

- `context deadline exceeded`
- `Client.Timeout exceeded while awaiting headers`

Actions:

1. Increase `--timeout` (default 300s, try 600s for slower backends).
2. Check connectivity and TLS/proxy settings for the gateway.

## Streaming warnings for non-audio SSE events

Symptoms:

- warning entries like `ignored one non-audio SSE event`

Actions:

1. If `ok=true` and bytes are positive, treat as non-fatal compatibility noise.
2. If audio is empty, retry with `--stream false` for comparison.
3. Keep `--audio-format pcm16` for streaming compatibility.

## Voice clone input rejected

Symptoms:

- errors related to invalid voice clone payload
- model validation errors on `mimo-v2.5-tts-voiceclone`

Actions:

1. Generate Data URI via `capability file_to_data_uri`.
2. Ensure `--clone-voice-data-uri` starts with `data:{mime};base64,`.
3. Ensure clone sample meets backend MIME/size limits.

## Argument list too long (shell)

Symptoms:

- `/usr/local/bin/go: Argument list too long`

Actions:

1. Use a shorter reference sample audio.
2. Avoid printing or exporting oversized Data URI values unnecessarily.
3. If needed, run in a shell that supports larger argv/env limits.

## Invalid model/arg combination

Symptoms:

- `invalid_request` style errors
- model-specific validation failures

Actions:

1. Check `tts-cli capability describe generate_speech` rules.
2. Use `tts-cli capability suggest_voices` to confirm per-model expectations.
