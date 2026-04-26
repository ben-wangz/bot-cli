package consolecap

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/policy"
)

type Request struct {
	Name  string
	Args  map[string]string
	Scope string
}

func buildResult(req Request, request map[string]any, result any, diagnostics map[string]any) map[string]any {
	return map[string]any{"capability": req.Name, "ok": true, "scope": req.Scope, "request": request, "result": result, "diagnostics": diagnostics}
}

func RequiredNode(args map[string]string) (string, error) {
	node := strings.TrimSpace(args["node"])
	if node == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --node")
	}
	for _, r := range node {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", apperr.New(apperr.CodeInvalidArgs, "node contains invalid character")
	}
	return node, nil
}

func RequiredOperationVMID(args map[string]string) (int, error) {
	raw := strings.TrimSpace(args["vmid"])
	if raw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --vmid")
	}
	vmid, err := strconv.Atoi(raw)
	if err != nil || vmid <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, "vmid must be a positive integer")
	}
	if err := policy.EnsureOperationVMID(vmid); err != nil {
		return 0, err
	}
	return vmid, nil
}

func RequiredInt(args map[string]string, key string) (int, error) {
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

func RequiredString(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --"+key)
	}
	return v, nil
}

func isOneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func mapArgsToForm(args map[string]string, excluded ...string) url.Values {
	blocked := map[string]bool{}
	for _, key := range excluded {
		blocked[strings.TrimSpace(key)] = true
	}
	form := url.Values{}
	for key, value := range args {
		if blocked[key] {
			continue
		}
		form.Set(key, value)
	}
	return form
}

func unwrapResultField(data any) any {
	m, ok := data.(map[string]any)
	if !ok {
		return data
	}
	v, ok := m["result"]
	if !ok {
		return data
	}
	return v
}

func firstObject(data any) map[string]any {
	if m, ok := data.(map[string]any); ok {
		return m
	}
	list, ok := data.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, _ := list[0].(map[string]any)
	return m
}
