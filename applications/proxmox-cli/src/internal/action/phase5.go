package action

import (
	"context"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase5(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "create_pve_user_with_root":
		return runCreatePVEUserWithRoot(ctx, client, req)
	case "create_pool_with_root":
		return runCreatePoolWithRoot(ctx, client, req)
	case "get_user_acl_binding":
		return runGetUserACLBinding(ctx, client, req)
	case "grant_user_acl":
		return runGrantUserACL(ctx, client, req)
	case "revoke_user_acl":
		return runRevokeUserACL(ctx, client, req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 5: "+req.Name)
	}
}
