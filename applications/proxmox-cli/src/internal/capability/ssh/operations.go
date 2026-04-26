package sshcap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func CheckService(ctx context.Context, req Request) (map[string]any, error) {
	t, err := parseTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	command := append(buildSSHBaseArgs(t, true), t.User+"@"+t.Host, "--", "true")
	stdout, stderr, exitCode, execErr := runCommandWithOutput(ctx, "ssh", command...)
	latency := time.Since(start).Milliseconds()
	reachable, authOK := analyzeProbe(exitCode, stderr)
	if execErr != nil && !reachable {
		return nil, apperr.Wrap(apperr.CodeNetwork, "ssh service probe failed", execErr)
	}
	result := map[string]any{"reachable": reachable, "auth_ok": authOK, "latency_ms": latency, "stderr_tail": tailText(stderr, 240), "stdout_tail": tailText(stdout, 120), "exit_code": exitCode}
	request := map[string]any{"host": t.Host, "port": t.Port, "user": t.User}
	if t.IdentityFile != "" {
		request["identity_file"] = t.IdentityFile
	}
	return buildResult(req, request, result, map[string]any{"connect_timeout_seconds": t.ConnectTimeoutSeconds}), nil
}

func Exec(ctx context.Context, req Request) (map[string]any, error) {
	t, err := parseTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	command, err := requiredString(req.Args, "command")
	if err != nil {
		return nil, err
	}
	timeoutSeconds := 30
	if raw := strings.TrimSpace(req.Args["timeout-seconds"]); raw != "" {
		v, parseErr := strconv.Atoi(raw)
		if parseErr != nil || v <= 0 {
			return nil, apperr.New(apperr.CodeInvalidArgs, "timeout-seconds must be a positive integer")
		}
		timeoutSeconds = v
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	start := time.Now()
	sshArgs := append(buildSSHBaseArgs(t, true), t.User+"@"+t.Host, "--", command)
	stdout, stderr, exitCode, execErr := runCommandWithOutput(execCtx, "ssh", sshArgs...)
	duration := time.Since(start).Milliseconds()
	if execErr != nil {
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			return nil, apperr.Wrap(apperr.CodeNetwork, "ssh command timed out", execErr)
		}
		if exitCode == -1 {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to execute ssh command", execErr)
		}
	}
	request := map[string]any{"host": t.Host, "port": t.Port, "user": t.User, "command": command, "timeout_seconds": timeoutSeconds}
	if t.IdentityFile != "" {
		request["identity_file"] = t.IdentityFile
	}
	result := map[string]any{"exit_code": exitCode, "stdout": stdout, "stderr": stderr, "duration_ms": duration}
	return buildResult(req, request, result, map[string]any{"timed_out": errors.Is(execCtx.Err(), context.DeadlineExceeded)}), nil
}

func ScpTransfer(ctx context.Context, req Request) (map[string]any, error) {
	t, err := parseTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	direction, err := requiredString(req.Args, "direction")
	if err != nil {
		return nil, err
	}
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "upload" && direction != "download" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "direction must be one of upload|download")
	}
	localPath, err := requiredString(req.Args, "local-path")
	if err != nil {
		return nil, err
	}
	remotePath, err := requiredString(req.Args, "remote-path")
	if err != nil {
		return nil, err
	}
	recursive, err := parseOptionalBoolArg(req.Args, "recursive")
	if err != nil {
		return nil, err
	}
	if direction == "upload" {
		if _, statErr := os.Stat(localPath); statErr != nil {
			return nil, apperr.Wrap(apperr.CodeConfig, "local-path is not readable", statErr)
		}
	} else {
		if mkErr := os.MkdirAll(filepath.Dir(localPath), 0o755); mkErr != nil {
			return nil, apperr.Wrap(apperr.CodeConfig, "failed to prepare local-path parent directory", mkErr)
		}
	}
	start := time.Now()
	scpArgs := buildScpBaseArgs(t, recursive)
	remoteSpec := t.User + "@" + t.Host + ":" + remotePath
	if direction == "upload" {
		scpArgs = append(scpArgs, localPath, remoteSpec)
	} else {
		scpArgs = append(scpArgs, remoteSpec, localPath)
	}
	_, stderr, exitCode, execErr := runCommandWithOutput(ctx, "scp", scpArgs...)
	duration := time.Since(start).Milliseconds()
	if execErr != nil {
		if exitCode == -1 {
			return nil, apperr.Wrap(apperr.CodeNetwork, "failed to execute scp transfer", execErr)
		}
		return nil, apperr.Wrap(apperr.CodeNetwork, "scp transfer failed", execErr)
	}
	verified := false
	bytesCount := int64(0)
	if direction == "download" {
		if info, statErr := os.Stat(localPath); statErr == nil {
			verified = true
			if !info.IsDir() {
				bytesCount = info.Size()
			}
		}
	} else {
		verified = true
		if info, statErr := os.Stat(localPath); statErr == nil && !info.IsDir() {
			bytesCount = info.Size()
		}
	}
	request := map[string]any{"direction": direction, "host": t.Host, "port": t.Port, "user": t.User, "local_path": localPath, "remote_path": remotePath, "recursive": recursive}
	if t.IdentityFile != "" {
		request["identity_file"] = t.IdentityFile
	}
	result := map[string]any{"direction": direction, "bytes": bytesCount, "duration_ms": duration, "verified_exists": verified, "exit_code": exitCode, "stderr_tail": tailText(stderr, 240)}
	return buildResult(req, request, result, map[string]any{}), nil
}

func PrintConnectCommand(req Request) (map[string]any, error) {
	t, err := parseTarget(req.Args, false)
	if err != nil {
		return nil, err
	}
	args := append(buildSSHBaseArgs(t, false), t.User+"@"+t.Host)
	command := "ssh " + shellJoin(args)
	request := map[string]any{"host": t.Host, "port": t.Port, "user": t.User}
	if t.IdentityFile != "" {
		request["identity_file"] = t.IdentityFile
	}
	return buildResult(req, request, map[string]any{"command": command}, map[string]any{"interactive_supported": false}), nil
}
