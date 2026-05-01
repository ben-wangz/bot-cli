package capability

import (
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
)

func ParseArgs(args []string) (map[string]string, error) {
	result := map[string]string{}
	for i := 0; i < len(args); i++ {
		name, inline, hasInline, err := splitArg(args[i])
		if err != nil {
			return nil, err
		}
		if hasInline {
			result[name] = strings.TrimSpace(inline)
			continue
		}
		if i+1 >= len(args) {
			return nil, apperr.New(apperr.CodeInvalidArgs, "missing value for capability arg --"+name)
		}
		result[name] = strings.TrimSpace(args[i+1])
		i++
	}
	return result, nil
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

func OptionalBool(args map[string]string, key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(args[key]))
	if raw == "" {
		return fallback
	}
	return raw == "1" || raw == "true" || raw == "yes"
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
