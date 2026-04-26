package action

import (
	"context"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase4(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "start_vnc_proxy":
		return runStartVNCProxy(ctx, client, req)
	case "connect_vnc_websocket":
		return runConnectVNCWebsocket(ctx, client, req)
	case "open_vm_termproxy":
		return runOpenVMTermproxy(ctx, client, req)
	case "validate_k1_serial_readable":
		return runValidateK1SerialReadable(ctx, client, req)
	case "serial_ws_session_control":
		return runSerialWSSessionControl(ctx, client, req)
	case "validate_serial_output_criterion2":
		return runValidateSerialOutputCriterion2(ctx, client, req)
	case "serial_ws_capture_to_file":
		return runSerialWSCaptureToFile(ctx, client, req)
	case "ssh_check_service":
		return runSSHCheckService(ctx, req)
	case "ssh_inject_pubkey_qga":
		return runSSHInjectPubKeyQGA(ctx, client, req)
	case "ssh_exec":
		return runSSHExec(ctx, req)
	case "ssh_scp_transfer":
		return runSSHScpTransfer(ctx, req)
	case "ssh_print_connect_command":
		return runSSHPrintConnectCommand(req)
	case "ssh_tunnel_start":
		return runSSHTunnelStart(ctx, req)
	case "ssh_tunnel_status":
		return runSSHTunnelStatus(req)
	case "ssh_tunnel_stop":
		return runSSHTunnelStop(req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 4: "+req.Name)
	}
}
