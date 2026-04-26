package action

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

func runStartVNCProxy(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	form := mapArgsToForm(req.Args, "node", "vmid")
	websocketMode := strings.TrimSpace(req.Args["websocket"])
	if websocketMode == "" {
		websocketMode = "1"
	}
	if !isOneOf(websocketMode, "0", "1") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "websocket must be 0 or 1")
	}
	form.Set("websocket", websocketMode)
	if rawGeneratePassword := strings.TrimSpace(req.Args["generate-password"]); rawGeneratePassword != "" {
		if !isOneOf(rawGeneratePassword, "0", "1") {
			return nil, apperr.New(apperr.CodeInvalidArgs, "generate-password must be 0 or 1")
		}
		form.Set("generate-password", rawGeneratePassword)
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy", node, vmid)
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	payload := firstObject(unwrapResultField(data))
	if payload == nil {
		return nil, apperr.New(apperr.CodeNetwork, "vncproxy response is not an object")
	}
	port := strings.TrimSpace(asString(payload["port"]))
	ticket := strings.TrimSpace(asString(payload["ticket"]))
	proxyNode := strings.TrimSpace(asString(payload["node"]))
	if proxyNode == "" {
		proxyNode = parseUPIDNode(asString(payload["upid"]))
	}
	if proxyNode == "" {
		proxyNode = node
	}
	result := map[string]any{
		"port":       port,
		"ticket":     ticket,
		"cert":       asString(payload["cert"]),
		"user":       asString(payload["user"]),
		"password":   asString(payload["password"]),
		"upid":       asString(payload["upid"]),
		"proxy_node": proxyNode,
		"websocket":  buildVNCWebsocketPath(proxyNode, vmid, port, ticket),
	}
	request := map[string]any{"node": node, "vmid": vmid, "websocket": websocketMode == "1"}
	return buildResult(req, request, result, map[string]any{}), nil
}

func runConnectVNCWebsocket(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	port := strings.TrimSpace(req.Args["port"])
	ticket := strings.TrimSpace(req.Args["ticket"])
	proxyNode := node
	if port == "" || ticket == "" {
		proxyData, proxyErr := runStartVNCProxy(ctx, client, Request{Name: "start_vnc_proxy", Args: req.Args, Scope: req.Scope})
		if proxyErr != nil {
			return nil, proxyErr
		}
		proxyResult, _ := proxyData["result"].(map[string]any)
		port = strings.TrimSpace(asString(proxyResult["port"]))
		ticket = strings.TrimSpace(asString(proxyResult["ticket"]))
		if fromProxy := strings.TrimSpace(asString(proxyResult["proxy_node"])); fromProxy != "" {
			proxyNode = fromProxy
		}
	}
	if port == "" || ticket == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "missing vnc websocket endpoint: provide --port and --ticket or allow proxy bootstrap")
	}
	probeSeconds := 2
	if rawProbe := strings.TrimSpace(req.Args["probe-seconds"]); rawProbe != "" {
		parsedProbe, parseErr := strconv.Atoi(rawProbe)
		if parseErr != nil || parsedProbe < 0 {
			return nil, apperr.New(apperr.CodeInvalidArgs, "probe-seconds must be an integer >= 0")
		}
		probeSeconds = parsedProbe
	}
	wsPath := buildVNCWebsocketPath(proxyNode, vmid, port, ticket)
	conn, _, dialErr := client.DialWebsocket(ctx, wsPath, url.Values{})
	if dialErr != nil {
		if proxyNode != node {
			fallbackPath := buildVNCWebsocketPath(node, vmid, port, ticket)
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
	messageType := ""
	probePreview := ""
	probeBytes := 0
	if probeSeconds > 0 {
		if deadlineErr := conn.SetReadDeadline(time.Now().Add(time.Duration(probeSeconds) * time.Second)); deadlineErr != nil {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to set websocket probe deadline", deadlineErr)
		}
		msgType, payload, readErr := conn.ReadMessage()
		if readErr != nil {
			if !isWebsocketReadTimeout(readErr) && !websocket.IsCloseError(readErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil, apperr.Wrap(apperr.CodeNetwork, "failed to probe vnc websocket", readErr)
			}
		} else {
			messageType = websocketMessageTypeName(msgType)
			probeBytes = len(payload)
			probePreview = previewWebsocketPayload(payload, 64)
		}
	}
	request := map[string]any{"node": node, "vmid": vmid, "port": port, "probe_seconds": probeSeconds}
	result := map[string]any{
		"connected":      true,
		"websocket":      wsPath,
		"probe_received": probeBytes > 0,
		"probe_bytes":    probeBytes,
		"probe_type":     messageType,
		"probe_preview":  probePreview,
	}
	return buildResult(req, request, result, map[string]any{"probe_seconds": probeSeconds}), nil
}

func runSerialWSCaptureToFile(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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

	termproxyData, err := runOpenVMTermproxy(ctx, client, Request{Name: "open_vm_termproxy", Args: req.Args, Scope: req.Scope})
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
	time.Sleep(300 * time.Millisecond)
	if len(commands) > 0 {
		time.Sleep(250 * time.Millisecond)
	}
	for _, command := range commands {
		if writeErr := sendTermproxyInput(conn, command); writeErr != nil {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send serial command", writeErr)
		}
		if strings.TrimSpace(command) != "" {
			time.Sleep(120 * time.Millisecond)
		}
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

func runOpenVMTermproxy(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	form := mapArgsToForm(req.Args, "node", "vmid")
	serial := strings.TrimSpace(req.Args["serial"])
	if serial == "" {
		serial = "serial0"
	}
	if serial == "0" {
		serial = "serial0"
	}
	form.Set("serial", serial)
	if !isOneOf(serial, "serial0", "serial1", "serial2", "serial3") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "serial must be one of serial0|serial1|serial2|serial3")
	}
	if escape := strings.TrimSpace(req.Args["escape"]); escape != "" {
		form.Set("escape", escape)
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/termproxy", node, vmid)
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	payload := firstObject(unwrapResultField(data))
	if payload == nil {
		return nil, apperr.New(apperr.CodeNetwork, "termproxy response is not an object")
	}
	port := strings.TrimSpace(asString(payload["port"]))
	ticket := strings.TrimSpace(asString(payload["ticket"]))
	user := strings.TrimSpace(asString(payload["user"]))
	proxyNode := strings.TrimSpace(asString(payload["node"]))
	if proxyNode == "" {
		proxyNode = parseUPIDNode(asString(payload["upid"]))
	}
	if proxyNode == "" {
		proxyNode = node
	}
	result := map[string]any{
		"port":       port,
		"ticket":     ticket,
		"user":       user,
		"upid":       asString(payload["upid"]),
		"cert":       asString(payload["cert"]),
		"proxy_node": proxyNode,
		"websocket":  buildSerialWebsocketPath(node, vmid, port, ticket),
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid}, result, map[string]any{}), nil
}

func runValidateK1SerialReadable(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	sessionArgs := map[string]string{}
	for key, value := range req.Args {
		sessionArgs[key] = value
	}
	if strings.TrimSpace(sessionArgs["timeout-seconds"]) == "" {
		sessionArgs["timeout-seconds"] = "20"
	}
	sessionData, err := runSerialWSSessionControl(ctx, client, Request{Name: "serial_ws_session_control", Args: sessionArgs, Scope: req.Scope})
	if err != nil {
		return nil, err
	}
	sessionResult, _ := sessionData["result"].(map[string]any)
	rawTranscript := asString(sessionResult["transcript"])
	cleanTranscript := normalizeSerialText(rawTranscript)
	if strings.TrimSpace(cleanTranscript) == "" {
		return nil, apperr.New(apperr.CodeNetwork, "serial output is empty; cannot validate readability")
	}
	bannerOnly := isOnlySerialStartupBanner(cleanTranscript)
	request := map[string]any{"node": node, "vmid": vmid, "timeout_seconds": sessionArgs["timeout-seconds"], "expect": strings.TrimSpace(sessionArgs["expect"])}
	result := map[string]any{
		"readable":         true,
		"banner_only":      bannerOnly,
		"transcript_clean": cleanTranscript,
		"transcript_tail":  tailText(cleanTranscript, 240),
		"bytes":            len(rawTranscript),
		"websocket":        asString(sessionResult["websocket"]),
	}
	return buildResult(req, request, result, map[string]any{"criterion": "k1_serial_readable", "banner_only": bannerOnly}), nil
}

func runValidateSerialOutputCriterion2(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	captureArgs := map[string]string{}
	for key, value := range req.Args {
		captureArgs[key] = value
	}
	if strings.TrimSpace(captureArgs["log-path"]) == "" {
		captureArgs["log-path"] = filepath.Join("build", fmt.Sprintf("serial-criterion2-%d.log", vmid))
	}
	if strings.TrimSpace(captureArgs["append"]) == "" {
		captureArgs["append"] = "1"
	}
	if strings.TrimSpace(captureArgs["timeout-seconds"]) == "" {
		captureArgs["timeout-seconds"] = "120"
	}
	captureData, err := runSerialWSCaptureToFile(ctx, client, Request{Name: "serial_ws_capture_to_file", Args: captureArgs, Scope: req.Scope})
	if err != nil {
		return nil, err
	}
	captureResult, _ := captureData["result"].(map[string]any)
	cleanTranscript := normalizeSerialText(asString(captureResult["transcript_clean"]))
	if strings.TrimSpace(cleanTranscript) == "" {
		return nil, apperr.New(apperr.CodeNetwork, "serial output is empty; criterion2 failed")
	}
	if isOnlySerialStartupBanner(cleanTranscript) {
		return nil, apperr.New(apperr.CodeNetwork, "serial output only contains termproxy startup banner; verify kernel cmdline enables serial console")
	}
	request := map[string]any{
		"node":            node,
		"vmid":            vmid,
		"log_path":        captureArgs["log-path"],
		"append":          captureArgs["append"] == "1",
		"timeout_seconds": captureArgs["timeout-seconds"],
		"expect":          strings.TrimSpace(captureArgs["expect"]),
	}
	result := map[string]any{
		"criterion2_passed": true,
		"log_path":          captureResult["log_path"],
		"bytes_written":     captureResult["bytes_written"],
		"matched":           captureResult["matched"],
		"transcript_clean":  cleanTranscript,
		"transcript_tail":   tailText(cleanTranscript, 240),
		"websocket":         captureResult["websocket"],
	}
	return buildResult(req, request, result, map[string]any{"criterion": "serial_output_criterion2"}), nil
}

func runSerialWSSessionControl(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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
	expect := strings.TrimSpace(req.Args["expect"])
	timeout := 60 * time.Second
	if strings.TrimSpace(req.Args["timeout-seconds"]) != "" {
		seconds, parseErr := RequiredInt(req.Args, "timeout-seconds")
		if parseErr != nil {
			return nil, parseErr
		}
		timeout = time.Duration(seconds) * time.Second
	}
	termproxyData, err := runOpenVMTermproxy(ctx, client, Request{Name: "open_vm_termproxy", Args: req.Args, Scope: req.Scope})
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
	conn, _, err := client.DialWebsocket(ctx, wsPath, url.Values{})
	if err != nil {
		proxyNode := strings.TrimSpace(asString(resultData["proxy_node"]))
		if proxyNode != "" && proxyNode != node {
			fallbackPath := buildSerialWebsocketPath(proxyNode, vmid, port, ticket)
			fallbackConn, _, fallbackErr := client.DialWebsocket(ctx, fallbackPath, url.Values{})
			if fallbackErr == nil {
				conn = fallbackConn
				wsPath = fallbackPath
				err = nil
			}
		}
	}
	if err != nil {
		return nil, err
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
	time.Sleep(300 * time.Millisecond)
	if len(commands) > 0 {
		time.Sleep(250 * time.Millisecond)
	}
	for _, command := range commands {
		if writeErr := sendTermproxyInput(conn, command); writeErr != nil {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send serial command", writeErr)
		}
		if strings.TrimSpace(command) != "" {
			time.Sleep(120 * time.Millisecond)
		}
	}
	if deadlineErr := conn.SetReadDeadline(time.Time{}); deadlineErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to reset websocket read deadline", deadlineErr)
	}
	transcript, matched, err := readSerialUntil(ctx, conn, expect, timeout)
	if err != nil {
		return nil, err
	}
	matched = matchesExpect(transcript, expect)
	diagnostics := map[string]any{"matched_expect": matched, "script_command_count": len(commands), "timeout_seconds": int(timeout.Seconds())}
	cleanTranscript := normalizeSerialText(transcript)
	result := map[string]any{"transcript": transcript, "transcript_clean": cleanTranscript, "expect": expect, "matched": matched, "websocket": wsPath}
	if expect != "" && !matched {
		snippet := cleanTranscript
		if len(snippet) > 240 {
			snippet = snippet[len(snippet)-240:]
		}
		return nil, apperr.New(apperr.CodeNetwork, "serial session timeout before expected output: "+expect+"; transcript_tail="+strconv.Quote(snippet))
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid, "expect": expect}, result, diagnostics), nil
}
