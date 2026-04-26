package consolecap

import (
	"strconv"

	"github.com/gorilla/websocket"
)

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
