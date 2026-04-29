package aria2rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
)

type Client struct {
	endpoint string
	secret   string
	http     *http.Client
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

func New(endpoint, secret string, timeout time.Duration) (*Client, error) {
	clean := strings.TrimSpace(endpoint)
	if clean == "" {
		return nil, apperr.New(apperr.CodeConfig, "rpc endpoint is required")
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		endpoint: clean,
		secret:   strings.TrimSpace(secret),
		http:     &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) Call(ctx context.Context, method string, params []any) (json.RawMessage, error) {
	if !strings.Contains(method, ".") {
		method = "aria2." + method
	}
	requestParams := append([]any{}, params...)
	if strings.HasPrefix(method, "aria2.") && c.secret != "" {
		requestParams = append([]any{"token:" + c.secret}, requestParams...)
	}
	payload := rpcRequest{JSONRPC: "2.0", ID: "aria2-cli", Method: method, Params: requestParams}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to encode rpc payload", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to create rpc request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to call rpc endpoint", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		if rpcErr := parseRPCErrorFromBody(method, b); rpcErr != nil {
			return nil, rpcErr
		}
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("rpc endpoint status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b))))
	}
	var parsed rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to decode rpc response", err)
	}
	if parsed.Error != nil {
		retryable := parsed.Error.Code == -32603
		return nil, apperr.New(apperr.CodeRPC, fmt.Sprintf("rpc method=%s code=%d retryable=%t message=%s", method, parsed.Error.Code, retryable, parsed.Error.Message))
	}
	return parsed.Result, nil
}

func (c *Client) Endpoint() string {
	return c.endpoint
}

func (c *Client) Secret() string {
	return c.secret
}

func parseRPCErrorFromBody(method string, body []byte) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil
	}
	var parsed rpcResponse
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil
	}
	if parsed.Error == nil {
		return nil
	}
	retryable := parsed.Error.Code == -32603
	return apperr.New(apperr.CodeRPC, fmt.Sprintf("rpc method=%s code=%d retryable=%t message=%s", method, parsed.Error.Code, retryable, parsed.Error.Message))
}
