package sshcap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

type Request struct {
	Name  string
	Args  map[string]string
	Scope string
}

func buildResult(req Request, request map[string]any, result any, diagnostics map[string]any) map[string]any {
	return map[string]any{
		"capability":  req.Name,
		"ok":          true,
		"scope":       req.Scope,
		"request":     request,
		"result":      result,
		"diagnostics": diagnostics,
	}
}

func requiredString(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --"+key)
	}
	return v, nil
}

func requiredInt(args map[string]string, key string) (int, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --"+key)
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	return v, nil
}

func parseOptionalBoolArg(args map[string]string, key string) (bool, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return false, nil
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, apperr.New(apperr.CodeInvalidArgs, key+" must be one of 1|0|true|false")
	}
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func tailText(raw string, max int) string {
	if max <= 0 || len(raw) <= max {
		return raw
	}
	return raw[len(raw)-max:]
}
