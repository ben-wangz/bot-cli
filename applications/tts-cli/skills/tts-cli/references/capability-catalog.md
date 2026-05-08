# Capability Catalog

## Commands

- `capability generate_speech`
- `capability suggest_voices`
- `capability file_to_data_uri`
- `capability describe [<name>]`

## generate_speech arguments

Required:

- `--assistant-text`

Optional:

- `--model` (default `mimo-v2.5-tts`)
- `--user-text`
- `--audio-format` (default `wav`)
- `--builtin-voice` (only for `mimo-v2.5-tts`)
- `--clone-voice-data-uri` (required for `mimo-v2.5-tts-voiceclone`)
- `--stream` (`true`/`false`)
- `--output-dir`
- `--output-name`

Model notes:

- `mimo-v2.5-tts`: built-in voice model; `--builtin-voice` optional.
- `mimo-v2.5-tts-voicedesign`: `--user-text` required; no built-in voice.
- `mimo-v2.5-tts-voiceclone`: `--clone-voice-data-uri` required; no built-in voice.

## suggest_voices arguments

Optional:

- `--model` (if omitted, return suggestions for all supported models)

## file_to_data_uri arguments

Required:

- `--file-path` (local audio file path)

Returns:

- `result.mime_type`
- `result.base64_size_bytes`
- `result.data_uri`

For the complete argument contract, run:

```bash
tts-cli capability describe generate_speech
tts-cli capability describe suggest_voices
tts-cli capability describe file_to_data_uri
```
