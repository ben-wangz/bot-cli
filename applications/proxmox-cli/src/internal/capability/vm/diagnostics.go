package vmcap

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func RunReviewInstallTasks(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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
	query.Set("source", "active")
	path := fmt.Sprintf("/nodes/%s/tasks", url.PathEscape(node))
	data, err := client.GetData(ctx, path, query)
	if err != nil {
		return nil, err
	}
	count := 0
	if list, ok := data.([]any); ok {
		count = len(list)
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid}, data, map[string]any{"active_task_count": count}), nil
}

func RunSendKey(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	key, err := RequiredString(req.Args, "key")
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("key", key)
	setIfPresent(form, "skiplock", req.Args["skiplock"])
	path := fmt.Sprintf("/nodes/%s/qemu/%d/sendkey", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	result := writeResult(req, map[string]any{"node": node, "vmid": vmid, "key": key}, data)
	result["diagnostics"] = map[string]any{"method": "PUT"}
	return result, nil
}
