package workflow

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/policy"
)

func EnsureAllowedArgs(args map[string]string, allowed ...string) error {
	allowedSet := map[string]bool{}
	for _, key := range allowed {
		allowedSet[key] = true
	}
	for key := range args {
		if !allowedSet[key] {
			return apperr.New(apperr.CodeInvalidArgs, "unsupported workflow arg --"+key)
		}
	}
	return nil
}

func RequiredString(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required workflow arg --"+key)
	}
	return v, nil
}

func ParseVMID(args map[string]string, key string, required bool) (int, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		if required {
			return 0, apperr.New(apperr.CodeInvalidArgs, "missing required workflow arg --"+key)
		}
		return 0, nil
	}
	vmid, err := strconv.Atoi(raw)
	if err != nil || vmid <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	if err := policy.EnsureOperationVMID(vmid); err != nil {
		return 0, err
	}
	return vmid, nil
}

func RequiredPositiveInt(args map[string]string, key string) (int, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required workflow arg --"+key)
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	return v, nil
}

func OptionalPositiveInt(args map[string]string, key string) (int, bool, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return 0, false, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, false, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	return v, true, nil
}

func ParseArtifactISOVolID(volid string) (string, string, error) {
	trimmed := strings.TrimSpace(volid)
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "artifact-iso must be in <storage>:iso/<file>.iso format")
	}
	storage := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])
	if storage == "" || path == "" || !strings.HasPrefix(path, "iso/") || !strings.HasSuffix(strings.ToLower(path), ".iso") {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "artifact-iso must be in <storage>:iso/<file>.iso format")
	}
	return storage, trimmed, nil
}

func ResolveProvisionSerialLogPath(targetVMID int) (string, error) {
	workspaceRoot, err := ResolveWorkspaceRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(workspaceRoot, "build", fmt.Sprintf("serial-provision-template-%d.log", targetVMID)), nil
}

func ResolveWorkspaceRoot() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", apperr.Wrap(apperr.CodeConfig, "failed to resolve working directory", err)
	}
	current := filepath.Clean(start)
	for {
		if isDir(filepath.Join(current, ".git")) {
			if isDir(filepath.Join(current, "applications", "proxmox-cli", "src")) {
				return current, nil
			}
		}
		if current == filepath.Dir(current) {
			break
		}
		current = filepath.Dir(current)
	}
	return start, nil
}

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func WrapStepError(step string, vmid int, err error) error {
	message := fmt.Sprintf("workflow step failed: %s; cleanup may be required for vmid=%d", step, vmid)
	return apperr.Wrap(apperr.CodeNetwork, message, err)
}

func AnyToInt(v any) (int, bool) {
	switch value := v.(type) {
	case int:
		return value, true
	case int64:
		return int(value), true
	case float64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func AnyToBool(v any) bool {
	switch value := v.(type) {
	case bool:
		return value
	case int:
		return value != 0
	case int64:
		return value != 0
	case float64:
		return value != 0
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(value))
		return trimmed == "1" || trimmed == "true" || trimmed == "yes"
	default:
		return false
	}
}

func IsSensitiveArg(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	if lower == "" {
		return false
	}
	sensitive := []string{"password", "token", "ticket", "csrf", "secret"}
	for _, marker := range sensitive {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func StepSummary(stepName string, payload map[string]any) map[string]any {
	result, _ := payload["result"]
	return map[string]any{
		"step":   stepName,
		"ok":     payload["ok"],
		"result": result,
	}
}

func GeneratePassword(length int) (string, error) {
	if length <= 0 {
		return "", apperr.New(apperr.CodeInvalidArgs, "password length must be positive")
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, length)
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to generate workflow password", err)
	}
	for i := 0; i < length; i++ {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
