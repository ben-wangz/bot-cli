package capability

import (
	"context"

	storagecap "github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability/storage"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runStorageUploadGuard(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return storagecap.RunStorageUploadGuard(ctx, client, storagecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runStorageUploadISO(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return storagecap.RunStorageUploadISO(ctx, client, storagecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runBuildUbuntuAutoinstallISO(req Request) (map[string]any, error) {
	return storagecap.RunBuildUbuntuAutoinstallISO(storagecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}
