package consolecap

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func RunOpenVMTermproxy(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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

func RunSerialWSSessionControl(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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
	if sendErr := sendSerialScriptCommands(conn, commands); sendErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send serial command", sendErr)
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
