package capability

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
)

func runEnsureDaemonStarted(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	version, state, err := probeVersion(ctx, client)
	if err == nil {
		return envelope(req, map[string]any{
			"already_running": true,
			"version":         version,
		}, map[string]any{"probe_state": state}), nil
	}
	if state == "rpc_reachable_auth_or_rpc_error" {
		return nil, apperr.New(apperr.CodeConfig, "rpc endpoint is reachable but authentication failed or rpc is rejected; refuse to start another daemon")
	}
	rpcURL, parseErr := url.Parse(client.Endpoint())
	if parseErr != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "invalid rpc endpoint", parseErr)
	}
	host := strings.TrimSpace(rpcURL.Hostname())
	if host == "" {
		host = "127.0.0.1"
	}
	if !isLocalHost(host) {
		return nil, apperr.New(apperr.CodeConfig, "ensure_daemon_started only supports localhost endpoints")
	}
	port := rpcURL.Port()
	if port == "" {
		port = "6800"
	}
	if _, convErr := strconv.Atoi(port); convErr != nil {
		return nil, apperr.New(apperr.CodeConfig, "rpc endpoint port must be numeric")
	}
	listenAll := false
	if host == "0.0.0.0" {
		listenAll = true
	}
	cmdArgs := []string{
		"--daemon=true",
		"--enable-rpc=true",
		"--rpc-listen-port=" + port,
		"--rpc-listen-all=" + strconv.FormatBool(listenAll),
	}
	if secret := strings.TrimSpace(client.Secret()); secret != "" {
		cmdArgs = append(cmdArgs, "--rpc-secret="+secret)
	}
	startCmd := exec.CommandContext(ctx, "aria2c", cmdArgs...)
	if runErr := startCmd.Run(); runErr != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to start aria2c daemon", runErr)
	}
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		version, state, callErr := probeVersion(ctx, client)
		if callErr == nil {
			return envelope(req, map[string]any{
				"already_running": false,
				"started":         true,
				"version":         version,
			}, map[string]any{"probe_state": state, "start_args": cmdArgs}), nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	return nil, apperr.New(apperr.CodeNetwork, "aria2c daemon start timed out waiting for rpc readiness")
}

func probeVersion(ctx context.Context, client *aria2rpc.Client) (map[string]any, string, error) {
	res, err := client.Call(ctx, "aria2.getVersion", []any{})
	if err == nil {
		out := map[string]any{}
		if decodeErr := json.Unmarshal(res, &out); decodeErr != nil {
			return map[string]any{"raw": string(res)}, "rpc_ready", nil
		}
		return out, "rpc_ready", nil
	}
	typed, ok := err.(*apperr.Error)
	if !ok {
		return nil, "unknown_error", err
	}
	if typed.Code == apperr.CodeRPC {
		return nil, "rpc_reachable_auth_or_rpc_error", err
	}
	if typed.Code == apperr.CodeNetwork {
		return nil, "rpc_unreachable", err
	}
	return nil, "unknown_error", err
}

func isLocalHost(host string) bool {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return true
	}
	parsed := net.ParseIP(h)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback()
}

func runAddURI(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	uri, err := RequiredString(req.Args, "uri")
	if err != nil {
		return nil, err
	}
	options, _ := OptionalJSONArray(req.Args, "options")
	position, err := OptionalInt(req.Args, "position", -1)
	if err != nil {
		return nil, err
	}
	params := []any{[]string{uri}}
	if len(options) > 0 {
		params = append(params, options[0])
	}
	if position >= 0 {
		params = append(params, position)
	}
	res, err := client.Call(ctx, "aria2.addUri", params)
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runAddMagnet(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	uri, err := RequiredString(req.Args, "uri")
	if err != nil {
		return nil, err
	}
	req.Args["uri"] = uri
	return runAddURI(ctx, client, req)
}

func runAddTorrent(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	path, err := RequiredString(req.Args, "file")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to read torrent file", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	res, err := client.Call(ctx, "aria2.addTorrent", []any{encoded})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), map[string]any{"source_file": path}), nil
}

func runAddMetalink(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	path, err := RequiredString(req.Args, "file")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to read metalink file", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	res, err := client.Call(ctx, "aria2.addMetalink", []any{encoded})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), map[string]any{"source_file": path}), nil
}

func runPause(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	force := OptionalBool(req.Args, "force", false)
	method := "aria2.pause"
	if force {
		method = "aria2.forcePause"
	}
	res, err := client.Call(ctx, method, []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runPauseAll(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	method := "aria2.pauseAll"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forcePauseAll"
	}
	res, err := client.Call(ctx, method, []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runResume(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.unpause", []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runResumeAll(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.unpauseAll", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runRemove(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	method := "aria2.remove"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forceRemove"
	}
	res, err := client.Call(ctx, method, []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runRemoveAll(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	active, err := client.Call(ctx, "aria2.tellActive", []any{})
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(active, &items); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to decode active list", err)
	}
	removed := make([]string, 0, len(items))
	method := "aria2.remove"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forceRemove"
	}
	for _, item := range items {
		gid := strings.TrimSpace(stringValue(item["gid"]))
		if gid == "" {
			continue
		}
		if _, err := client.Call(ctx, method, []any{gid}); err == nil {
			removed = append(removed, gid)
		}
	}
	result := map[string]any{"removed_gids": removed, "count": len(removed)}
	return envelope(req, result, nil), nil
}

func runPurgeDownloadResult(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.purgeDownloadResult", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runTellStatus(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellStatus", []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runListActive(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.tellActive", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runListWaiting(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	offset, err := OptionalInt(req.Args, "offset", 0)
	if err != nil {
		return nil, err
	}
	limit, err := OptionalInt(req.Args, "limit", 100)
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellWaiting", []any{offset, limit})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runListStopped(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	offset, err := OptionalInt(req.Args, "offset", 0)
	if err != nil {
		return nil, err
	}
	limit, err := OptionalInt(req.Args, "limit", 100)
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellStopped", []any{offset, limit})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runGetGlobalStat(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.getGlobalStat", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runRPCCall(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	method, err := RequiredString(req.Args, "method")
	if err != nil {
		return nil, err
	}
	params, err := OptionalJSONArray(req.Args, "params")
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func envelope(req Request, result any, diagnostics map[string]any) map[string]any {
	if diagnostics == nil {
		diagnostics = map[string]any{}
	}
	return map[string]any{
		"ok":          true,
		"request":     map[string]any{"capability": req.Name, "args": req.Args},
		"result":      result,
		"diagnostics": diagnostics,
	}
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.TrimSpace(toJSONScalar(v))), "\"", ""))
}

func toJSONScalar(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
