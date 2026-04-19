package redact

import (
	"regexp"
	"strings"
)

var keyValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(token\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(cookie\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(ticket\s*[=:]\s*)([^\s,;]+)`),
}

var jsonPattern = regexp.MustCompile(`(?i)("(?:password|token|cookie|ticket)"\s*:\s*")([^"]+)(")`)

func String(input string) string {
	if input == "" {
		return ""
	}
	redacted := input
	for _, pattern := range keyValuePatterns {
		redacted = pattern.ReplaceAllString(redacted, `${1}***REDACTED***`)
	}
	redacted = jsonPattern.ReplaceAllString(redacted, `${1}***REDACTED***${3}`)
	return redacted
}

func Headers(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(headers))
	for key, value := range headers {
		if isSecretHeader(key) {
			result[key] = "***REDACTED***"
			continue
		}
		result[key] = value
	}
	return result
}

func isSecretHeader(name string) bool {
	lower := strings.ToLower(name)
	if lower == "authorization" || lower == "cookie" {
		return true
	}
	return strings.Contains(lower, "token") || strings.Contains(lower, "ticket") || strings.Contains(lower, "password")
}
