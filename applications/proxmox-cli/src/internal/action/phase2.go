package action

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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
	case "start_installer_and_console_ticket":
		return runStartInstallerAndConsoleTicket(ctx, client, req)
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

func runCloneTemplate(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	sourceVMID, err := RequiredOperationInt(req.Args, "source-vmid")
	if err != nil {
		return nil, err
	}
	targetVMID, err := RequiredOperationInt(req.Args, "target-vmid")
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("newid", strconv.Itoa(targetVMID))
	setIfPresent(form, "name", req.Args["name"])
	setIfPresent(form, "target", req.Args["target"])
	setIfPresent(form, "full", req.Args["full"])
	path := fmt.Sprintf("/nodes/%s/qemu/%d/clone", url.PathEscape(node), sourceVMID)
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "source_vmid": sourceVMID, "target_vmid": targetVMID}, data), nil
}

func runMigrateVM(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	target, err := RequiredString(req.Args, "target")
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("target", target)
	setIfPresent(form, "online", req.Args["online"])
	setIfPresent(form, "with-local-disks", req.Args["with-local-disks"])
	path := fmt.Sprintf("/nodes/%s/qemu/%d/migrate", url.PathEscape(node), vmid)
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid, "target": target}, data), nil
}

func runConvertVMToTemplate(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/template", url.PathEscape(node), vmid)
	data, err := client.PostFormData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid}, data), nil
}

func runUpdateVMConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	form := mapArgsToForm(req.Args, "node", "vmid")
	if len(form) == 0 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "update_vm_config requires at least one config key")
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid, "changed_keys": formKeys(form)}, data), nil
}

func runVMPower(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	mode, err := RequiredString(req.Args, "mode")
	if err != nil {
		return nil, err
	}
	if !isOneOf(mode, "start", "stop", "shutdown", "reboot", "reset") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "mode must be one of start|stop|shutdown|reboot|reset")
	}
	form := mapArgsToForm(req.Args, "node", "vmid", "mode")
	path := fmt.Sprintf("/nodes/%s/qemu/%d/status/%s", url.PathEscape(node), vmid, url.PathEscape(mode))
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid, "mode": mode}, data), nil
}

func runSetVMAgent(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	enabled := strings.TrimSpace(req.Args["enabled"])
	if enabled == "" {
		enabled = "1"
	}
	if !isOneOf(enabled, "0", "1") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "enabled must be 0 or 1")
	}
	form := url.Values{}
	form.Set("agent", enabled)
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid, "enabled": enabled == "1"}, data), nil
}

func runCreateVM(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	form := mapArgsToForm(req.Args, "node")
	form.Set("vmid", strconv.Itoa(vmid))
	if len(form) == 1 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "create_vm requires vm config args like --name --memory --cores")
	}
	path := fmt.Sprintf("/nodes/%s/qemu", url.PathEscape(node))
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid}, data), nil
}

func runAttachCDROMISO(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	iso, err := RequiredString(req.Args, "iso")
	if err != nil {
		return nil, err
	}
	ideSlot := strings.TrimSpace(req.Args["slot"])
	if ideSlot == "" {
		ideSlot = "ide2"
	}
	form := url.Values{}
	form.Set(ideSlot, iso)
	setIfPresent(form, "media", req.Args["media"])
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid, "slot": ideSlot, "iso": iso}, data), nil
}

func runSetNetBootConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	net0, err := RequiredString(req.Args, "net0")
	if err != nil {
		return nil, err
	}
	boot, err := RequiredString(req.Args, "boot")
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("net0", net0)
	form.Set("boot", boot)
	setIfPresent(form, "bootdisk", req.Args["bootdisk"])
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid, "net0": net0, "boot": boot}, data), nil
}

func runStartInstallerAndConsoleTicket(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	startPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", url.PathEscape(node), vmid)
	startData, err := client.PostFormData(ctx, startPath, url.Values{})
	if err != nil {
		return nil, err
	}
	vncPath := fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy", url.PathEscape(node), vmid)
	vncData, vncErr := client.PostFormData(ctx, vncPath, url.Values{})
	diagnostics := map[string]any{}
	if vncErr != nil {
		diagnostics["vncproxy_warning"] = vncErr.Error()
	}
	return map[string]any{
		"action":  req.Name,
		"ok":      true,
		"scope":   req.Scope,
		"request": map[string]any{"node": node, "vmid": vmid},
		"result": map[string]any{
			"upid":       asString(startData),
			"vnc_ticket": vncData,
		},
		"diagnostics": diagnostics,
	}, nil
}

func runEnableSerialConsole(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("serial0", "socket")
	form.Set("vga", "serial0")
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	return writeResult(req, map[string]any{"node": node, "vmid": vmid}, data), nil
}

func runReviewInstallTasks(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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

func runSendKey(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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

func writeResult(req Request, request map[string]any, data any) map[string]any {
	return buildResult(req, request, map[string]any{"upid": asString(data)}, map[string]any{})
}

func mapArgsToForm(args map[string]string, excluded ...string) url.Values {
	skips := map[string]bool{}
	for _, key := range excluded {
		skips[key] = true
	}
	form := url.Values{}
	for key, value := range args {
		if skips[key] || strings.TrimSpace(value) == "" {
			continue
		}
		form.Set(key, value)
	}
	return form
}

func formKeys(form url.Values) []string {
	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	return keys
}

func setIfPresent(form url.Values, key string, value string) {
	v := strings.TrimSpace(value)
	if v != "" {
		form.Set(key, v)
	}
}

func isOneOf(v string, values ...string) bool {
	for _, each := range values {
		if v == each {
			return true
		}
	}
	return false
}
