# Quickstart

## Prerequisites

1. `tts-cli` binary available in `PATH`, or run via source with `go run`.
2. Environment variables set:
   - `TTS_API_BASE_URL`
   - `TTS_API_KEY`

## Minimal Run

```bash
tts-cli \
  --api-base-url "$TTS_API_BASE_URL" \
  --api-key "$TTS_API_KEY" \
  --timeout 300 \
  --output json \
  --output-dir "./build/regression-tts" \
  capability generate_speech \
  --model mimo-v2.5-tts \
  --assistant-text "这是一次 quickstart 回归测试。" \
  --user-text "语速自然，表达清晰。" \
  --builtin-voice Chloe \
  --audio-format wav
```

Expected success shape:

- Check `ok=true`, `result.output_file`, `result.response_id`, and `result.bytes > 0`.
