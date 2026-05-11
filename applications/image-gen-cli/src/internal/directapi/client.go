package directapi

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
	Prompt            string
	ImageModel        string
	Stream            bool
	Size              string
	Quality           string
	OutputFormat      string
	OutputCompression int
	Background        string
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
	body := map[string]any{
		"model":         p.ImageModel,
		"prompt":        p.Prompt,
		"size":          p.Size,
		"quality":       p.Quality,
		"output_format": p.OutputFormat,
		"background":    p.Background,
	}
	if p.OutputCompression >= 0 {
		body["output_compression"] = p.OutputCompression
	}
	if p.Stream {
		body["stream"] = true
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeInternal, "failed to encode request", err)
	}
	req, err := c.newRequest(ctx, encoded)
	if err != nil {
		return GenerateResult{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeNetwork, "request failed (direct image generation)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return GenerateResult{}, parseHTTPError(resp.Body, resp.StatusCode)
	}
	if p.Stream {
		return decodeStreamingResponse(resp.Body)
	}
	return decodeResponse(resp.Body)
}

func (c *Client) newRequest(ctx context.Context, payload []byte) (*http.Request, error) {
	url := c.baseURL + "/v1/images/generations"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func decodeResponse(body io.Reader) (GenerateResult, error) {
	decoded := map[string]any{}
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeRPC, "failed to decode direct response", err)
	}
	data, _ := decoded["data"].([]any)
	if len(data) == 0 {
		return GenerateResult{}, apperr.New(apperr.CodeRPC, "missing image data in response")
	}
	first, _ := data[0].(map[string]any)
	imageBase64 := asString(first["b64_json"])
	if imageBase64 == "" {
		return GenerateResult{}, apperr.New(apperr.CodeRPC, "missing b64_json in response image data")
	}
	if _, err := base64.StdEncoding.DecodeString(imageBase64); err != nil {
		return GenerateResult{}, apperr.Wrap(apperr.CodeRPC, "final image result is not valid base64", err)
	}
	return GenerateResult{ResponseID: asString(decoded["id"]), FinalImageBase64: imageBase64, OutputFormat: normalizedFormat(asString(first["output_format"])), PreviewCount: 0}, nil
}

func parseHTTPError(body io.Reader, status int) error {
	decoded := map[string]any{}
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return apperr.New(apperr.CodeNetwork, fmt.Sprintf("http request failed with status %d (direct)", status))
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

func normalizedFormat(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return "png"
	}
	return raw
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
