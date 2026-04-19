package pveapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/redact"
)

type Config struct {
	BaseURL     string
	Timeout     time.Duration
	InsecureTLS bool
	Headers     map[string]string
	Logger      *log.Logger
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	headers    map[string]string
	logger     *log.Logger
}

func New(config Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		return nil, apperr.New(apperr.CodeConfig, "api-base is required")
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureTLS}} //nolint:gosec
	httpClient := &http.Client{Timeout: timeout, Transport: transport}

	headers := map[string]string{}
	for key, value := range config.Headers {
		headers[key] = value
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		headers:    headers,
		logger:     config.Logger,
	}, nil
}

func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if method == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "http method is required")
	}
	if path == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "api path is required")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to build request", err)
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.logger != nil {
		redactedHeaders := redact.Headers(headersToMap(req.Header))
		c.logger.Printf("request method=%s path=%s headers=%v", req.Method, redact.String(req.URL.String()), redactedHeaders)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "request failed", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return resp, apperr.New(apperr.CodeNetwork, fmt.Sprintf("remote returned status %d", resp.StatusCode))
	}
	return resp, nil
}

func headersToMap(header http.Header) map[string]string {
	result := map[string]string{}
	for key, values := range header {
		result[key] = strings.Join(values, ";")
	}
	return result
}
