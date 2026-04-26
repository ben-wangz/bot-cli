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
