package action

import (
	"context"
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
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runSSHCheckService(ctx context.Context, req Request) (map[string]any, error) {
	target, err := parseSSHTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	command := append(buildSSHBaseArgs(target, true), target.User+"@"+target.Host, "--", "true")
	stdout, stderr, exitCode, execErr := runCommandWithOutput(ctx, "ssh", command...)
	latency := time.Since(start).Milliseconds()
	reachable, authOK := analyzeSSHProbe(exitCode, stderr)
	if execErr != nil && !reachable {
		return nil, apperr.Wrap(apperr.CodeNetwork, "ssh service probe failed", execErr)
	}
	result := map[string]any{
		"reachable":   reachable,
		"auth_ok":     authOK,
		"latency_ms":  latency,
		"stderr_tail": tailText(stderr, 240),
		"stdout_tail": tailText(stdout, 120),
		"exit_code":   exitCode,
	}
	request := map[string]any{"host": target.Host, "port": target.Port, "user": target.User}
	if target.IdentityFile != "" {
		request["identity_file"] = target.IdentityFile
	}
	return buildResult(req, request, result, map[string]any{"connect_timeout_seconds": target.ConnectTimeoutSeconds}), nil
}

func runSSHInjectPubKeyQGA(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	username, err := RequiredString(req.Args, "username")
	if err != nil {
		return nil, err
	}
	pubKey, source, err := resolvePublicKey(req.Args)
	if err != nil {
		return nil, err
	}
	fingerprint, err := publicKeyFingerprint(pubKey)
	if err != nil {
		return nil, err
	}
	script := buildInjectPubKeyScript(username, pubKey)
	agentReq := Request{
		Name: "agent_exec",
		Args: map[string]string{
			"node":            node,
			"vmid":            strconv.Itoa(vmid),
			"command":         "true",
			"shell":           "1",
			"script":          script,
			"timeout-seconds": "45",
		},
		Scope: req.Scope,
	}
	agentResult, err := runAgentExec(ctx, client, agentReq)
	if err != nil {
		return nil, err
	}
	exitCode, exited := extractAgentExecExit(agentResult)
	if !exited || exitCode != 0 {
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("qga key injection command failed with exit code %d", exitCode))
	}
	home := "/home/" + username
	if username == "root" {
		home = "/root"
	}
	request := map[string]any{"node": node, "vmid": vmid, "username": username, "key_source": source}
	result := map[string]any{
		"username":             username,
		"authorized_keys_path": filepath.ToSlash(filepath.Join(home, ".ssh", "authorized_keys")),
		"fingerprint":          fingerprint,
		"qga_exit_code":        exitCode,
	}
	return buildResult(req, request, result, map[string]any{"pubkey_source": source}), nil
}

func runSSHExec(ctx context.Context, req Request) (map[string]any, error) {
	target, err := parseSSHTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	command, err := RequiredString(req.Args, "command")
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
	sshArgs := append(buildSSHBaseArgs(target, true), target.User+"@"+target.Host, "--", command)
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
	request := map[string]any{"host": target.Host, "port": target.Port, "user": target.User, "command": command, "timeout_seconds": timeoutSeconds}
	if target.IdentityFile != "" {
		request["identity_file"] = target.IdentityFile
	}
	result := map[string]any{"exit_code": exitCode, "stdout": stdout, "stderr": stderr, "duration_ms": duration}
	return buildResult(req, request, result, map[string]any{"timed_out": errors.Is(execCtx.Err(), context.DeadlineExceeded)}), nil
}

func runSSHScpTransfer(ctx context.Context, req Request) (map[string]any, error) {
	target, err := parseSSHTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	direction, err := RequiredString(req.Args, "direction")
	if err != nil {
		return nil, err
	}
	direction = strings.ToLower(strings.TrimSpace(direction))
	if !isOneOf(direction, "upload", "download") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "direction must be one of upload|download")
	}
	localPath, err := RequiredString(req.Args, "local-path")
	if err != nil {
		return nil, err
	}
	remotePath, err := RequiredString(req.Args, "remote-path")
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
	scpArgs := buildScpBaseArgs(target, recursive)
	remoteSpec := target.User + "@" + target.Host + ":" + remotePath
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
	request := map[string]any{"direction": direction, "host": target.Host, "port": target.Port, "user": target.User, "local_path": localPath, "remote_path": remotePath, "recursive": recursive}
	if target.IdentityFile != "" {
		request["identity_file"] = target.IdentityFile
	}
	result := map[string]any{"direction": direction, "bytes": bytesCount, "duration_ms": duration, "verified_exists": verified, "exit_code": exitCode, "stderr_tail": tailText(stderr, 240)}
	return buildResult(req, request, result, map[string]any{}), nil
}

func runSSHPrintConnectCommand(req Request) (map[string]any, error) {
	target, err := parseSSHTarget(req.Args, false)
	if err != nil {
		return nil, err
	}
	args := append(buildSSHBaseArgs(target, false), target.User+"@"+target.Host)
	command := "ssh " + shellJoin(args)
	request := map[string]any{"host": target.Host, "port": target.Port, "user": target.User}
	if target.IdentityFile != "" {
		request["identity_file"] = target.IdentityFile
	}
	return buildResult(req, request, map[string]any{"command": command}, map[string]any{"interactive_supported": false}), nil
}

func runSSHTunnelStart(ctx context.Context, req Request) (map[string]any, error) {
	target, err := parseSSHTarget(req.Args, true)
	if err != nil {
		return nil, err
	}
	localPort, err := RequiredInt(req.Args, "local-port")
	if err != nil {
		return nil, err
	}
	remoteHost, err := RequiredString(req.Args, "remote-host")
	if err != nil {
		return nil, err
	}
	remotePort, err := RequiredInt(req.Args, "remote-port")
	if err != nil {
		return nil, err
	}
	pidFile, logFile := resolveTunnelPaths(req.Args, target, localPort, remoteHost, remotePort)
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
	commandArgs := append(buildSSHBaseArgs(target, true), "-o", "ExitOnForwardFailure=yes", "-o", "ServerAliveInterval=30", "-o", "ServerAliveCountMax=3", "-N", "-L", forward, target.User+"@"+target.Host)
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
	meta := tunnelMeta{LocalEndpoint: endpoint, LocalPort: localPort, RemoteHost: remoteHost, RemotePort: remotePort, Host: target.Host, Port: target.Port, User: target.User, PID: pid}
	if err := writeTunnelMeta(pidFile, meta); err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}
	time.Sleep(250 * time.Millisecond)
	if !processExists(pid) {
		logTail := readFileTail(logFile, 240)
		return nil, apperr.New(apperr.CodeNetwork, "ssh tunnel exited immediately; log_tail="+logTail)
	}
	request := map[string]any{"host": target.Host, "port": target.Port, "user": target.User, "local_port": localPort, "remote_host": remoteHost, "remote_port": remotePort, "pid_file": pidFile, "log_file": logFile}
	if target.IdentityFile != "" {
		request["identity_file"] = target.IdentityFile
	}
	result := map[string]any{"pid": pid, "pid_file": pidFile, "log_file": logFile, "local_endpoint": endpoint}
	return buildResult(req, request, result, map[string]any{"background": true}), nil
}

func runSSHTunnelStatus(req Request) (map[string]any, error) {
	pidFile, err := RequiredString(req.Args, "pid-file")
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

func runSSHTunnelStop(req Request) (map[string]any, error) {
	pidFile, err := RequiredString(req.Args, "pid-file")
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
