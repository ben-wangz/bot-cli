package capability

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
)

func ParseArgs(args []string) (map[string]string, error) {
	result := map[string]string{}
	for i := 0; i < len(args); i++ {
		name, inline, hasInline, err := splitArg(args[i])
		if err != nil {
			return nil, err
		}
		if hasInline {
			result[name] = mergeArgValue(result[name], inline)
			continue
		}
		if i+1 >= len(args) {
			return nil, apperr.New(apperr.CodeInvalidArgs, "missing value for capability arg --"+name)
		}
		result[name] = mergeArgValue(result[name], args[i+1])
		i++
	}
	return result, nil
}

func mergeArgValue(existing, next string) string {
	next = strings.TrimSpace(next)
	if existing == "" {
		return next
	}
	if next == "" {
		return existing
	}
	return existing + "\n" + next
}

func splitArg(token string) (string, string, bool, error) {
	if !strings.HasPrefix(token, "--") {
		return "", "", false, apperr.New(apperr.CodeInvalidArgs, "capability args must be --key value or --key=value")
	}
	trimmed := strings.TrimPrefix(token, "--")
	if trimmed == "" {
		return "", "", false, apperr.New(apperr.CodeInvalidArgs, "invalid empty capability arg")
	}
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true, nil
	}
	return trimmed, "", false, nil
}

func RequiredString(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --"+key)
	}
	return v, nil
}

func OptionalInt(args map[string]string, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be integer")
	}
	return v, nil
}

func OptionalBool(args map[string]string, key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(args[key]))
	if raw == "" {
		return fallback
	}
	return raw == "1" || raw == "true" || raw == "yes"
}

func OptionalJSONArray(args map[string]string, key string) ([]any, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return []any{}, nil
	}
	var out []any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, apperr.New(apperr.CodeInvalidArgs, key+" must be a json array")
	}
	return out, nil
}

func OptionalJSONObject(args map[string]string, key string) (map[string]any, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, apperr.New(apperr.CodeInvalidArgs, key+" must be a json object")
	}
	return out, nil
}

func OptionalKeyValueList(args map[string]string, key string) ([]string, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return []string{}, nil
	}
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out, nil
}
