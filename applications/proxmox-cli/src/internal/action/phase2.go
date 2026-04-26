package action

import (
	"context"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase2(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "clone_template":
		return runCloneTemplate(ctx, client, req)
	case "migrate_vm":
		return runMigrateVM(ctx, client, req)
	case "convert_vm_to_template":
		return runConvertVMToTemplate(ctx, client, req)
	case "update_vm_config":
		return runUpdateVMConfig(ctx, client, req)
	case "vm_power":
		return runVMPower(ctx, client, req)
	case "set_vm_agent":
		return runSetVMAgent(ctx, client, req)
	case "create_vm":
		return runCreateVM(ctx, client, req)
	case "attach_cdrom_iso":
		return runAttachCDROMISO(ctx, client, req)
	case "set_net_boot_config":
		return runSetNetBootConfig(ctx, client, req)
	case "enable_serial_console":
		return runEnableSerialConsole(ctx, client, req)
	case "review_install_tasks":
		return runReviewInstallTasks(ctx, client, req)
	case "sendkey":
		return runSendKey(ctx, client, req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 2: "+req.Name)
	}
}
