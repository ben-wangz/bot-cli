package action

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/auth"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase5(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "node_termproxy_shell_exec":
		return runNodeTermproxyShellExec(ctx, client, req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 5: "+req.Name)
	}
}

func runNodeTermproxyShellExec(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	if req.Scope != auth.ScopeRoot && req.Scope != auth.ScopeRootToken {
		return nil, apperr.New(apperr.CodeAuth, "node_termproxy_shell_exec requires --auth-scope root or root-token")
	}
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	cmd := strings.TrimSpace(req.Args["cmd"])
	if cmd == "" {
		cmd = "login"
	}
	if !isOneOf(cmd, "login", "upgrade", "ceph_install") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "cmd must be one of login|upgrade|ceph_install")
	}
	if cmd != "login" && req.Scope == auth.ScopeRootToken {
		return nil, apperr.New(apperr.CodeAuth, "cmd requires --auth-scope root with root@pam session; root-token is limited to login shell path")
	}
	form := url.Values{}
	form.Set("cmd", cmd)
	if cmdOpt := strings.TrimSpace(req.Args["cmd-opt"]); cmdOpt != "" {
		encodedOpts := encodeCmdOpts(cmdOpt)
		if encodedOpts != "" {
			form.Set("cmd-opts", encodedOpts)
		}
	}
	if cmdOpts := strings.TrimSpace(req.Args["cmd-opts"]); cmdOpts != "" {
		form.Set("cmd-opts", cmdOpts)
	}
	path := fmt.Sprintf("/nodes/%s/termproxy", url.PathEscape(node))
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	payload := firstObject(unwrapResultField(data))
	if payload == nil {
		return nil, apperr.New(apperr.CodeNetwork, "node termproxy response is not an object")
	}
	port := strings.TrimSpace(asString(payload["port"]))
	ticket := strings.TrimSpace(asString(payload["ticket"]))
	user := strings.TrimSpace(asString(payload["user"]))
	if port == "" || ticket == "" || user == "" {
		return nil, apperr.New(apperr.CodeNetwork, "node termproxy did not return port/ticket/user")
	}
	proxyNode := strings.TrimSpace(asString(payload["node"]))
	if proxyNode == "" {
		proxyNode = parseUPIDNode(asString(payload["upid"]))
	}
	if proxyNode == "" {
		proxyNode = node
	}

	wsPath := buildNodeWebsocketPath(node, port, ticket)
	conn, _, err := client.DialWebsocket(ctx, wsPath, url.Values{})
	if err != nil {
		if proxyNode != node {
			fallback := buildNodeWebsocketPath(proxyNode, port, ticket)
			fallbackConn, _, fallbackErr := client.DialWebsocket(ctx, fallback, url.Values{})
			if fallbackErr == nil {
				conn = fallbackConn
				wsPath = fallback
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
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send node websocket auth line", writeErr)
		}
	}
	if resizeErr := sendTermproxyResize(conn, 120, 40); resizeErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send node terminal resize", resizeErr)
	}
	commands := parseScriptCommands(req.Args["script"])
	time.Sleep(3 * time.Second)
	if len(commands) > 0 {
		time.Sleep(250 * time.Millisecond)
	}
	for _, command := range commands {
		if writeErr := sendTermproxyInput(conn, command); writeErr != nil {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to send node shell command", writeErr)
		}
		if strings.TrimSpace(command) != "" {
			time.Sleep(120 * time.Millisecond)
		}
	}
	if deadlineErr := conn.SetReadDeadline(time.Time{}); deadlineErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to reset node websocket read deadline", deadlineErr)
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
	transcript, matched, err := readSerialUntil(ctx, conn, expect, timeout)
	if err != nil {
		return nil, err
	}
	matched = matchesExpect(transcript, expect)
	clean := normalizeSerialText(transcript)
	result := map[string]any{
		"port":             port,
		"ticket":           ticket,
		"user":             user,
		"cmd":              cmd,
		"proxy_node":       proxyNode,
		"websocket":        wsPath,
		"expect":           expect,
		"matched":          matched,
		"transcript":       transcript,
		"transcript_clean": clean,
	}
	diagnostics := map[string]any{
		"script_command_count": len(commands),
		"timeout_seconds":      int(timeout.Seconds()),
		"matched_expect":       matched,
	}
	if expect != "" && !matched {
		tail := clean
		if len(tail) > 240 {
			tail = tail[len(tail)-240:]
		}
		return nil, apperr.New(apperr.CodeNetwork, "node termproxy timeout before expected output: "+expect+"; transcript_tail="+tail)
	}
	return buildResult(req, map[string]any{"node": node, "cmd": cmd, "expect": expect}, result, diagnostics), nil
}

func buildNodeWebsocketPath(node string, port string, ticket string) string {
	query := url.Values{}
	query.Set("port", port)
	query.Set("vncticket", ticket)
	return fmt.Sprintf("/nodes/%s/vncwebsocket?%s", url.PathEscape(node), query.Encode())
}

func encodeCmdOpts(raw string) string {
	parts := strings.Split(raw, ",")
	builder := strings.Builder{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		builder.WriteString(trimmed)
		builder.WriteByte(0)
	}
	return builder.String()
}
