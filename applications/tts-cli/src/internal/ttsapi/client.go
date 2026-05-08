package ttsapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type GenerateParams struct {
	Model             string
	AssistantText     string
	UserText          string
	AudioFormat       string
	BuiltinVoice      string
	CloneVoiceDataURI string
	Stream            bool
}

type GenerateResult struct {
	ResponseID     string
	AudioBytes     []byte
	AudioFormat    string
	AudioFormatRaw string
	Stream         bool
	ChunkCount     int
	AudioByteSize  int
	Warnings       []string
}

func New(baseURL string, apiKey string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, apperr.New(apperr.CodeConfig, "api base url is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, apperr.New(apperr.CodeConfig, "api key is required")
	}
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	return &Client{baseURL: baseURL, apiKey: strings.TrimSpace(apiKey), http: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) Generate(ctx context.Context, p GenerateParams) (GenerateResult, error) {
	payload := map[string]any{
		"model":    p.Model,
		"messages": buildMessages(p.UserText, p.AssistantText),
		"audio": map[string]any{
			"format": p.AudioFormat,
		},
		"stream": p.Stream,
	}
	voice := strings.TrimSpace(p.BuiltinVoice)
	if voice == "" {
		voice = strings.TrimSpace(p.CloneVoiceDataURI)
	}
	if voice != "" {
		audio := payload["audio"].(map[string]any)
		audio["voice"] = voice
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeInternal, "failed to marshal request", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeInternal, "failed to build request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeNetwork, "request failed", err)
	}
	defer resp.Body.Close()
	if p.Stream {
		return decodeStreaming(resp)
	}
	return decodeNonStreaming(resp)
}

func buildMessages(userText string, assistantText string) []map[string]string {
	messages := make([]map[string]string, 0, 2)
	if strings.TrimSpace(userText) != "" {
		messages = append(messages, map[string]string{"role": "user", "content": strings.TrimSpace(userText)})
	}
	messages = append(messages, map[string]string{"role": "assistant", "content": strings.TrimSpace(assistantText)})
	return messages
}

func decodeNonStreaming(resp *http.Response) (GenerateResult, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeNetwork, "failed to read response body", err)
	}
	if resp.StatusCode >= 400 {
		return GenerateResult{}, parseUpstreamError(resp.StatusCode, body)
	}
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeParse, "failed to parse response JSON", err)
	}
	audioData := extractAudioData(root)
	if strings.TrimSpace(audioData) == "" {
		return GenerateResult{}, apperr.New(apperr.CodeRPC, "upstream response missing audio data")
	}
	decoded, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeParse, "failed to decode audio base64", err)
	}
	rawFormat := extractAudioFormat(root)
	return GenerateResult{
		ResponseID:     asString(root["id"]),
		AudioBytes:     decoded,
		AudioFormat:    rawFormat,
		AudioFormatRaw: rawFormat,
		Stream:         false,
		ChunkCount:     0,
		AudioByteSize:  len(decoded),
		Warnings:       []string{},
	}, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
