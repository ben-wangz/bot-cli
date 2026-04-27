package sshcap

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

type target struct {
	Host                  string
	Port                  int
	User                  string
	IdentityFile          string
	ConnectTimeoutSeconds int
	ExtraArgs             []string
}

func parseTarget(args map[string]string, batch bool, capabilityName string) (target, error) {
	if err := rejectUnsupportedPassword(args, capabilityName); err != nil {
		return target{}, err
	}
	host, err := requiredString(args, "host")
	if err != nil {
		return target{}, err
	}
	user, err := requiredString(args, "user")
	if err != nil {
		return target{}, err
	}
	port := 22
	if rawPort := strings.TrimSpace(args["port"]); rawPort != "" {
		v, parseErr := strconv.Atoi(rawPort)
		if parseErr != nil || v <= 0 || v > 65535 {
			return target{}, apperr.New(apperr.CodeInvalidArgs, "port must be an integer in range 1-65535")
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
			return target{}, apperr.New(apperr.CodeInvalidArgs, "connect-timeout-seconds must be a positive integer")
		}
		connectTimeout = v
	}
	identityFile := strings.TrimSpace(args["identity-file"])
	if identityFile != "" {
		if _, statErr := os.Stat(identityFile); statErr != nil {
			return target{}, apperr.Wrap(apperr.CodeConfig, "identity-file is not readable", statErr)
		}
	}
	extra := strings.TrimSpace(args["extra-args"])
	extraArgs := []string{}
	if extra != "" {
		extraArgs = strings.Fields(extra)
	}
	return target{Host: host, Port: port, User: user, IdentityFile: identityFile, ConnectTimeoutSeconds: connectTimeout, ExtraArgs: extraArgs}, nil
}

func rejectUnsupportedPassword(args map[string]string, capabilityName string) error {
	if _, provided := args["password"]; !provided {
		return nil
	}
	name := strings.TrimSpace(capabilityName)
	if name == "" {
		name = "ssh capability"
	}
	return apperr.New(apperr.CodeInvalidArgs, "--password is not supported for "+name+" in batch/key mode; use --identity-file")
}

func buildSSHBaseArgs(t target, batch bool) []string {
	args := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
	if batch {
		args = append(args, "-o", "BatchMode=yes")
	}
	if t.ConnectTimeoutSeconds > 0 {
		args = append(args, "-o", "ConnectTimeout="+strconv.Itoa(t.ConnectTimeoutSeconds))
	}
	if t.IdentityFile != "" {
		args = append(args, "-i", t.IdentityFile)
	}
	args = append(args, "-p", strconv.Itoa(t.Port))
	args = append(args, t.ExtraArgs...)
	return args
}

func buildScpBaseArgs(t target, recursive bool) []string {
	args := []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-o", "BatchMode=yes", "-o", "ConnectTimeout=" + strconv.Itoa(t.ConnectTimeoutSeconds)}
	if t.IdentityFile != "" {
		args = append(args, "-i", t.IdentityFile)
	}
	args = append(args, "-P", strconv.Itoa(t.Port))
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, t.ExtraArgs...)
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

func analyzeProbe(exitCode int, stderr string) (bool, bool) {
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

func ResolvePublicKey(args map[string]string) (string, string, error) {
	fromArg := strings.TrimSpace(args["pub-key"])
	fromFile := strings.TrimSpace(args["pub-key-file"])
	if fromArg == "" && fromFile == "" {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --pub-key-file or --pub-key")
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

func PublicKeyFingerprint(pubKey string) (string, error) {
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

func BuildInjectPubKeyScript(username string, pubKey string) string {
	userQuoted := shellQuote(username)
	keyQuoted := shellQuote(pubKey)
	return strings.Join([]string{
		"set -eu",
		"USER_NAME=" + userQuoted,
		"PASSWD_LINE=$(getent passwd \"$USER_NAME\" || true)",
		"if [ -z \"$PASSWD_LINE\" ]; then echo \"user not found: $USER_NAME\" >&2; exit 2; fi",
		"HOME_DIR=$(printf '%s' \"$PASSWD_LINE\" | cut -d: -f6)",
		"if [ -z \"$HOME_DIR\" ]; then echo \"home directory is empty for user: $USER_NAME\" >&2; exit 3; fi",
		"if [ \"$USER_NAME\" = \"root\" ]; then HOME_DIR=/root; fi",
		"USER_GROUP=$(id -gn \"$USER_NAME\")",
		"mkdir -p \"$HOME_DIR/.ssh\"",
		"chmod 700 \"$HOME_DIR/.ssh\"",
		"touch \"$HOME_DIR/.ssh/authorized_keys\"",
		"chmod 600 \"$HOME_DIR/.ssh/authorized_keys\"",
		"if ! grep -qxF -- " + keyQuoted + " \"$HOME_DIR/.ssh/authorized_keys\"; then printf '%s\\n' " + keyQuoted + " >> \"$HOME_DIR/.ssh/authorized_keys\"; fi",
		"chown -R \"$USER_NAME\":\"$USER_GROUP\" \"$HOME_DIR/.ssh\"",
	}, "\n")
}

func ExtractAgentExecExit(result map[string]any) (int, bool) {
	res, ok := result["result"].(map[string]any)
	if !ok {
		return -1, false
	}
	status, ok := res["status"].(map[string]any)
	if !ok {
		return -1, false
	}
	exited, _ := status["exited"].(bool)
	if !exited {
		switch v := status["exited"].(type) {
		case float64:
			exited = int(v) != 0
		case int:
			exited = v != 0
		case int64:
			exited = v != 0
		}
	}
	if !exited {
		if s, ok := status["exited"].(string); ok {
			exited = strings.EqualFold(strings.TrimSpace(s), "true") || strings.TrimSpace(s) == "1"
		}
	}
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
	case uint64:
		return int(v), true
	case uint:
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

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
