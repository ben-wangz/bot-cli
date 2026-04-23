package action

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func ParseArgs(args []string) (map[string]string, error) {
	result := map[string]string{}
	for i := 0; i < len(args); i++ {
		name, inline, hasInline, err := splitActionArg(args[i])
		if err != nil {
			return nil, err
		}
		if hasInline {
			result[name] = inline
			continue
		}
		if i+1 >= len(args) {
			return nil, apperr.New(apperr.CodeInvalidArgs, "missing value for action arg --"+name)
		}
		result[name] = args[i+1]
		i++
	}
	return result, nil
}

func splitActionArg(token string) (name string, value string, hasValue bool, err error) {
	if !strings.HasPrefix(token, "--") {
		return "", "", false, apperr.New(apperr.CodeInvalidArgs, "action args must be --key value or --key=value")
	}
	trimmed := strings.TrimPrefix(token, "--")
	if trimmed == "" {
		return "", "", false, apperr.New(apperr.CodeInvalidArgs, "invalid empty action arg")
	}
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true, nil
	}
	return trimmed, "", false, nil
}

func RequiredNode(args map[string]string) (string, error) {
	node := strings.TrimSpace(args["node"])
	if node == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --node")
	}
	for _, r := range node {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", apperr.New(apperr.CodeInvalidArgs, "node contains invalid character")
	}
	return node, nil
}

func RequiredVMID(args map[string]string) (int, error) {
	vmidRaw := strings.TrimSpace(args["vmid"])
	if vmidRaw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required action arg --vmid")
	}
	vmid, err := strconv.Atoi(vmidRaw)
	if err != nil || vmid <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, "vmid must be a positive integer")
	}
	return vmid, nil
}

func RequiredOperationVMID(args map[string]string) (int, error) {
	vmid, err := RequiredVMID(args)
	if err != nil {
		return 0, err
	}
	if err := EnsureOperationVMID(vmid); err != nil {
		return 0, err
	}
	return vmid, nil
}

func RequiredInt(args map[string]string, key string) (int, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required action arg --"+key)
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	return v, nil
}

func RequiredOperationInt(args map[string]string, key string) (int, error) {
	v, err := RequiredInt(args, key)
	if err != nil {
		return 0, err
	}
	if err := EnsureOperationVMID(v); err != nil {
		return 0, err
	}
	return v, nil
}

func RequiredString(args map[string]string, key string) (string, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --"+key)
	}
	return v, nil
}

func RequiredUPID(args map[string]string) (string, error) {
	upid := strings.TrimSpace(args["upid"])
	if upid == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --upid")
	}
	if !strings.HasPrefix(strings.ToUpper(upid), "UPID:") {
		return "", apperr.New(apperr.CodeInvalidArgs, "upid must start with UPID:")
	}
	return upid, nil
}

func RequiredOperationUPID(args map[string]string) (string, error) {
	upid, err := RequiredUPID(args)
	if err != nil {
		return "", err
	}
	parts := strings.Split(upid, ":")
	if len(parts) < 7 {
		return "", apperr.New(apperr.CodeInvalidArgs, "invalid upid format")
	}
	idPart := strings.TrimSpace(parts[6])
	if idPart == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "upid id field is empty")
	}
	vmid, convErr := strconv.Atoi(idPart)
	if convErr != nil {
		return "", apperr.New(apperr.CodeInvalidArgs, "upid must reference numeric vmid id")
	}
	if err := EnsureOperationVMID(vmid); err != nil {
		return "", err
	}
	return upid, nil
}

func IsPhase1Action(name string) bool {
	switch name {
	case "list_nodes", "list_cluster_resources", "list_vms_by_node", "get_vm_config", "get_effective_permissions",
		"get_task_status", "get_next_vmid", "get_vm_status", "list_tasks_by_vmid":
		return true
	default:
		return false
	}
}

func IsPhase2Action(name string) bool {
	switch name {
	case "clone_template", "migrate_vm", "convert_vm_to_template", "update_vm_config", "vm_power",
		"set_vm_agent", "create_vm", "attach_cdrom_iso", "set_net_boot_config",
		"start_installer_and_console_ticket", "enable_serial_console", "review_install_tasks", "sendkey":
		return true
	default:
		return false
	}
}

func IsPhase3Action(name string) bool {
	switch name {
	case "agent_network_get_interfaces", "agent_exec", "agent_exec_status", "dump_cloudinit", "storage_upload_guard", "storage_upload_snippet", "storage_upload_iso", "build_ubuntu_autoinstall_iso", "render_and_serve_seed":
		return true
	default:
		return false
	}
}

func IsPhase4Action(name string) bool {
	switch name {
	case "start_vnc_proxy", "connect_vnc_websocket", "open_vm_termproxy", "validate_k1_serial_readable", "serial_ws_session_control", "validate_serial_output_criterion2", "serial_ws_capture_to_file",
		"ssh_check_service", "ssh_inject_pubkey_qga", "ssh_exec", "ssh_scp_transfer", "ssh_print_connect_command", "ssh_tunnel_start", "ssh_tunnel_status", "ssh_tunnel_stop":
		return true
	default:
		return false
	}
}

func IsPhase5Action(name string) bool {
	switch name {
	case "node_termproxy_shell_exec":
		return true
	default:
		return false
	}
}

func IsActionAsync(name string) bool {
	switch name {
	case "clone_template", "migrate_vm", "convert_vm_to_template", "vm_power", "create_vm", "start_installer_and_console_ticket":
		return true
	default:
		return false
	}
}

func IsPhase2AsyncAction(name string) bool {
	return IsActionAsync(name)
}

func WaitSkipReason(name string) string {
	switch {
	case IsPhase1Action(name):
		return "action is read-only"
	case IsPhase3Action(name):
		return "action is synchronous or self-polled"
	case IsPhase4Action(name), IsPhase5Action(name):
		return "action is synchronous or session-driven"
	default:
		return "action is synchronous"
	}
}
