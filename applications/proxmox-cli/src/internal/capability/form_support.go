package capability

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func writeResult(req Request, request map[string]any, data any) map[string]any {
	return buildResult(req, request, map[string]any{"upid": asString(data)}, map[string]any{})
}

func mapArgsToForm(args map[string]string, excluded ...string) url.Values {
	skips := map[string]bool{}
	for _, key := range excluded {
		skips[key] = true
	}
	form := url.Values{}
	for key, value := range args {
		if skips[key] || strings.TrimSpace(value) == "" {
			continue
		}
		form.Set(key, value)
	}
	return form
}

func formKeys(form url.Values) []string {
	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	return keys
}

func setIfPresent(form url.Values, key string, value string) {
	v := strings.TrimSpace(value)
	if v != "" {
		form.Set(key, v)
	}
}

func mergeDiagnosticsMap(result map[string]any, extra map[string]any) {
	diagnostics := map[string]any{}
	if current, ok := result["diagnostics"].(map[string]any); ok {
		for k, v := range current {
			diagnostics[k] = v
		}
	}
	for k, v := range extra {
		diagnostics[k] = v
	}
	result["diagnostics"] = diagnostics
}

func vmExistsOnNode(ctx context.Context, client *pveapi.Client, node string, vmid int) (bool, map[string]any, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "does not exist") || strings.Contains(message, "status 404") || strings.Contains(message, "not found") {
			return false, nil, nil
		}
		return false, nil, err
	}
	status := firstObject(unwrapResultField(data))
	if status == nil {
		return true, map[string]any{}, nil
	}
	return true, status, nil
}

func isOneOf(v string, values ...string) bool {
	for _, each := range values {
		if v == each {
			return true
		}
	}
	return false
}
