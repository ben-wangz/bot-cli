package storagecap

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

type Request struct {
	Name  string
	Args  map[string]string
	Scope string
}

func buildResult(req Request, request map[string]any, result any, diagnostics map[string]any) map[string]any {
	return map[string]any{"action": req.Name, "ok": true, "scope": req.Scope, "request": request, "result": result, "diagnostics": diagnostics}
}

func RequiredNode(args map[string]string) (string, error) {
	node := strings.TrimSpace(args["node"])
	if node == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --node")
	}
	for _, r := range node {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", apperr.New(apperr.CodeInvalidArgs, "node contains invalid character")
	}
	return node, nil
}

func RequiredString(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --"+key)
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

func parseUPIDNode(upid string) string {
	parts := strings.Split(strings.TrimSpace(upid), ":")
	if len(parts) < 2 {
		return ""
	}
	if strings.ToUpper(parts[0]) != "UPID" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func tailText(raw string, max int) string {
	if max <= 0 || len(raw) <= max {
		return raw
	}
	return raw[len(raw)-max:]
}

func stringifyNumeric(v any) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(asString(v)))
	if err != nil {
		return 0
	}
	return parsed
}

func pathEscapeNodeStorage(node string, storage string) string {
	return fmt.Sprintf("/nodes/%s/storage/%s", url.PathEscape(node), url.PathEscape(storage))
}
