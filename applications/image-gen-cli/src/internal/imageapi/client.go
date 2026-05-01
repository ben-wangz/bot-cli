package imageapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type GenerateParams struct {
	Prompt             string
	Model              string
	ImageModel         string
	Stream             bool
	Store              bool
	PreviousResponseID string
	Size               string
	Quality            string
	OutputFormat       string
	OutputCompression  int
	Background         string
}

type GenerateResult struct {
	ResponseID       string
	FinalImageBase64 string
	OutputFormat     string
	PreviewCount     int
}

func New(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	trimmedBase := strings.TrimSpace(baseURL)
	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedBase == "" {
		return nil, apperr.New(apperr.CodeConfig, "api base url is required")
	}
	if trimmedKey == "" {
		return nil, apperr.New(apperr.CodeConfig, "api key is required")
	}
	return &Client{baseURL: strings.TrimRight(trimmedBase, "/"), apiKey: trimmedKey, httpClient: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) Generate(ctx context.Context, p GenerateParams) (GenerateResult, error) {
	if strings.TrimSpace(p.Prompt) == "" {
		return GenerateResult{}, apperr.New(apperr.CodeInvalidArgs, "prompt is required")
	}
	body := buildRequestBody(p)
	encoded, err := json.Marshal(body)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeInternal, "failed to encode request", err)
	}
	if p.Stream {
		return c.generateStreaming(ctx, encoded)
	}
	return c.generateNonStreaming(ctx, encoded)
}

func buildRequestBody(p GenerateParams) map[string]any {
	tools := map[string]any{"type": "image_generation", "model": p.ImageModel, "size": p.Size, "quality": p.Quality, "output_format": p.OutputFormat, "background": p.Background}
	if p.OutputCompression >= 0 {
		tools["output_compression"] = p.OutputCompression
	}
	body := map[string]any{
		"model":       p.Model,
		"stream":      p.Stream,
		"store":       p.Store,
		"tool_choice": "auto",
		"input":       []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_text", "text": p.Prompt}}}},
		"tools":       []any{tools},
	}
	if strings.TrimSpace(p.PreviousResponseID) != "" {
		body["previous_response_id"] = p.PreviousResponseID
	}
	return body
}

func (c *Client) generateNonStreaming(ctx context.Context, payload []byte) (GenerateResult, error) {
	req, err := c.newRequest(ctx, payload)
	if err != nil {
		return GenerateResult{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeNetwork, "request failed", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return GenerateResult{}, parseHTTPError(resp.Body, resp.StatusCode)
	}
	decoded := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeRPC, "failed to decode response", err)
	}
	return extractFinalResult(decoded)
}

func (c *Client) generateStreaming(ctx context.Context, payload []byte) (GenerateResult, error) {
	req, err := c.newRequest(ctx, payload)
	if err != nil {
		return GenerateResult{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeNetwork, "request failed", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return GenerateResult{}, parseHTTPError(resp.Body, resp.StatusCode)
	}
	return parseSSEStream(resp.Body)
}

func (c *Client) newRequest(ctx context.Context, payload []byte) (*http.Request, error) {
	url := c.baseURL + "/v1/responses"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func extractFinalResult(root map[string]any) (GenerateResult, error) {
	if root == nil {
		return GenerateResult{}, apperr.New(apperr.CodeRPC, "empty response payload")
	}
	out, _ := root["output"].([]any)
	for _, item := range out {
		obj, _ := item.(map[string]any)
		if asString(obj["type"]) != "image_generation_call" {
			continue
		}
		finalB64 := asString(obj["result"])
		if finalB64 == "" {
			continue
		}
		if _, err := base64.StdEncoding.DecodeString(finalB64); err != nil {
			return GenerateResult{}, apperr.Wrap(apperr.CodeRPC, "final image result is not valid base64", err)
		}
		return GenerateResult{ResponseID: asString(root["id"]), FinalImageBase64: finalB64, OutputFormat: normalizedFormat(asString(obj["output_format"]))}, nil
	}
	return GenerateResult{}, apperr.New(apperr.CodeRPC, "missing final image result in output")
}

func normalizedFormat(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return "png"
	}
	return raw
}

func parseHTTPError(body io.Reader, status int) error {
	decoded := map[string]any{}
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return apperr.New(apperr.CodeNetwork, fmt.Sprintf("http request failed with status %d", status))
	}
	errObj, _ := decoded["error"].(map[string]any)
	message := asString(errObj["message"])
	if message == "" {
		message = fmt.Sprintf("http request failed with status %d", status)
	}
	code := asString(errObj["code"])
	if code != "" {
		message = message + " (" + code + ")"
	}
	return apperr.New(apperr.CodeRPC, message)
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(str)
}
