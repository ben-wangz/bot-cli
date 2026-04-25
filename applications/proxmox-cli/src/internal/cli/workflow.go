package cli

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/action"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const (
	workflowBootstrapUserPoolACL      = "bootstrap-bot-user-pool-acl"
	workflowProvisionTemplateArtifact = "provision-template-from-artifact"
	workflowISOStorage                = "local"
	workflowDiskStorage               = "local-lvm"
	workflowDiskSizeGB                = 32
	workflowMemoryMB                  = 8192
	workflowCPUCores                  = 2
	workflowNet0                      = "virtio,bridge=vmbr0"
	workflowBootInstall               = "order=ide2"
	workflowBootTemplate              = "order=scsi0"
)

func executeWorkflow(rt commandRuntime, name string, args map[string]string) (map[string]any, error) {
	switch name {
	case workflowBootstrapUserPoolACL:
		return runWorkflowBootstrapTestUserPoolACL(rt, args)
	case workflowProvisionTemplateArtifact:
		return runWorkflowProvisionTemplateFromArtifact(rt, args)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow not implemented yet: "+name)
	}
}

func runWorkflowProvisionTemplateFromArtifact(rt commandRuntime, args map[string]string) (map[string]any, error) {
	if err := ensureAllowedWorkflowArgs(args, "node", "target-vmid", "artifact-iso", "install-timeout-seconds", "template-name", "memory", "cores", "disk-size-gb", "net0", "install-boot-order", "runtime-boot-order", "bootdisk", "resume-from", "pool"); err != nil {
		return nil, err
	}
	node, err := requiredWorkflowString(args, "node")
	if err != nil {
		return nil, err
	}
	targetVMID, err := parseWorkflowVMID(args, "target-vmid", true)
	if err != nil {
		return nil, err
	}
	artifactISO, err := requiredWorkflowString(args, "artifact-iso")
	if err != nil {
		return nil, err
	}
	storageID, _, err := parseArtifactISOVolID(artifactISO)
	if err != nil {
		return nil, err
	}
	installTimeoutSeconds, err := requiredWorkflowPositiveInt(args, "install-timeout-seconds")
	if err != nil {
		return nil, err
	}
	poolID := strings.TrimSpace(args["pool"])
	resumeFrom := strings.TrimSpace(strings.ToLower(args["resume-from"]))
	if resumeFrom == "" {
		resumeFrom = "none"
	}
	if resumeFrom != "none" && resumeFrom != "serial_wait" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "resume-from must be one of none|serial_wait")
	}
	if resumeFrom == "none" {
		if _, err := runWorkflowAction(rt, "list_nodes", map[string]string{}, false); err != nil {
			return nil, wrapWorkflowStepError("list_nodes", targetVMID, err)
		}
		targetResource, err := getWorkflowTargetResource(rt, targetVMID)
		if err != nil {
			return nil, err
		}
		if targetResource != nil {
			return nil, apperr.New(apperr.CodeInvalidArgs, fmt.Sprintf("target-vmid %d already exists", targetVMID))
		}
	}
	if err := ensureUploadedISOArtifactExists(rt, node, storageID, artifactISO); err != nil {
		return nil, err
	}

	cfg, err := resolveProvisionTemplateConfig(args, targetVMID)
	if err != nil {
		return nil, err
	}
	serialLogPath, err := resolveProvisionSerialLogPath(targetVMID)
	if err != nil {
		return nil, err
	}
	steps := make([]map[string]any, 0, 20)
	if resumeFrom == "none" {
		createVMArgs := map[string]string{
			"node":      node,
			"vmid":      strconv.Itoa(targetVMID),
			"name":      cfg.TemplateName,
			"memory":    strconv.Itoa(cfg.MemoryMB),
			"cores":     strconv.Itoa(cfg.Cores),
			"net0":      cfg.Net0,
			"scsi0":     fmt.Sprintf("%s:%d", workflowDiskStorage, cfg.DiskSizeGB),
			"if-exists": "fail",
		}
		if poolID != "" {
			createVMArgs["pool"] = poolID
		}
		if _, err := runWorkflowAction(rt, "create_vm", createVMArgs, true); err != nil {
			return nil, wrapWorkflowStepError("create_vm", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "create_vm", "action": "create_vm"})
		if _, err := runWorkflowAction(rt, "enable_serial_console", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, true); err != nil {
			return nil, wrapWorkflowStepError("enable_serial_console", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "enable_serial_console", "action": "enable_serial_console"})
		if _, err := runWorkflowAction(rt, "set_vm_agent", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "enabled": "1"}, true); err != nil {
			return nil, wrapWorkflowStepError("set_vm_agent", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "set_vm_agent", "action": "set_vm_agent", "enabled": 1})
		if _, err := runWorkflowAction(rt, "attach_cdrom_iso", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "iso": artifactISO, "slot": "ide2", "media": "cdrom"}, true); err != nil {
			return nil, wrapWorkflowStepError("attach_cdrom_iso(install)", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "attach_cdrom_iso", "action": "attach_cdrom_iso", "iso": artifactISO})
		if _, err := runWorkflowAction(rt, "set_net_boot_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "net0": cfg.Net0, "boot": cfg.InstallBootOrder, "bootdisk": cfg.BootDisk}, true); err != nil {
			return nil, wrapWorkflowStepError("set_net_boot_config(install)", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "set_net_boot_config_install", "action": "set_net_boot_config", "boot": cfg.InstallBootOrder})
		if _, err := runWorkflowAction(rt, "vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "start", "desired-state": "running"}, true); err != nil {
			return nil, wrapWorkflowStepError("vm_power(start install)", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "vm_power_start_install", "action": "vm_power"})
	} else {
		steps = append(steps, map[string]any{"step": "resume_from_serial_wait", "warning": "resume-from serial_wait is not fully validated; use with caution"})
	}

	appendMode := "0"
	if resumeFrom == "serial_wait" {
		appendMode = "1"
	}
	capturePayload, err := runWorkflowAction(rt, "serial_ws_capture_to_file", map[string]string{
		"node":            node,
		"vmid":            strconv.Itoa(targetVMID),
		"log-path":        serialLogPath,
		"append":          appendMode,
		"timeout-seconds": strconv.Itoa(installTimeoutSeconds),
	}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("serial_ws_capture_to_file", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "serial_ws_capture_to_file", "action": "serial_ws_capture_to_file", "log_path": serialLogPath, "timeout_seconds": installTimeoutSeconds, "append": appendMode == "1"})
	captureResult, _ := capturePayload["result"].(map[string]any)
	closedByRemote := workflowAnyToBool(captureResult["closed_by_remote"])
	timedOut := !closedByRemote

	statusPayload, err := runWorkflowAction(rt, "get_vm_status", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("get_vm_status(after capture)", targetVMID, err)
	}
	statusResult, _ := statusPayload["result"].(map[string]any)
	vmStatus := strings.TrimSpace(asStringValue(statusResult["status"]))
	steps = append(steps, map[string]any{"step": "get_vm_status_after_capture", "action": "get_vm_status", "status": vmStatus})
	if timedOut && vmStatus != "stopped" {
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("serial wait timed out and vm did not stop; serial_log_path=%s vm_status=%s next_action_hint=check serial log and current vm status", serialLogPath, vmStatus))
	}

	if _, err := runWorkflowAction(rt, "attach_cdrom_iso", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "iso": "none,media=cdrom", "slot": "ide2"}, true); err != nil {
		return nil, wrapWorkflowStepError("attach_cdrom_iso(detach)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "detach_cdrom_iso", "action": "attach_cdrom_iso"})
	if _, err := runWorkflowAction(rt, "set_net_boot_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "net0": cfg.Net0, "boot": cfg.RuntimeBootOrder, "bootdisk": cfg.BootDisk}, true); err != nil {
		return nil, wrapWorkflowStepError("set_net_boot_config(runtime)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "set_net_boot_config_runtime", "action": "set_net_boot_config", "boot": cfg.RuntimeBootOrder})
	if _, err := runWorkflowAction(rt, "vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "stop", "desired-state": "stopped"}, true); err != nil {
		return nil, wrapWorkflowStepError("vm_power(stop)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "vm_power_stop", "action": "vm_power"})
	if _, err := runWorkflowAction(rt, "convert_vm_to_template", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, true); err != nil {
		return nil, wrapWorkflowStepError("convert_vm_to_template", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "convert_vm_to_template", "action": "convert_vm_to_template"})
	configPayload, err := runWorkflowAction(rt, "get_vm_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("get_vm_config", targetVMID, err)
	}
	configResult, _ := configPayload["result"].(map[string]any)
	steps = append(steps, map[string]any{"step": "get_vm_config", "action": "get_vm_config", "agent": asStringValue(configResult["agent"])})

	return map[string]any{
		"workflow": workflowProvisionTemplateArtifact,
		"ok":       true,
		"scope":    rt.Opts.AuthScope,
		"request": map[string]any{
			"node":                    node,
			"target_vmid":             targetVMID,
			"artifact_iso":            artifactISO,
			"install_timeout_seconds": installTimeoutSeconds,
			"resume_from":             resumeFrom,
			"pool":                    poolID,
		},
		"result": map[string]any{
			"template_vmid":   targetVMID,
			"template_name":   cfg.TemplateName,
			"serial_log_path": serialLogPath,
			"steps":           steps,
		},
		"diagnostics": map[string]any{"step_count": len(steps), "resumed_from": resumeFrom},
	}, nil
}

type provisionTemplateConfig struct {
	TemplateName     string
	MemoryMB         int
	Cores            int
	DiskSizeGB       int
	Net0             string
	InstallBootOrder string
	RuntimeBootOrder string
	BootDisk         string
}

func resolveProvisionTemplateConfig(args map[string]string, targetVMID int) (provisionTemplateConfig, error) {
	cfg := provisionTemplateConfig{
		TemplateName:     fmt.Sprintf("template-%d", targetVMID),
		MemoryMB:         workflowMemoryMB,
		Cores:            workflowCPUCores,
		DiskSizeGB:       workflowDiskSizeGB,
		Net0:             workflowNet0,
		InstallBootOrder: workflowBootInstall,
		RuntimeBootOrder: workflowBootTemplate,
		BootDisk:         "scsi0",
	}
	if v := strings.TrimSpace(args["template-name"]); v != "" {
		cfg.TemplateName = v
	}
	if v, ok, err := optionalWorkflowPositiveInt(args, "memory"); err != nil {
		return provisionTemplateConfig{}, err
	} else if ok {
		cfg.MemoryMB = v
	}
	if v, ok, err := optionalWorkflowPositiveInt(args, "cores"); err != nil {
		return provisionTemplateConfig{}, err
	} else if ok {
		cfg.Cores = v
	}
	if v, ok, err := optionalWorkflowPositiveInt(args, "disk-size-gb"); err != nil {
		return provisionTemplateConfig{}, err
	} else if ok {
		cfg.DiskSizeGB = v
	}
	if v := strings.TrimSpace(args["net0"]); v != "" {
		cfg.Net0 = v
	}
	if v := strings.TrimSpace(args["install-boot-order"]); v != "" {
		cfg.InstallBootOrder = v
	}
	if v := strings.TrimSpace(args["runtime-boot-order"]); v != "" {
		cfg.RuntimeBootOrder = v
	}
	if v := strings.TrimSpace(args["bootdisk"]); v != "" {
		cfg.BootDisk = v
	}
	return cfg, nil
}

func resolveProvisionSerialLogPath(targetVMID int) (string, error) {
	root, err := resolveWorkspaceRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "build", fmt.Sprintf("serial-provision-template-%d.log", targetVMID)), nil
}

func requiredWorkflowPositiveInt(args map[string]string, key string) (int, error) {
	v, err := requiredWorkflowString(args, key)
	if err != nil {
		return 0, err
	}
	n, convErr := strconv.Atoi(v)
	if convErr != nil || n <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	return n, nil
}

func optionalWorkflowPositiveInt(args map[string]string, key string) (int, bool, error) {
	v := strings.TrimSpace(args[key])
	if v == "" {
		return 0, false, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, false, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	return n, true, nil
}

func parseArtifactISOVolID(volid string) (string, string, error) {
	trimmed := strings.TrimSpace(volid)
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "artifact-iso must be in <storage>:iso/<file>.iso format")
	}
	storage := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])
	if storage == "" || path == "" || !strings.HasPrefix(path, "iso/") || !strings.HasSuffix(strings.ToLower(path), ".iso") {
		return "", "", apperr.New(apperr.CodeInvalidArgs, "artifact-iso must be in <storage>:iso/<file>.iso format")
	}
	return storage, trimmed, nil
}

func ensureUploadedISOArtifactExists(rt commandRuntime, node string, storage string, volid string) error {
	query := url.Values{}
	query.Set("content", "iso")
	path := fmt.Sprintf("/nodes/%s/storage/%s/content", url.PathEscape(node), url.PathEscape(storage))
	data, err := rt.Client.GetData(context.Background(), path, query)
	if err != nil {
		return apperr.Wrap(apperr.CodeNetwork, "failed to verify artifact-iso existence", err)
	}
	list, ok := data.([]any)
	if !ok {
		return apperr.New(apperr.CodeNetwork, "unexpected storage content response while verifying artifact-iso")
	}
	for _, item := range list {
		entry, isMap := item.(map[string]any)
		if !isMap {
			continue
		}
		if strings.TrimSpace(asStringValue(entry["volid"])) == volid {
			return nil
		}
	}
	return apperr.New(apperr.CodeInvalidArgs, "artifact-iso not found in target storage: "+volid)
}

func runWorkflowBootstrapTestUserPoolACL(rt commandRuntime, args map[string]string) (map[string]any, error) {
	if err := ensureAllowedWorkflowArgs(args, "userid", "poolid", "password", "pool-comment", "user-comment", "if-exists", "sdn-acl-path", "sdn-role"); err != nil {
		return nil, err
	}
	if rt.Opts.AuthScope != "root" && rt.Opts.AuthScope != "root-token" {
		return nil, apperr.New(apperr.CodeAuth, "workflow bootstrap-bot-user-pool-acl requires --auth-scope root or root-token")
	}
	userID, err := requiredWorkflowString(args, "userid")
	if err != nil {
		return nil, err
	}
	if !strings.Contains(userID, "@") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "userid must include realm suffix, for example user@pve")
	}
	poolID, err := requiredWorkflowString(args, "poolid")
	if err != nil {
		return nil, err
	}
	ifExists := strings.TrimSpace(strings.ToLower(args["if-exists"]))
	if ifExists == "" {
		ifExists = "reuse"
	}
	if ifExists != "fail" && ifExists != "reuse" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "if-exists must be one of fail|reuse")
	}
	sdnACLPath := strings.TrimSpace(args["sdn-acl-path"])
	if sdnACLPath == "" {
		sdnACLPath = "/sdn/zones/localnetwork"
	}
	if !strings.HasPrefix(sdnACLPath, "/") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "sdn-acl-path must start with /")
	}
	sdnRole := strings.TrimSpace(args["sdn-role"])
	if sdnRole == "" {
		sdnRole = "PVEAdmin"
	}
	password := strings.TrimSpace(args["password"])
	passwordGenerated := false
	if password == "" {
		generated, genErr := generateWorkflowPassword(20)
		if genErr != nil {
			return nil, genErr
		}
		password = generated
		passwordGenerated = true
	}

	actions := make([]map[string]any, 0, 8)

	createUserArgs := map[string]string{"userid": userID, "password": password, "if-exists": ifExists}
	if userComment := strings.TrimSpace(args["user-comment"]); userComment != "" {
		createUserArgs["comment"] = userComment
	}
	createUserPayload, err := runWorkflowAction(rt, "create_pve_user_with_root", createUserArgs, false)
	if err != nil {
		return nil, err
	}
	actions = append(actions, workflowActionSummary("create_pve_user_with_root", createUserPayload))

	createPoolArgs := map[string]string{"poolid": poolID, "if-exists": ifExists}
	if poolComment := strings.TrimSpace(args["pool-comment"]); poolComment != "" {
		createPoolArgs["comment"] = poolComment
	}
	createPoolPayload, err := runWorkflowAction(rt, "create_pool_with_root", createPoolArgs, false)
	if err != nil {
		return nil, err
	}
	actions = append(actions, workflowActionSummary("create_pool_with_root", createPoolPayload))

	grants := []map[string]any{
		{"path": "/pool/" + poolID, "role": "PVEAdmin", "propagate": true},
		{"path": "/", "role": "PVEAuditor", "propagate": true},
		{"path": "/storage", "role": "PVEDatastoreAdmin", "propagate": true},
		{"path": sdnACLPath, "role": sdnRole, "propagate": true},
	}
	for _, grant := range grants {
		path := asStringValue(grant["path"])
		role := asStringValue(grant["role"])
		grantPayload, grantErr := runWorkflowAction(rt, "grant_user_acl", map[string]string{
			"userid":    userID,
			"path":      path,
			"role":      role,
			"propagate": "1",
		}, false)
		if grantErr != nil {
			return nil, grantErr
		}
		actions = append(actions, workflowActionSummary("grant_user_acl", grantPayload))
	}

	bindingPayload, err := runWorkflowAction(rt, "get_user_acl_binding", map[string]string{"userid": userID}, false)
	if err != nil {
		return nil, err
	}
	actions = append(actions, workflowActionSummary("get_user_acl_binding", bindingPayload))
	bindingResult, _ := bindingPayload["result"].(map[string]any)
	bindingRows := 0
	if count, ok := bindingResult["count"].(int); ok {
		bindingRows = count
	} else if n, ok := workflowAnyToInt(bindingResult["count"]); ok {
		bindingRows = n
	}

	result := map[string]any{
		"userid":             userID,
		"poolid":             poolID,
		"password":           password,
		"password_generated": passwordGenerated,
		"grants":             grants,
		"actions":            actions,
	}
	request := map[string]any{"userid": userID, "poolid": poolID, "if_exists": ifExists}
	if strings.TrimSpace(args["user-comment"]) != "" {
		request["user_comment"] = strings.TrimSpace(args["user-comment"])
	}
	if strings.TrimSpace(args["pool-comment"]) != "" {
		request["pool_comment"] = strings.TrimSpace(args["pool-comment"])
	}
	request["sdn_acl_path"] = sdnACLPath
	request["sdn_role"] = sdnRole
	diagnostics := map[string]any{"action_count": len(actions), "acl_bindings_total": bindingRows}

	return map[string]any{
		"workflow":    workflowBootstrapUserPoolACL,
		"ok":          true,
		"scope":       rt.Opts.AuthScope,
		"request":     request,
		"result":      result,
		"diagnostics": diagnostics,
	}, nil
}

func runWorkflowAction(rt commandRuntime, name string, actionArgs map[string]string, wait bool) (map[string]any, error) {
	logWorkflowAction(rt, "start", name, actionArgs)
	stepRT := rt
	stepRT.Opts.Wait = wait
	args := make([]string, 0, 1+len(actionArgs)*2)
	args = append(args, name)
	keys := make([]string, 0, len(actionArgs))
	for key := range actionArgs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "--"+key, actionArgs[key])
	}
	payload, err := runActionCommand(stepRT, args)
	if err != nil {
		logWorkflowActionError(rt, name, err)
		return nil, err
	}
	logWorkflowAction(rt, "done", name, map[string]string{"ok": "true"})
	return payload, nil
}

func ensureAllowedWorkflowArgs(args map[string]string, allowed ...string) error {
	allowedSet := map[string]bool{}
	for _, key := range allowed {
		allowedSet[key] = true
	}
	for key := range args {
		if !allowedSet[key] {
			return apperr.New(apperr.CodeInvalidArgs, "unsupported workflow arg --"+key)
		}
	}
	return nil
}

func requiredWorkflowString(args map[string]string, key string) (string, error) {
	value := strings.TrimSpace(args[key])
	if value == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required workflow arg --"+key)
	}
	return value, nil
}

func parseWorkflowVMID(args map[string]string, key string, required bool) (int, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		if required {
			return 0, apperr.New(apperr.CodeInvalidArgs, "missing required workflow arg --"+key)
		}
		return 0, nil
	}
	vmid, err := strconv.Atoi(raw)
	if err != nil || vmid <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, key+" must be a positive integer")
	}
	if err := action.EnsureOperationVMID(vmid); err != nil {
		return 0, err
	}
	return vmid, nil
}

func getWorkflowTargetResource(rt commandRuntime, targetVMID int) (map[string]any, error) {
	listPayload, err := runWorkflowAction(rt, "list_cluster_resources", map[string]string{"type": "vm"}, false)
	if err != nil {
		return nil, err
	}
	list, _ := listPayload["result"].([]any)
	for _, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vmid, ok := workflowAnyToInt(entry["vmid"])
		if ok && vmid == targetVMID {
			return entry, nil
		}
	}
	return nil, nil
}

func workflowAnyToInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func workflowAnyToBool(v any) bool {
	switch value := v.(type) {
	case bool:
		return value
	case int:
		return value != 0
	case int64:
		return value != 0
	case float64:
		return value != 0
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(value))
		return trimmed == "1" || trimmed == "true" || trimmed == "yes"
	default:
		return false
	}
}

func resolveWorkspaceRoot() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", apperr.Wrap(apperr.CodeConfig, "failed to resolve working directory", err)
	}
	current := filepath.Clean(start)
	for {
		if isDir(filepath.Join(current, ".git")) {
			if isDir(filepath.Join(current, "applications", "proxmox-cli", "src")) {
				return current, nil
			}
		}
		if current == filepath.Dir(current) {
			break
		}
		current = filepath.Dir(current)
	}
	return start, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func wrapWorkflowStepError(step string, vmid int, err error) error {
	message := fmt.Sprintf("workflow step failed: %s; cleanup may be required for vmid=%d", step, vmid)
	return apperr.Wrap(apperr.CodeNetwork, message, err)
}

func logWorkflowAction(rt commandRuntime, phase string, name string, args map[string]string) {
	if rt.Stderr == nil {
		return
	}
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := args[key]
		if isWorkflowSensitiveArg(key) {
			value = "***REDACTED***"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	_, _ = fmt.Fprintf(rt.Stderr, "[workflow] %s action=%s args={%s}\n", phase, name, strings.Join(parts, " "))
}

func logWorkflowActionError(rt commandRuntime, name string, err error) {
	if rt.Stderr == nil {
		return
	}
	_, _ = fmt.Fprintf(rt.Stderr, "[workflow] fail action=%s err=%s\n", name, err.Error())
}

func isWorkflowSensitiveArg(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	if lower == "" {
		return false
	}
	sensitive := []string{"password", "token", "ticket", "csrf", "secret"}
	for _, marker := range sensitive {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func workflowActionSummary(actionName string, payload map[string]any) map[string]any {
	result, _ := payload["result"]
	return map[string]any{
		"action": actionName,
		"ok":     payload["ok"],
		"result": result,
	}
}

func generateWorkflowPassword(length int) (string, error) {
	if length <= 0 {
		return "", apperr.New(apperr.CodeInvalidArgs, "password length must be positive")
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, length)
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to generate workflow password", err)
	}
	for i := 0; i < length; i++ {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}
