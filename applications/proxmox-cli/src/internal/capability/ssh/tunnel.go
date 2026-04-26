package sshcap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

type tunnelMeta struct {
	LocalEndpoint string `json:"local_endpoint"`
	LocalPort     int    `json:"local_port"`
	RemoteHost    string `json:"remote_host"`
	RemotePort    int    `json:"remote_port"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	User          string `json:"user"`
	PID           int    `json:"pid"`
}

func TunnelStart(ctx context.Context, req Request) (map[string]any, error) {
	t, err := parseTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	localPort, err := requiredInt(req.Args, "local-port")
	if err != nil {
		return nil, err
	}
	remoteHost, err := requiredString(req.Args, "remote-host")
	if err != nil {
		return nil, err
	}
	remotePort, err := requiredInt(req.Args, "remote-port")
	if err != nil {
		return nil, err
	}
	pidFile, logFile := resolveTunnelPaths(req.Args, t, localPort, remoteHost, remotePort)
	if err := os.MkdirAll(filepath.Dir(pidFile), 0o755); err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create tunnel pid directory", err)
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create tunnel log directory", err)
	}
	if pid, ok := readPIDFile(pidFile); ok && processExists(pid) {
		return nil, apperr.New(apperr.CodeInvalidArgs, fmt.Sprintf("tunnel already running with pid %d", pid))
	}
	logHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to open tunnel log file", err)
	}
	defer logHandle.Close()

	forward := fmt.Sprintf("%d:%s:%d", localPort, remoteHost, remotePort)
	commandArgs := append(buildSSHBaseArgs(t, true), "-o", "ExitOnForwardFailure=yes", "-o", "ServerAliveInterval=30", "-o", "ServerAliveCountMax=3", "-N", "-L", forward, t.User+"@"+t.Host)
	cmd := exec.CommandContext(ctx, "ssh", commandArgs...)
	cmd.Stdout = logHandle
	cmd.Stderr = logHandle
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, apperr.Wrap(apperr.CodeNetwork, "failed to start ssh tunnel", err)
	}
	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0o644); err != nil {
		_ = cmd.Process.Kill()
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to write tunnel pid file", err)
	}
	endpoint := fmt.Sprintf("127.0.0.1:%d -> %s:%d", localPort, remoteHost, remotePort)
	meta := tunnelMeta{LocalEndpoint: endpoint, LocalPort: localPort, RemoteHost: remoteHost, RemotePort: remotePort, Host: t.Host, Port: t.Port, User: t.User, PID: pid}
	if err := writeTunnelMeta(pidFile, meta); err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}
	time.Sleep(250 * time.Millisecond)
	if !processExists(pid) {
		logTail := readFileTail(logFile, 240)
		return nil, apperr.New(apperr.CodeNetwork, "ssh tunnel exited immediately; log_tail="+logTail)
	}
	request := map[string]any{"host": t.Host, "port": t.Port, "user": t.User, "local_port": localPort, "remote_host": remoteHost, "remote_port": remotePort, "pid_file": pidFile, "log_file": logFile}
	if t.IdentityFile != "" {
		request["identity_file"] = t.IdentityFile
	}
	result := map[string]any{"pid": pid, "pid_file": pidFile, "log_file": logFile, "local_endpoint": endpoint}
	return buildResult(req, request, result, map[string]any{"background": true}), nil
}

func TunnelStatus(req Request) (map[string]any, error) {
	pidFile, err := requiredString(req.Args, "pid-file")
	if err != nil {
		return nil, err
	}
	pid, ok := readPIDFile(pidFile)
	if !ok {
		result := map[string]any{"running": false, "pid": 0, "pid_file": pidFile, "last_error": "pid file missing or invalid"}
		return buildResult(req, map[string]any{"pid_file": pidFile}, result, map[string]any{"pid_file_present": false}), nil
	}
	running := processExists(pid)
	meta, _ := readTunnelMeta(pidFile)
	endpoint := meta.LocalEndpoint
	lastErr := ""
	if !running {
		lastErr = "process not running"
	}
	result := map[string]any{"running": running, "pid": pid, "pid_file": pidFile, "local_endpoint": endpoint, "last_error": lastErr}
	return buildResult(req, map[string]any{"pid_file": pidFile}, result, map[string]any{"pid_file_present": true}), nil
}

func TunnelStop(req Request) (map[string]any, error) {
	pidFile, err := requiredString(req.Args, "pid-file")
	if err != nil {
		return nil, err
	}
	pid, ok := readPIDFile(pidFile)
	cleanup := map[string]any{"removed_pid_file": false, "removed_meta_file": false, "forced_kill": false}
	if !ok {
		cleanupFiles(pidFile, cleanup)
		result := map[string]any{"stopped": true, "pid": 0, "pid_file": pidFile, "cleanup": cleanup}
		return buildResult(req, map[string]any{"pid_file": pidFile}, result, map[string]any{"already_stopped": true}), nil
	}
	stopped := true
	if processExists(pid) {
		proc, findErr := os.FindProcess(pid)
		if findErr != nil {
			return nil, apperr.Wrap(apperr.CodeConfig, "failed to resolve tunnel process", findErr)
		}
		if sigErr := proc.Signal(syscall.SIGTERM); sigErr != nil && !errors.Is(sigErr, os.ErrProcessDone) {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to signal tunnel process", sigErr)
		}
		if !waitProcessExit(pid, 5*time.Second) {
			cleanup["forced_kill"] = true
			_ = proc.Signal(syscall.SIGKILL)
			if !waitProcessExit(pid, 2*time.Second) {
				stopped = false
			}
		}
	}
	cleanupFiles(pidFile, cleanup)
	result := map[string]any{"stopped": stopped, "pid": pid, "pid_file": pidFile, "cleanup": cleanup}
	if !stopped {
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("failed to stop tunnel process pid=%d", pid))
	}
	return buildResult(req, map[string]any{"pid_file": pidFile}, result, map[string]any{}), nil
}

func readPIDFile(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, convErr := strconv.Atoi(strings.TrimSpace(string(data)))
	if convErr != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	return !errors.Is(err, syscall.ESRCH)
}

func waitProcessExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return !processExists(pid)
}

func sanitizeName(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	builder := strings.Builder{}
	for _, r := range v {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteByte('-')
	}
	name := strings.Trim(builder.String(), "-")
	if name == "" {
		return "tunnel"
	}
	return name
}

func writeTunnelMeta(pidFile string, meta tunnelMeta) error {
	metaPath := pidFile + ".meta.json"
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to encode tunnel metadata", err)
	}
	if err := os.WriteFile(metaPath, append(data, '\n'), 0o644); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to write tunnel metadata", err)
	}
	return nil
}

func readTunnelMeta(pidFile string) (tunnelMeta, bool) {
	metaPath := pidFile + ".meta.json"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return tunnelMeta{}, false
	}
	var meta tunnelMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return tunnelMeta{}, false
	}
	return meta, true
}

func cleanupFiles(pidFile string, cleanup map[string]any) {
	if err := os.Remove(pidFile); err == nil || errors.Is(err, os.ErrNotExist) {
		cleanup["removed_pid_file"] = true
	}
	metaPath := pidFile + ".meta.json"
	if err := os.Remove(metaPath); err == nil || errors.Is(err, os.ErrNotExist) {
		cleanup["removed_meta_file"] = true
	}
}

func readFileTail(path string, max int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(data))
	return tailText(text, max)
}
