package vmcap

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/policy"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

type Request struct {
	Name  string
	Args  map[string]string
	Scope string
}

func buildResult(req Request, request map[string]any, result any, diagnostics map[string]any) map[string]any {
	return map[string]any{"capability": req.Name, "ok": true, "scope": req.Scope, "request": request, "result": result, "diagnostics": diagnostics}
}

func writeResult(req Request, request map[string]any, data any) map[string]any {
	return buildResult(req, request, map[string]any{"upid": asString(data)}, map[string]any{})
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

func formKeys(form url.Values) []string {
	keys := make([]string, 0, len(form))
	for key := range form {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func setIfPresent(form url.Values, key string, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}
	form.Set(key, trimmed)
}

func mergeDiagnosticsMap(result map[string]any, extra map[string]any) {
	if result == nil || len(extra) == 0 {
		return
	}
	diagnostics, _ := result["diagnostics"].(map[string]any)
	if diagnostics == nil {
		diagnostics = map[string]any{}
	}
	for key, value := range extra {
		diagnostics[key] = value
	}
	result["diagnostics"] = diagnostics
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

func RequiredVMID(args map[string]string) (int, error) {
	vmidRaw := strings.TrimSpace(args["vmid"])
	if vmidRaw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --vmid")
	}
	vmid, err := strconv.Atoi(vmidRaw)
	if err != nil || vmid <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, "vmid must be a positive integer")
	}
	return vmid, nil
}

func RequiredOperationVMID(args map[string]string) (int, error) {
	vmid, err := RequiredVMID(args)
	if err != nil {
		return 0, err
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

func RequiredOperationInt(args map[string]string, key string) (int, error) {
	v, err := RequiredInt(args, key)
	if err != nil {
		return 0, err
	}
	if err := policy.EnsureOperationVMID(v); err != nil {
		return 0, err
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

func vmExistsOnNode(ctx context.Context, client *pveapi.Client, node string, vmid int) (bool, map[string]any, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		message := strings.ToLower(strings.TrimSpace(err.Error()))
		if strings.Contains(message, "404") || strings.Contains(message, "not found") || isMissingVMConfigError(message) {
			return false, nil, nil
		}
		return false, nil, err
	}
	statusMap, _ := data.(map[string]any)
	return true, statusMap, nil
}

func isMissingVMConfigError(message string) bool {
	if !strings.Contains(message, "configuration file") {
		return false
	}
	if !strings.Contains(message, "qemu-server") {
		return false
	}
	return strings.Contains(message, "does not exist")
}

func isVMRunning(ctx context.Context, client *pveapi.Client, node string, vmid int) (bool, error) {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return false, err
	}
	obj, _ := data.(map[string]any)
	if obj == nil {
		return false, apperr.New(apperr.CodeNetwork, "vm status response is not an object")
	}
	status := strings.TrimSpace(asString(obj["status"]))
	return status == "running", nil
}
