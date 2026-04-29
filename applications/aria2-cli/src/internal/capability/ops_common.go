package capability

import (
	"encoding/json"
	"strings"
)

func envelope(req Request, result any, diagnostics map[string]any) map[string]any {
	if diagnostics == nil {
		diagnostics = map[string]any{}
	}
	return map[string]any{
		"ok":          true,
		"request":     map[string]any{"capability": req.Name, "args": req.Args},
		"result":      result,
		"diagnostics": diagnostics,
	}
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.TrimSpace(toJSONScalar(v))), "\"", ""))
}

func toJSONScalar(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
