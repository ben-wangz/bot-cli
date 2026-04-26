package pveapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	timeout    time.Duration
	insecure   bool
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
		timeout:    timeout,
		insecure:   config.InsecureTLS,
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
		c.logger.Printf("request method=%s path=%s headers=%v content_length=%d transfer_encoding=%v", req.Method, redact.String(req.URL.String()), redactedHeaders, req.ContentLength, req.TransferEncoding)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "request failed", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		_ = resp.Body.Close()
		message := fmt.Sprintf("remote returned status %d", resp.StatusCode)
		text := strings.TrimSpace(string(body))
		if text != "" {
			message += ": " + text
		}
		return resp, apperr.New(apperr.CodeNetwork, message)
	}
	return resp, nil
}

func (c *Client) GetData(ctx context.Context, path string, query url.Values) (any, error) {
	requestPath := withQuery(path, query)
	req, err := c.NewRequest(ctx, http.MethodGet, requestPath, nil)
	if err != nil {
		return nil, err
	}
	return c.doAndDecode(req)
}

func (c *Client) PostFormData(ctx context.Context, path string, form url.Values) (any, error) {
	req, err := c.NewRequest(ctx, http.MethodPost, path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.doAndDecode(req)
}

func (c *Client) PutFormData(ctx context.Context, path string, form url.Values) (any, error) {
	req, err := c.NewRequest(ctx, http.MethodPut, path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.doAndDecode(req)
}

func (c *Client) DeleteFormData(ctx context.Context, path string, form url.Values) (any, error) {
	req, err := c.NewRequest(ctx, http.MethodDelete, path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.doAndDecode(req)
}

func (c *Client) PostMultipartFile(ctx context.Context, path string, fields map[string]string, fileField string, filePath string, uploadFilename string) (any, error) {
	if strings.TrimSpace(fileField) == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "multipart file field is required")
	}
	name := strings.TrimSpace(uploadFilename)
	if name == "" {
		name = filepath.Base(filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to open upload file", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to stat upload file", err)
	}
	if info.Size() < 0 {
		return nil, apperr.New(apperr.CodeConfig, "upload file size is invalid")
	}

	contentType, prefix, suffix, err := buildMultipartPrefixSuffix(fields, fileField, name)
	if err != nil {
		return nil, err
	}
	body := io.MultiReader(bytes.NewReader(prefix), file, bytes.NewReader(suffix))

	req, err := c.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(prefix)) + info.Size() + int64(len(suffix))
	return c.doAndDecode(req)
}

func (c *Client) doAndDecode(req *http.Request) (any, error) {
	resp, err := c.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data any `json:"data"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&envelope); decodeErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to decode api response", decodeErr)
	}
	return envelope.Data, nil
}

func withQuery(path string, query url.Values) string {
	if len(query) == 0 {
		return path
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + query.Encode()
}

func headersToMap(header http.Header) map[string]string {
	result := map[string]string{}
	for key, values := range header {
		result[key] = strings.Join(values, ";")
	}
	return result
}
