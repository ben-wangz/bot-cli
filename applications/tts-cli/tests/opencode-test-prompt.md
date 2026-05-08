You are testing `applications/tts-cli` in this repository.

## Goal
Validate the current capabilities end-to-end:
1. `capability list`
2. `capability describe generate_speech`
3. `capability suggest_voices`
4. `capability file_to_data_uri`
5. `capability generate_speech` with:
   - built-in voice model
   - voicedesign model
   - voiceclone model (using Data URI converted from local Chinese sample)

## Constraints
- Do not modify source code.
- Do not create temporary test code files.
- Use only CLI commands and report outputs.
- Keep secrets out of logs where possible (mask API key in final report).

## Test Inputs
- Working directory: `applications/tts-cli/src`
- Local sample audio (copyright-safe Chinese speech):
  `../tests/artofwar_01_sun_64kb_trimmed_32k.mp3`
- API base URL: `https://token-plan-cn.xiaomimimo.com/v1`
- API key: `<REPLACE_WITH_VALID_TP_KEY>`

## Steps

### 1) Build check
Run:
`go build ./...`

### 2) Basic discovery checks
Run:
- `go run ./cmd/tts-cli --help`
- `go run ./cmd/tts-cli capability list`
- `go run ./cmd/tts-cli capability describe generate_speech`
- `go run ./cmd/tts-cli capability describe file_to_data_uri`
- `go run ./cmd/tts-cli capability suggest_voices`

### 3) Convert local file to Data URI
Run:
`go run ./cmd/tts-cli capability file_to_data_uri --file-path ../tests/artofwar_01_sun_64kb_trimmed_32k.mp3`

From JSON output, capture:
- `result.mime_type`
- `result.base64_size_bytes`
- `result.data_uri`

Store `result.data_uri` in a shell variable named `DATA_URI` for the next step.

### 4) End-to-end synthesis tests
Use global flags:
- `--api-base-url https://token-plan-cn.xiaomimimo.com/v1`
- `--api-key <REPLACE_WITH_VALID_TP_KEY>`
- `--output-dir build/tts-phase2`

#### 4.1 Built-in voice test
Run `generate_speech` with:
- model: `mimo-v2.5-tts`
- assistant text: short Chinese sentence
- user text: short style instruction
- builtin voice: `Chloe`
- audio format: `wav`

#### 4.2 Voice design test
Run `generate_speech` with:
- model: `mimo-v2.5-tts-voicedesign`
- assistant text: short Chinese sentence
- user text: voice identity/style description
- audio format: `wav`

#### 4.3 Voice clone test
Run `generate_speech` with:
- model: `mimo-v2.5-tts-voiceclone`
- assistant text: short Chinese sentence
- user text: short style instruction
- clone voice data uri: `$DATA_URI`
- audio format: `wav`

### 5) Streaming compatibility test
Run `generate_speech` with:
- model: `mimo-v2.5-tts`
- assistant text: short Chinese sentence
- stream: `true`
- audio format: `pcm16`

## Expected Pass Criteria
- All commands exit successfully.
- `file_to_data_uri` returns non-empty `data_uri` and a valid audio MIME.
- Each synthesis command returns JSON with:
  - `ok: true`
  - `result.output_file` exists
  - non-empty `result.response_id`
  - positive `result.bytes`
- Streaming test returns `stream: true` and `chunk_count >= 1`.

## Final Report Format
Return a concise markdown report with:
1. Environment (cwd, go version)
2. Commands run (summarized)
3. Results per test case (pass/fail)
4. Key JSON fields (response_id, bytes, output_file, elapsed_ms)
5. Any warnings/errors
6. Final verdict: PASS or FAIL
