package action

import (
	"context"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase1(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "list_nodes":
		return runListNodes(ctx, client, req)
	case "list_cluster_resources":
		return runListClusterResources(ctx, client, req)
	case "list_vms_by_node":
		return runListVMsByNode(ctx, client, req)
	case "get_vm_config":
		return runGetVMConfig(ctx, client, req)
	case "get_effective_permissions":
		return runGetEffectivePermissions(ctx, client, req)
	case "get_task_status":
		return runGetTaskStatus(ctx, client, req)
	case "get_next_vmid":
		return runGetNextVMID(ctx, client, req)
	case "get_vm_status":
		return runGetVMStatus(ctx, client, req)
	case "list_tasks_by_vmid":
		return runListTasksByVMID(ctx, client, req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 1: "+req.Name)
	}
}
