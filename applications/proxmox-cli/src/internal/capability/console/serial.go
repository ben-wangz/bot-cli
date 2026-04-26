package consolecap

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func RunSerialWSCaptureToFile(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	running, statusErr := isVMRunning(ctx, client, node, vmid)
	if statusErr != nil {
		return nil, statusErr
	}
	if !running {
		return nil, apperr.New(apperr.CodeInvalidArgs, fmt.Sprintf("vm %d is not running", vmid))
	}
	logPath, err := RequiredString(req.Args, "log-path")
	if err != nil {
		return nil, err
	}
	appendMode := strings.TrimSpace(req.Args["append"])
	if appendMode == "" {
		appendMode = "1"
	}
	if !isOneOf(appendMode, "0", "1") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "append must be 0 or 1")
	}
	if mkErr := os.MkdirAll(filepath.Dir(logPath), 0o755); mkErr != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create log directory", mkErr)
	}
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode == "1" {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, openErr := os.OpenFile(logPath, flags, 0o644)
	if openErr != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to open log file", openErr)
	}
	defer file.Close()

	expect := strings.TrimSpace(req.Args["expect"])
	timeout := 120 * time.Second
	if rawTimeout := strings.TrimSpace(req.Args["timeout-seconds"]); rawTimeout != "" {
		seconds, parseErr := strconv.Atoi(rawTimeout)
		if parseErr != nil || seconds < 0 {
			return nil, apperr.New(apperr.CodeInvalidArgs, "timeout-seconds must be an integer >= 0")
		}
		timeout = time.Duration(seconds) * time.Second
	}

	termproxyData, err := RunOpenVMTermproxy(ctx, client, Request{Name: "open_vm_termproxy", Args: req.Args, Scope: req.Scope})
	if err != nil {
		return nil, err
	}
	resultData, _ := termproxyData["result"].(map[string]any)
	port := strings.TrimSpace(asString(resultData["port"]))
	ticket := strings.TrimSpace(asString(resultData["ticket"]))
	user := strings.TrimSpace(asString(resultData["user"]))
	if port == "" || ticket == "" || user == "" {
		return nil, apperr.New(apperr.CodeNetwork, "termproxy did not return port/ticket/user")
	}
	wsPath := buildSerialWebsocketPath(node, vmid, port, ticket)
	conn, _, dialErr := client.DialWebsocket(ctx, wsPath, url.Values{})
	if dialErr != nil {
		proxyNode := strings.TrimSpace(asString(resultData["proxy_node"]))
		if proxyNode != "" && proxyNode != node {
			fallbackPath := buildSerialWebsocketPath(proxyNode, vmid, port, ticket)
			fallbackConn, _, fallbackErr := client.DialWebsocket(ctx, fallbackPath, url.Values{})
			if fallbackErr == nil {
				conn = fallbackConn
				wsPath = fallbackPath
				dialErr = nil
			}
		}
	}
	if dialErr != nil {
		return nil, dialErr
	}
	defer conn.Close()

	if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(user+":"+ticket+"\n")); writeErr != nil {
		if !strings.Contains(strings.ToLower(writeErr.Error()), "close") {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send websocket auth line", writeErr)
		}
	}
	if resizeErr := sendTermproxyResize(conn, 120, 40); resizeErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send initial terminal resize", resizeErr)
	}
	commands := parseScriptCommands(req.Args["script"])
	if sendErr := sendSerialScriptCommands(conn, commands); sendErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send serial command", sendErr)
	}

	type wsRead struct {
		payload []byte
		err     error
	}
	reads := make(chan wsRead, 4)
	go func() {
		for {
			_, payload, readErr := conn.ReadMessage()
			reads <- wsRead{payload: payload, err: readErr}
			if readErr != nil {
				return
			}
		}
	}()

	buffer := strings.Builder{}
	maxTranscriptBytes := 32768
	bytesWritten := int64(0)
	closedByRemote := false
	matched := false
	stopOnExpect := expect != ""

	var timer *time.Timer
	var timeoutC <-chan time.Time
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		timeoutC = timer.C
		defer timer.Stop()
	}
	keepaliveTicker := time.NewTicker(20 * time.Second)
	defer keepaliveTicker.Stop()

	for {
		if stopOnExpect && matched {
			break
		}
		select {
		case <-ctx.Done():
			return nil, apperr.Wrap(apperr.CodeNetwork, "serial websocket interrupted", ctx.Err())
		case <-timeoutC:
			goto finished
		case <-keepaliveTicker.C:
			if keepaliveErr := sendWebsocketKeepalive(conn); keepaliveErr != nil {
				if websocket.IsCloseError(keepaliveErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					closedByRemote = true
					goto finished
				}
				return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send serial websocket keepalive", keepaliveErr)
			}
		case read := <-reads:
			if read.err != nil {
				if websocket.IsCloseError(read.err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					closedByRemote = true
					goto finished
				}
				return nil, apperr.Wrap(apperr.CodeNetwork, "failed to read serial websocket", read.err)
			}
			if len(read.payload) == 0 {
				continue
			}
			written, writeErr := file.Write(read.payload)
			if writeErr != nil {
				return nil, apperr.Wrap(apperr.CodeConfig, "failed to write serial log file", writeErr)
			}
			bytesWritten += int64(written)
			if syncErr := file.Sync(); syncErr != nil {
				return nil, apperr.Wrap(apperr.CodeConfig, "failed to flush serial log file", syncErr)
			}
			buffer.Write(read.payload)
			current := buffer.String()
			if len(current) > maxTranscriptBytes {
				current = current[len(current)-maxTranscriptBytes:]
				buffer.Reset()
				buffer.WriteString(current)
			}
			if stopOnExpect && matchesExpect(current, expect) {
				matched = true
			}
		}
	}

finished:
	transcriptTail := buffer.String()
	cleanTail := normalizeSerialText(transcriptTail)
	if stopOnExpect && !matched {
		snippet := cleanTail
		if len(snippet) > 240 {
			snippet = snippet[len(snippet)-240:]
		}
		return nil, apperr.New(apperr.CodeNetwork, "serial capture ended before expected output: "+expect+"; transcript_tail="+strconv.Quote(snippet))
	}
	result := map[string]any{
		"log_path":           logPath,
		"bytes_written":      bytesWritten,
		"expect":             expect,
		"matched":            matched,
		"transcript_tail":    transcriptTail,
		"transcript_clean":   cleanTail,
		"websocket":          wsPath,
		"closed_by_remote":   closedByRemote,
		"append_mode":        appendMode == "1",
		"script_command_cnt": len(commands),
	}
	diagnostics := map[string]any{"timeout_seconds": int(timeout.Seconds()), "matched_expect": matched}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid, "log_path": logPath, "expect": expect}, result, diagnostics), nil
}
