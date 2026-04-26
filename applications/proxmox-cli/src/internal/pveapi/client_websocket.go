package pveapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

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
