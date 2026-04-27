package capability

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/policy"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runListNodes(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	data, err := client.GetData(ctx, "/nodes", url.Values{})
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{}, data, map[string]any{}), nil
}

func runListClusterResources(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	query := url.Values{}
	if t := strings.TrimSpace(req.Args["type"]); t != "" {
		query.Set("type", t)
	}
	data, err := client.GetData(ctx, "/cluster/resources", query)
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"type": req.Args["type"]}, data, map[string]any{}), nil
}

func runListVMsByNode(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/qemu", url.PathEscape(node))
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"node": node}, data, map[string]any{}), nil
}

func runGetVMConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid}, data, map[string]any{}), nil
}

func runGetEffectivePermissions(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	pathArg := req.Args["path"]
	if strings.TrimSpace(pathArg) == "" {
		pathArg = "/"
	}
	query := url.Values{}
	query.Set("path", pathArg)
	data, err := client.GetData(ctx, "/access/permissions", query)
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"path": pathArg}, data, map[string]any{}), nil
}

func runGetTaskStatus(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	upid, err := RequiredOperationUPID(req.Args)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/tasks/%s/status", url.PathEscape(node), url.PathEscape(upid))
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	diagnostics := map[string]any{}
	if m, ok := data.(map[string]any); ok {
		diagnostics["status"] = asString(m["status"])
		diagnostics["exitstatus"] = asString(m["exitstatus"])
		diagnostics["upid"] = asString(m["upid"])
	}
	return buildResult(req, map[string]any{"node": node, "upid": upid}, data, diagnostics), nil
}

func runGetNextVMID(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	rangePolicy, err := policy.OperationVMIDRange()
	if err != nil {
		return nil, err
	}
	data, err := client.GetData(ctx, "/cluster/nextid", url.Values{})
	if err != nil {
		return nil, err
	}
	clusterNextID := parseNumeric(data)
	usedVMIDs, err := fetchUsedVMIDs(ctx, client)
	if err != nil {
		return nil, err
	}
	nextID, ok := pickNextVMIDInRange(usedVMIDs, clusterNextID, rangePolicy.Min, rangePolicy.Max)
	if !ok {
		return nil, apperr.New(apperr.CodeConfig, "no available vmid inside allowed operation range")
	}
	result := map[string]any{"next_vmid": nextID}
	diagnostics := map[string]any{
		"allowed_range":     fmt.Sprintf("%d..%d", rangePolicy.Min, rangePolicy.Max),
		"cluster_next_vmid": clusterNextID,
		"selected_vmid":     nextID,
	}
	return buildResult(req, map[string]any{}, result, diagnostics), nil
}

func fetchUsedVMIDs(ctx context.Context, client *pveapi.Client) (map[int]struct{}, error) {
	query := url.Values{}
	query.Set("type", "vm")
	data, err := client.GetData(ctx, "/cluster/resources", query)
	if err != nil {
		return nil, err
	}
	used := map[int]struct{}{}
	list, _ := data.([]any)
	for _, item := range list {
		entry, _ := item.(map[string]any)
		if entry == nil {
			continue
		}
		vmid := parseNumeric(entry["vmid"])
		if vmid > 0 {
			used[vmid] = struct{}{}
		}
	}
	return used, nil
}

func pickNextVMIDInRange(used map[int]struct{}, clusterNextID int, minVMID int, maxVMID int) (int, bool) {
	start := clusterNextID
	if start < minVMID || start > maxVMID {
		start = minVMID
	}
	for vmid := start; vmid <= maxVMID; vmid++ {
		if _, exists := used[vmid]; !exists {
			return vmid, true
		}
	}
	for vmid := minVMID; vmid < start; vmid++ {
		if _, exists := used[vmid]; !exists {
			return vmid, true
		}
	}
	return 0, false
}

func runGetVMStatus(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid}, data, map[string]any{}), nil
}

func runListTasksByVMID(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("vmid", strconv.Itoa(vmid))
	if source := strings.TrimSpace(req.Args["source"]); source != "" {
		query.Set("source", source)
	}
	path := fmt.Sprintf("/nodes/%s/tasks", url.PathEscape(node))
	data, err := client.GetData(ctx, path, query)
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid}, data, map[string]any{}), nil
}
