package capability

import (
	"context"

	consolecap "github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability/console"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runOpenVMTermproxy(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return consolecap.RunOpenVMTermproxy(ctx, client, consolecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runValidateK1SerialReadable(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return consolecap.RunValidateK1SerialReadable(ctx, client, consolecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSerialWSSessionControl(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return consolecap.RunSerialWSSessionControl(ctx, client, consolecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runValidateSerialOutputCriterion2(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return consolecap.RunValidateSerialOutputCriterion2(ctx, client, consolecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSerialWSCaptureToFile(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return consolecap.RunSerialWSCaptureToFile(ctx, client, consolecap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}
