package action

import (
	"fmt"
	"strconv"
	"strings"
)

type Request struct {
	Name  string
	Args  map[string]string
	Scope string
}

func buildResult(req Request, request map[string]any, result any, diagnostics map[string]any) map[string]any {
	return map[string]any{
		"action":      req.Name,
		"ok":          true,
		"scope":       req.Scope,
		"request":     request,
		"result":      result,
		"diagnostics": diagnostics,
	}
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func parseNumeric(v any) int {
	s := asString(v)
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}
