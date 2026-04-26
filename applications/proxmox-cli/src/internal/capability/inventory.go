package capability

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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
	data, err := client.GetData(ctx, "/cluster/nextid", url.Values{})
	if err != nil {
		return nil, err
	}
	nextID := parseNumeric(data)
	return buildResult(req, map[string]any{}, map[string]any{"next_vmid": nextID}, map[string]any{}), nil
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
