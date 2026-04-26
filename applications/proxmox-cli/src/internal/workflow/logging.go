package workflow

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func LogStep(writer io.Writer, stage string, name string, args map[string]string) {
	if writer == nil {
		return
	}
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := args[key]
		if IsSensitiveArg(key) {
			value = "***REDACTED***"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	_, _ = fmt.Fprintf(writer, "[workflow] %s step=%s args={%s}\n", stage, name, strings.Join(parts, " "))
}

func LogStepError(writer io.Writer, name string, err error) {
	if writer == nil {
		return
	}
	_, _ = fmt.Fprintf(writer, "[workflow] fail step=%s err=%s\n", name, err.Error())
}
