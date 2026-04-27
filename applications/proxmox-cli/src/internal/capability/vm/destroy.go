package vmcap

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func RunDestroyVM(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	ifMissing := strings.TrimSpace(strings.ToLower(req.Args["if-missing"]))
	if ifMissing == "" {
		ifMissing = "ok"
	}
	if !isOneOf(ifMissing, "ok", "fail") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "if-missing must be one of ok|fail")
	}
	exists, _, existsErr := vmExistsOnNode(ctx, client, node, vmid)
	if existsErr != nil {
		return nil, existsErr
	}
	request := map[string]any{"node": node, "vmid": vmid, "if_missing": ifMissing}
	if !exists {
		if ifMissing == "ok" {
			result := map[string]any{"node": node, "vmid": vmid, "destroyed": false, "missing": true}
			diagnostics := map[string]any{"wait_skipped": "vm does not exist", "if_missing": ifMissing}
			return buildResult(req, request, result, diagnostics), nil
		}
		return nil, apperr.New(apperr.CodeInvalidArgs, fmt.Sprintf("vm %d does not exist on node %s", vmid, node))
	}
	query := url.Values{}
	if err := setOptional01(query, req.Args, "purge"); err != nil {
		return nil, err
	}
	if err := setOptional01(query, req.Args, "destroy-unreferenced-disks"); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d", url.PathEscape(node), vmid)
	data, err := client.DeleteData(ctx, path, query)
	if err != nil {
		if ifMissing == "ok" && isNotFoundVMError(err) {
			result := map[string]any{"node": node, "vmid": vmid, "destroyed": false, "missing": true}
			diagnostics := map[string]any{"wait_skipped": "vm does not exist", "if_missing": ifMissing}
			return buildResult(req, request, result, diagnostics), nil
		}
		return nil, err
	}
	if purge := strings.TrimSpace(req.Args["purge"]); purge != "" {
		request["purge"] = purge == "1"
	}
	if destroyDisks := strings.TrimSpace(req.Args["destroy-unreferenced-disks"]); destroyDisks != "" {
		request["destroy_unreferenced_disks"] = destroyDisks == "1"
	}
	result := writeResult(req, request, data)
	mergeDiagnosticsMap(result, map[string]any{"if_missing": ifMissing})
	return result, nil
}

func setOptional01(form url.Values, args map[string]string, key string) error {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return nil
	}
	if !isOneOf(raw, "0", "1") {
		return apperr.New(apperr.CodeInvalidArgs, key+" must be 0 or 1")
	}
	form.Set(key, raw)
	return nil
}
