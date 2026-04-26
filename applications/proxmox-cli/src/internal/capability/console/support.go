package consolecap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func parseUPIDNode(upid string) string {
	parts := strings.Split(strings.TrimSpace(upid), ":")
	if len(parts) < 2 {
		return ""
	}
	if strings.ToUpper(parts[0]) != "UPID" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func buildSerialWebsocketPath(node string, vmid int, port string, ticket string) string {
	query := url.Values{}
	query.Set("port", port)
	query.Set("vncticket", ticket)
	return fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?%s", url.PathEscape(node), vmid, query.Encode())
}

func parseScriptCommands(script string) []string {
	trimmed := strings.TrimSpace(script)
	if trimmed == "" {
		return []string{}
	}
	parts := strings.Split(trimmed, "\n")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" {
			continue
		}
		result = append(result, decodeScriptCommand(line))
	}
	return result
}

func decodeScriptCommand(line string) string {
	if strings.EqualFold(line, "<ENTER>") {
		return "\r"
	}
	if strings.EqualFold(line, "<CR>") {
		return "\r"
	}
	replacer := strings.NewReplacer("\\r", "\r", "\\n", "\n", "\\t", "\t")
	decoded := replacer.Replace(line)
	if !strings.HasSuffix(decoded, "\n") && !strings.HasSuffix(decoded, "\r") {
		decoded += "\r"
	}
	return decoded
}

func readSerialUntil(ctx context.Context, conn *websocket.Conn, expect string, timeout time.Duration) (string, bool, error) {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	buffer := strings.Builder{}
	maxTranscriptBytes := 16384
	type wsRead struct {
		payload []byte
		err     error
	}
	reads := make(chan wsRead, 4)
	go func() {
		for {
			_, payload, err := conn.ReadMessage()
			reads <- wsRead{payload: payload, err: err}
			if err != nil {
				return
			}
		}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	keepaliveTicker := time.NewTicker(20 * time.Second)
	defer keepaliveTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return buffer.String(), false, apperr.Wrap(apperr.CodeNetwork, "serial websocket interrupted", ctx.Err())
		case <-timer.C:
			raw := buffer.String()
			return raw, matchesExpect(raw, expect), nil
		case <-keepaliveTicker.C:
			if keepaliveErr := sendWebsocketKeepalive(conn); keepaliveErr != nil {
				if websocket.IsCloseError(keepaliveErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					raw := buffer.String()
					return raw, matchesExpect(raw, expect), nil
				}
				return buffer.String(), false, apperr.Wrap(apperr.CodeNetwork, "failed to send serial websocket keepalive", keepaliveErr)
			}
		case read := <-reads:
			if read.err != nil {
				if websocket.IsCloseError(read.err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					raw := buffer.String()
					return raw, matchesExpect(raw, expect), nil
				}
				return buffer.String(), false, apperr.Wrap(apperr.CodeNetwork, "failed to read serial websocket", read.err)
			}
			if len(read.payload) > 0 {
				buffer.Write(read.payload)
			}
			current := buffer.String()
			if len(current) > maxTranscriptBytes {
				current = current[len(current)-maxTranscriptBytes:]
				buffer.Reset()
				buffer.WriteString(current)
			}
			if expect == "" {
				if len(current) > 0 {
					return current, true, nil
				}
				continue
			}
			clean := normalizeSerialText(current)
			if strings.Contains(clean, "VM ") && strings.Contains(clean, "not running") {
				return current, false, apperr.New(apperr.CodeNetwork, "serial session ended because vm is not running")
			}
			if strings.Contains(clean, "Connection to ") && strings.Contains(clean, " closed") {
				return current, false, apperr.New(apperr.CodeNetwork, "serial connection closed by remote host")
			}
			if matchesExpect(current, expect) {
				return current, true, nil
			}
		}
	}
}

func isVMRunning(ctx context.Context, client *pveapi.Client, node string, vmid int) (bool, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return false, err
	}
	obj := firstObject(unwrapResultField(data))
	if obj == nil {
		return false, apperr.New(apperr.CodeNetwork, "vm status response is not an object")
	}
	status := strings.TrimSpace(asString(obj["status"]))
	return status == "running", nil
}

func sendTermproxyInput(conn *websocket.Conn, input string) error {
	payload := fmt.Sprintf("0:%d:%s", len([]byte(input)), input)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		return err
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(payload)); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "close") {
			return err
		}
	}
	return nil
}

func sendTermproxyResize(conn *websocket.Conn, cols int, rows int) error {
	payload := fmt.Sprintf("1:%d:%d:", cols, rows)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		return err
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(payload)); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "close") {
			return err
		}
	}
	return nil
}

func sendWebsocketKeepalive(conn *websocket.Conn) error {
	deadline := time.Now().Add(3 * time.Second)
	if err := conn.WriteControl(websocket.PingMessage, []byte("keepalive"), deadline); err == nil {
		return nil
	}
	if err := sendTermproxyResize(conn, 120, 40); err != nil {
		return err
	}
	return nil
}

func isWebsocketReadTimeout(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") || strings.Contains(message, "deadline exceeded")
}

func websocketMessageTypeName(messageType int) string {
	switch messageType {
	case websocket.TextMessage:
		return "text"
	case websocket.BinaryMessage:
		return "binary"
	case websocket.PingMessage:
		return "ping"
	case websocket.PongMessage:
		return "pong"
	case websocket.CloseMessage:
		return "close"
	default:
		return strconv.Itoa(messageType)
	}
}

func previewWebsocketPayload(payload []byte, max int) string {
	if len(payload) == 0 {
		return ""
	}
	limit := len(payload)
	if max > 0 && limit > max {
		limit = max
	}
	buf := make([]byte, 0, limit)
	for _, b := range payload[:limit] {
		if b >= 32 && b <= 126 {
			buf = append(buf, b)
			continue
		}
		buf = append(buf, '.')
	}
	return string(buf)
}
