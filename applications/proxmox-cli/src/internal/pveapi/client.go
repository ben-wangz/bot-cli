package pveapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"

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

func buildMultipartPrefixSuffix(fields map[string]string, fileField string, filename string) (string, []byte, []byte, error) {
	const placeholder = "__BOT_CLI_MULTIPART_FILE_PLACEHOLDER__"
	buffer := bytes.Buffer{}
	writer := multipart.NewWriter(&buffer)
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if strings.TrimSpace(key) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writer.WriteField(key, fields[key]); err != nil {
			return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to build multipart field", err)
		}
	}
	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to create multipart file part", err)
	}
	if _, err := io.WriteString(part, placeholder); err != nil {
		return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to build multipart placeholder", err)
	}
	if err := writer.Close(); err != nil {
		return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to finalize multipart payload metadata", err)
	}
	raw := buffer.Bytes()
	marker := []byte(placeholder)
	index := bytes.Index(raw, marker)
	if index < 0 {
		return "", nil, nil, apperr.New(apperr.CodeInternal, "multipart placeholder marker not found")
	}
	prefix := append([]byte(nil), raw[:index]...)
	suffix := append([]byte(nil), raw[index+len(marker):]...)
	return writer.FormDataContentType(), prefix, suffix, nil
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

func (c *Client) DialWebsocket(ctx context.Context, path string, query url.Values) (*websocket.Conn, *http.Response, error) {
	requestPath := withQuery(path, query)
	fullURL, err := c.websocketURL(requestPath)
	if err != nil {
		return nil, nil, err
	}
	headers := http.Header{}
	for key, value := range c.headers {
		headers.Set(key, value)
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: c.timeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: c.insecure}, //nolint:gosec
	}
	conn, resp, dialErr := dialer.DialContext(ctx, fullURL, headers)
	if dialErr != nil {
		if resp != nil {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			_ = resp.Body.Close()
			message := fmt.Sprintf("websocket dial failed status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
			return nil, resp, apperr.Wrap(apperr.CodeNetwork, message, dialErr)
		}
		return nil, resp, apperr.Wrap(apperr.CodeNetwork, "websocket dial failed", dialErr)
	}
	return conn, resp, nil
}

func (c *Client) websocketURL(path string) (string, error) {
	baseParsed, err := url.Parse(c.baseURL)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeConfig, "invalid api base url", err)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	pathParsed, err := url.Parse(path)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeInvalidArgs, "invalid websocket path", err)
	}
	baseParsed.Path = strings.TrimRight(baseParsed.Path, "/") + pathParsed.Path
	baseParsed.RawQuery = pathParsed.RawQuery
	if baseParsed.Scheme == "https" {
		baseParsed.Scheme = "wss"
	} else {
		baseParsed.Scheme = "ws"
	}
	return baseParsed.String(), nil
}
