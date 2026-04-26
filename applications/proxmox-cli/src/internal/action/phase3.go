package action

import (
	"context"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase3(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "agent_network_get_interfaces":
		return runAgentNetworkGetInterfaces(ctx, client, req)
	case "agent_exec":
		return runAgentExec(ctx, client, req)
	case "agent_exec_status":
		return runAgentExecStatus(ctx, client, req)
	case "storage_upload_guard":
		return runStorageUploadGuard(ctx, client, req)
	case "storage_upload_iso":
		return runStorageUploadISO(ctx, client, req)
	case "build_ubuntu_autoinstall_iso":
		return runBuildUbuntuAutoinstallISO(req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 3: "+req.Name)
	}
}
