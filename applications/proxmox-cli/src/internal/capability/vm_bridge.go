package capability

import (
	"context"

	vmcap "github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability/vm"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runCloneTemplate(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunCloneTemplate(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runMigrateVM(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunMigrateVM(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runConvertVMToTemplate(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunConvertVMToTemplate(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runUpdateVMConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunUpdateVMConfig(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runVMPower(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunVMPower(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runDestroyVM(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunDestroyVM(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSetVMAgent(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunSetVMAgent(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runCreateVM(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunCreateVM(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runAttachCDROMISO(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunAttachCDROMISO(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSetNetBootConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunSetNetBootConfig(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runEnableSerialConsole(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunEnableSerialConsole(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runReviewInstallTasks(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunReviewInstallTasks(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSendKey(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	return vmcap.RunSendKey(ctx, client, vmcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}
