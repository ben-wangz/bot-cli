package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const (
	FormatJSON  = "json"
	FormatYAML  = "yaml"
	FormatTable = "table"
)

func ValidateFormat(format string) error {
	if format == FormatJSON || format == FormatYAML || format == FormatTable {
		return nil
	}
	return apperr.New(apperr.CodeInvalidArgs, "output must be json, yaml, or table")
}

func Render(w io.Writer, format string, payload any) error {
	switch format {
	case FormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	case FormatYAML:
		_, err := io.WriteString(w, toYAML(payload, 0))
		return err
	case FormatTable:
		_, err := io.WriteString(w, toTable(payload))
		return err
	default:
		return ValidateFormat(format)
	}
}

func toTable(payload any) string {
	if payload == nil {
		return "\n"
	}
	v := reflect.ValueOf(payload)
	if v.Kind() == reflect.Map {
		keys := mapKeysSorted(v)
		rows := make([]string, 0, len(keys)+1)
		rows = append(rows, "KEY	VALUE")
		for _, key := range keys {
			value := v.MapIndex(reflect.ValueOf(key))
			rows = append(rows, fmt.Sprintf("%s	%v", key, value.Interface()))
		}
		return strings.Join(rows, "\n") + "\n"
	}
	if v.Kind() == reflect.Slice {
		rows := []string{"INDEX	VALUE"}
		for i := 0; i < v.Len(); i++ {
			rows = append(rows, fmt.Sprintf("%d	%v", i, v.Index(i).Interface()))
		}
		return strings.Join(rows, "\n") + "\n"
	}
	return fmt.Sprintf("VALUE\n%v\n", payload)
}

func mapKeysSorted(v reflect.Value) []string {
	keys := make([]string, 0, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		keys = append(keys, fmt.Sprintf("%v", iter.Key().Interface()))
	}
	sort.Strings(keys)
	return keys
}

func toYAML(payload any, indent int) string {
	spacing := strings.Repeat("  ", indent)
	if payload == nil {
		return spacing + "null\n"
	}

	v := reflect.ValueOf(payload)
	if !v.IsValid() {
		return spacing + "null\n"
	}

	switch v.Kind() {
	case reflect.Map:
		keys := mapKeysSorted(v)
		lines := make([]string, 0, len(keys))
		for _, key := range keys {
			mapValue := v.MapIndex(reflect.ValueOf(key)).Interface()
			if isScalar(mapValue) {
				lines = append(lines, fmt.Sprintf("%s%s: %s", spacing, key, scalarValue(mapValue)))
				continue
			}
			lines = append(lines, fmt.Sprintf("%s%s:\n%s", spacing, key, toYAML(mapValue, indent+1)))
		}
		if len(lines) == 0 {
			return spacing + "{}\n"
		}
		return strings.Join(lines, "\n") + "\n"
	case reflect.Slice:
		if v.Len() == 0 {
			return spacing + "[]\n"
		}
		lines := make([]string, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			if isScalar(item) {
				lines = append(lines, fmt.Sprintf("%s- %s", spacing, scalarValue(item)))
				continue
			}
			lines = append(lines, fmt.Sprintf("%s-\n%s", spacing, toYAML(item, indent+1)))
		}
		return strings.Join(lines, "\n") + "\n"
	default:
		return spacing + scalarValue(payload) + "\n"
	}
}

func isScalar(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.Kind() != reflect.Map && v.Kind() != reflect.Slice
}

func scalarValue(value any) string {
	if value == nil {
		return "null"
	}
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return `""`
		}
		if strings.ContainsAny(typed, ":#{}[]\n	") || strings.HasPrefix(typed, " ") || strings.HasSuffix(typed, " ") {
			return fmt.Sprintf("%q", typed)
		}
		return typed
	default:
		return fmt.Sprintf("%v", value)
	}
}
