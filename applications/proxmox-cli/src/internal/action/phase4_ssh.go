package action

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

type sshTarget struct {
	Host                  string
	Port                  int
	User                  string
	IdentityFile          string
	ConnectTimeoutSeconds int
	ExtraArgs             []string
}

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

func parseSSHTarget(args map[string]string, batch bool) (sshTarget, error) {
	host, err := RequiredString(args, "host")
	if err != nil {
		return sshTarget{}, err
	}
	user, err := RequiredString(args, "user")
	if err != nil {
		return sshTarget{}, err
	}
	port := 22
	if rawPort := strings.TrimSpace(args["port"]); rawPort != "" {
		v, parseErr := strconv.Atoi(rawPort)
		if parseErr != nil || v <= 0 || v > 65535 {
			return sshTarget{}, apperr.New(apperr.CodeInvalidArgs, "port must be an integer in range 1-65535")
		}
		port = v
	}
	connectTimeout := 10
	if batch {
		connectTimeout = 5
	}
	if rawTimeout := strings.TrimSpace(args["connect-timeout-seconds"]); rawTimeout != "" {
		v, parseErr := strconv.Atoi(rawTimeout)
		if parseErr != nil || v <= 0 {
			return sshTarget{}, apperr.New(apperr.CodeInvalidArgs, "connect-timeout-seconds must be a positive integer")
		}
		connectTimeout = v
	}
	identityFile := strings.TrimSpace(args["identity-file"])
	if identityFile != "" {
		if _, statErr := os.Stat(identityFile); statErr != nil {
			return sshTarget{}, apperr.Wrap(apperr.CodeConfig, "identity-file is not readable", statErr)
		}
	}
	extra := strings.TrimSpace(args["extra-args"])
	extraArgs := []string{}
	if extra != "" {
		extraArgs = strings.Fields(extra)
	}
	return sshTarget{Host: host, Port: port, User: user, IdentityFile: identityFile, ConnectTimeoutSeconds: connectTimeout, ExtraArgs: extraArgs}, nil
}

func buildSSHBaseArgs(target sshTarget, batch bool) []string {
	args := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
	if batch {
		args = append(args, "-o", "BatchMode=yes")
	}
	if target.ConnectTimeoutSeconds > 0 {
		args = append(args, "-o", "ConnectTimeout="+strconv.Itoa(target.ConnectTimeoutSeconds))
	}
	if target.IdentityFile != "" {
		args = append(args, "-i", target.IdentityFile)
	}
	args = append(args, "-p", strconv.Itoa(target.Port))
	args = append(args, target.ExtraArgs...)
	return args
}

func buildScpBaseArgs(target sshTarget, recursive bool) []string {
	args := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-o", "BatchMode=yes", "-o", "ConnectTimeout=" + strconv.Itoa(target.ConnectTimeoutSeconds)}
	if target.IdentityFile != "" {
		args = append(args, "-i", target.IdentityFile)
	}
	args = append(args, "-P", strconv.Itoa(target.Port))
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, target.ExtraArgs...)
	return args
}

func runCommandWithOutput(ctx context.Context, name string, args ...string) (string, string, int, error) {
	bin, err := exec.LookPath(name)
	if err != nil {
		return "", "", -1, apperr.Wrap(apperr.CodeConfig, "required command not found: "+name, err)
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), exitCode, err
}

func analyzeSSHProbe(exitCode int, stderr string) (bool, bool) {
	if exitCode == 0 {
		return true, true
	}
	lower := strings.ToLower(stderr)
	if strings.Contains(lower, "permission denied") || strings.Contains(lower, "publickey") || strings.Contains(lower, "authentication failed") {
		return true, false
	}
	if strings.Contains(lower, "connection refused") || strings.Contains(lower, "operation timed out") || strings.Contains(lower, "no route to host") || strings.Contains(lower, "could not resolve hostname") {
		return false, false
	}
	return exitCode == 255, false
}

func resolvePublicKey(args map[string]string) (string, string, error) {
	fromArg := strings.TrimSpace(args["pub-key"])
	fromFile := strings.TrimSpace(args["pub-key-file"])
	if fromArg == "" && fromFile == "" {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --pub-key-file or --pub-key")
	}
	if fromArg != "" && fromFile != "" {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "--pub-key-file and --pub-key are mutually exclusive")
	}
	if fromFile != "" {
		data, err := os.ReadFile(fromFile)
		if err != nil {
			return "", "", apperr.Wrap(apperr.CodeConfig, "failed to read pub-key-file", err)
		}
		key := strings.TrimSpace(string(data))
		if key == "" {
			return "", "", apperr.New(apperr.CodeInvalidArgs, "pub-key-file is empty")
		}
		return key, "file", nil
	}
	return fromArg, "inline", nil
}

func publicKeyFingerprint(pubKey string) (string, error) {
	parts := strings.Fields(strings.TrimSpace(pubKey))
	if len(parts) < 2 {
		return "", apperr.New(apperr.CodeInvalidArgs, "pub key format is invalid")
	}
	blob, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", apperr.Wrap(apperr.CodeInvalidArgs, "failed to decode pub key", err)
	}
	sum := sha256.Sum256(blob)
	encoded := strings.TrimRight(base64.StdEncoding.EncodeToString(sum[:]), "=")
	return "SHA256:" + encoded, nil
}

func buildInjectPubKeyScript(username string, pubKey string) string {
	userQuoted := shellQuote(username)
	keyQuoted := shellQuote(pubKey)
	return strings.Join([]string{
		"set -eu",
		"USER_NAME=" + userQuoted,
		"HOME_DIR=$(getent passwd \"$USER_NAME\" | cut -d: -f6 || true)",
		"if [ -z \"$HOME_DIR\" ]; then HOME_DIR=\"/home/$USER_NAME\"; fi",
		"if [ \"$USER_NAME\" = \"root\" ]; then HOME_DIR=/root; fi",
		"mkdir -p \"$HOME_DIR/.ssh\"",
		"chmod 700 \"$HOME_DIR/.ssh\"",
		"touch \"$HOME_DIR/.ssh/authorized_keys\"",
		"chmod 600 \"$HOME_DIR/.ssh/authorized_keys\"",
		"if ! grep -qxF -- " + keyQuoted + " \"$HOME_DIR/.ssh/authorized_keys\"; then printf '%s\\n' " + keyQuoted + " >> \"$HOME_DIR/.ssh/authorized_keys\"; fi",
		"chown -R \"$USER_NAME\":\"$USER_NAME\" \"$HOME_DIR/.ssh\" || true",
	}, "\n")
}

func extractAgentExecExit(result map[string]any) (int, bool) {
	res, ok := result["result"].(map[string]any)
	if !ok {
		return -1, false
	}
	status, ok := res["status"].(map[string]any)
	if !ok {
		return -1, false
	}
	exited := toBool(status["exited"])
	if !exited {
		return -1, false
	}
	switch v := status["exitcode"].(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return -1, true
		}
		return n, true
	default:
		return -1, true
	}
}

func resolveTunnelPaths(args map[string]string, target sshTarget, localPort int, remoteHost string, remotePort int) (string, string) {
	pidFile := strings.TrimSpace(args["pid-file"])
	logFile := strings.TrimSpace(args["log-file"])
	name := fmt.Sprintf("%s-%d-%s-%d", sanitizeName(target.Host), localPort, sanitizeName(remoteHost), remotePort)
	if pidFile == "" {
		pidFile = filepath.Join("build", "ssh-tunnels", name+".pid")
	}
	if logFile == "" {
		logFile = filepath.Join("build", "ssh-tunnels", name+".log")
	}
	return pidFile, logFile
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

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func shellJoin(parts []string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellQuote(part))
	}
	return strings.Join(quoted, " ")
}
