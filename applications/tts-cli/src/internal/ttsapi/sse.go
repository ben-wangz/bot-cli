package ttsapi

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
)

func decodeStreaming(resp *http.Response) (GenerateResult, error) {
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return GenerateResult{}, parseUpstreamError(resp.StatusCode, body)
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var responseID string
	var chunkCount int
	var warnings []string
	buf := bytes.NewBuffer(nil)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var root map[string]any
		if err := json.Unmarshal([]byte(payload), &root); err != nil {
			warnings = append(warnings, "ignored one malformed SSE event")
			continue
		}
		if responseID == "" {
			responseID = asString(root["id"])
		}
		if err := appendChunkAudio(buf, root); err == nil {
			chunkCount++
		}
	}
	if err := scanner.Err(); err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeNetwork, "failed to read streaming response", err)
	}
	if buf.Len() == 0 {
		return GenerateResult{}, apperr.New(apperr.CodeRPC, "streaming completed without audio chunks")
	}
	return GenerateResult{ResponseID: responseID, AudioBytes: buf.Bytes(), Stream: true, ChunkCount: chunkCount, AudioByteSize: buf.Len(), Warnings: warnings}, nil
}

func appendChunkAudio(dst *bytes.Buffer, root map[string]any) error {
	choices, ok := root["choices"].([]any)
	if !ok || len(choices) == 0 {
		return apperr.New(apperr.CodeParse, "missing choices in stream chunk")
	}
	first, ok := choices[0].(map[string]any)
	if !ok {
		return apperr.New(apperr.CodeParse, "invalid stream choice payload")
	}
	delta, ok := first["delta"].(map[string]any)
	if !ok {
		return apperr.New(apperr.CodeParse, "missing delta payload")
	}
	audio, ok := delta["audio"].(map[string]any)
	if !ok {
		return apperr.New(apperr.CodeParse, "missing delta audio payload")
	}
	data := strings.TrimSpace(asString(audio["data"]))
	if data == "" {
		return apperr.New(apperr.CodeParse, "empty audio chunk")
	}
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return apperr.Wrap(apperr.CodeParse, "invalid base64 audio chunk", err)
	}
	_, _ = dst.Write(b)
	return nil
}
