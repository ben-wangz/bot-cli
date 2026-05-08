package ttsapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
)

func parseUpstreamError(statusCode int, body []byte) error {
	var root map[string]any
	if err := json.Unmarshal(body, &root); err == nil {
		if payload, ok := root["error"].(map[string]any); ok {
			msg := strings.TrimSpace(asString(payload["message"]))
			if msg == "" {
				msg = strings.TrimSpace(asString(payload["code"]))
			}
			if msg != "" {
				meta := map[string]any{
					"upstream_message": msg,
					"classification":   classifyUpstreamError(statusCode, msg),
					"retryable":        isRetryableUpstream(statusCode, msg),
				}
				return apperr.NewWithMeta(apperr.CodeRPC, msg, meta)
			}
		}
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return apperr.New(apperr.CodeConfig, "authentication failed")
	}
	msg := "upstream request failed"
	meta := map[string]any{
		"upstream_message": msg,
		"classification":   classifyUpstreamError(statusCode, msg),
		"retryable":        isRetryableUpstream(statusCode, msg),
	}
	return apperr.NewWithMeta(apperr.CodeRPC, msg, meta)
}

func classifyUpstreamError(statusCode int, message string) string {
	m := strings.ToLower(strings.TrimSpace(message))
	switch {
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return "auth"
	case statusCode == http.StatusTooManyRequests || strings.Contains(m, "rate limit"):
		return "rate_limit"
	case statusCode >= 500:
		return "transient"
	case statusCode == http.StatusBadRequest || strings.Contains(m, "param") || strings.Contains(m, "invalid"):
		return "invalid_request"
	default:
		return "unknown"
	}
}

func isRetryableUpstream(statusCode int, message string) bool {
	m := strings.ToLower(strings.TrimSpace(message))
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode >= 500 {
		return true
	}
	if strings.Contains(m, "timeout") || strings.Contains(m, "temporar") || strings.Contains(m, "try again") {
		return true
	}
	return false
}

func asStringPath(root map[string]any, path ...string) string {
	current := any(root)
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = m[key]
		if !ok {
			return ""
		}
	}
	return strings.TrimSpace(asString(current))
}

func extractAudioData(root map[string]any) string {
	if v := asStringPath(root, "audio", "data"); v != "" {
		return v
	}
	if output, ok := root["output"].([]any); ok && len(output) > 0 {
		if first, ok := output[0].(map[string]any); ok {
			if v := asStringPath(first, "audio", "data"); v != "" {
				return v
			}
		}
	}
	choices, ok := root["choices"].([]any)
	if !ok || len(choices) == 0 {
		return ""
	}
	first, ok := choices[0].(map[string]any)
	if !ok {
		return ""
	}
	msg, ok := first["message"].(map[string]any)
	if !ok {
		return ""
	}
	audio, ok := msg["audio"].(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(asString(audio["data"]))
}

func extractAudioFormat(root map[string]any) string {
	if v := asStringPath(root, "audio", "format"); v != "" {
		return v
	}
	if output, ok := root["output"].([]any); ok && len(output) > 0 {
		if first, ok := output[0].(map[string]any); ok {
			if v := asStringPath(first, "audio", "format"); v != "" {
				return v
			}
		}
	}
	choices, ok := root["choices"].([]any)
	if !ok || len(choices) == 0 {
		return ""
	}
	first, ok := choices[0].(map[string]any)
	if !ok {
		return ""
	}
	msg, ok := first["message"].(map[string]any)
	if !ok {
		return ""
	}
	audio, ok := msg["audio"].(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(asString(audio["format"]))
}
