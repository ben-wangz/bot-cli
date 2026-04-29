package capability

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
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
		version, probeState, callErr := probeVersion(ctx, client)
		if callErr == nil {
			return envelope(req, map[string]any{
				"already_running": false,
				"started":         true,
				"version":         version,
			}, map[string]any{"probe_state": probeState, "start_args": cmdArgs}), nil
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
