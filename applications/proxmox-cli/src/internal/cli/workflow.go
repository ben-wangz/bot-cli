package cli

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/action"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const (
	workflowUbuntu24WithAgentTemplate = "ubuntu24-with-agent-template"
	workflowBootstrapUserPoolACL      = "bootstrap-bot-user-pool-acl"
	workflowISOStorage                = "local"
	workflowDiskStorage               = "local-lvm"
	workflowDiskSizeGB                = 32
	workflowMemoryMB                  = 4096
	workflowCPUCores                  = 2
	workflowNet0                      = "virtio,bridge=vmbr0"
	workflowBootInstall               = "order=ide2;scsi0;net0"
	workflowBootTemplate              = "order=scsi0"
	workflowInstallExpect             = "subiquity/Shutdown/shutdown: mode=POWEROFF"
)

type workflowPaths struct {
	SourceISO      string
	PrebuiltISO    string
	BuildOutputISO string
	UserData       string
	MetaData       string
	WorkDir        string
	SerialLog      string
	TemplateIDFile string
}

func executeWorkflow(rt commandRuntime, name string, args map[string]string) (map[string]any, error) {
	switch name {
	case workflowUbuntu24WithAgentTemplate:
		return runWorkflowUbuntu24WithAgentTemplate(rt, args)
	case workflowBootstrapUserPoolACL:
		return runWorkflowBootstrapTestUserPoolACL(rt, args)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow not implemented yet: "+name)
	}
}

func runWorkflowBootstrapTestUserPoolACL(rt commandRuntime, args map[string]string) (map[string]any, error) {
	if err := ensureAllowedWorkflowArgs(args, "userid", "poolid", "password", "pool-comment", "user-comment", "if-exists"); err != nil {
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

func runWorkflowUbuntu24WithAgentTemplate(rt commandRuntime, args map[string]string) (map[string]any, error) {
	if err := ensureAllowedWorkflowArgs(args, "node", "target-vmid"); err != nil {
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
	targetResource, err := getWorkflowTargetResource(rt, targetVMID)
	if err != nil {
		return nil, err
	}
	paths, err := resolveWorkflowPaths(targetVMID)
	if err != nil {
		return nil, err
	}
	if err := ensureWorkflowInputFiles(paths); err != nil {
		return nil, err
	}
	if targetResource != nil && parseWorkflowTemplateFlag(targetResource["template"]) {
		if err := writeWorkflowTemplateIDFile(paths.TemplateIDFile, targetVMID); err != nil {
			return nil, err
		}
		steps := []map[string]any{
			{"step": "reuse_existing_template", "vmid": targetVMID},
			{"step": "write_template_id_file", "path": paths.TemplateIDFile},
		}
		return map[string]any{
			"workflow": workflowUbuntu24WithAgentTemplate,
			"ok":       true,
			"scope":    rt.Opts.AuthScope,
			"request": map[string]any{
				"node":        node,
				"target_vmid": targetVMID,
			},
			"result": map[string]any{
				"node":             node,
				"template_vmid":    targetVMID,
				"template_id_path": paths.TemplateIDFile,
				"reused_template":  true,
				"steps":            steps,
			},
			"diagnostics": map[string]any{"step_count": len(steps), "skipped": "target vmid already template"},
		}, nil
	}

	steps := make([]map[string]any, 0, 16)
	templateName := fmt.Sprintf("ubuntu-24-with-agent-template-%d", targetVMID)
	installerISO := ""
	isoFilename := ""
	isoWithMedia := isoVolIDWithMedia(isoFilename)

	if isFile(paths.PrebuiltISO) {
		installerISO = paths.PrebuiltISO
		isoFilename = filepath.Base(installerISO)
		steps = append(steps, map[string]any{"step": "reuse_prebuilt_autoinstall_iso", "path": installerISO})
	} else {
		buildPayload, buildErr := runWorkflowAction(rt, "build_ubuntu_autoinstall_iso", map[string]string{
			"source-iso":     paths.SourceISO,
			"output-iso":     paths.BuildOutputISO,
			"user-data-path": paths.UserData,
			"meta-data-path": paths.MetaData,
			"work-dir":       paths.WorkDir,
		}, false)
		if buildErr != nil {
			return nil, wrapWorkflowStepError("build_ubuntu_autoinstall_iso", targetVMID, buildErr)
		}
		installerISO = paths.BuildOutputISO
		isoFilename = filepath.Base(installerISO)
		steps = append(steps, map[string]any{"step": "build_autoinstall_iso", "action": "build_ubuntu_autoinstall_iso", "output_iso": installerISO, "diagnostics": buildPayload["diagnostics"]})
	}

	uploadPayload, err := runWorkflowAction(rt, "storage_upload_iso", map[string]string{
		"node":        node,
		"storage":     workflowISOStorage,
		"source-path": installerISO,
		"filename":    isoFilename,
		"if-exists":   "skip",
	}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("storage_upload_iso", targetVMID, err)
	}
	uploadResult, _ := uploadPayload["result"].(map[string]any)
	isoVolID := strings.TrimSpace(asStringValue(uploadResult["volid"]))
	if isoVolID == "" {
		isoVolID = fmt.Sprintf("%s:iso/%s", workflowISOStorage, isoFilename)
	}
	isoWithMedia = isoVolIDWithMedia(isoVolID)
	steps = append(steps, map[string]any{"step": "upload_iso", "action": "storage_upload_iso", "iso_volid": isoVolID})

	createPayload, err := runWorkflowAction(rt, "create_vm", map[string]string{
		"node":      node,
		"vmid":      strconv.Itoa(targetVMID),
		"name":      templateName,
		"memory":    strconv.Itoa(workflowMemoryMB),
		"cores":     strconv.Itoa(workflowCPUCores),
		"net0":      workflowNet0,
		"scsi0":     fmt.Sprintf("%s:%d", workflowDiskStorage, workflowDiskSizeGB),
		"if-exists": "reuse",
	}, true)
	if err != nil {
		return nil, wrapWorkflowStepError("create_vm", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "create_vm", "action": "create_vm", "diagnostics": createPayload["diagnostics"]})

	if _, err := runWorkflowAction(rt, "enable_serial_console", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, true); err != nil {
		return nil, wrapWorkflowStepError("enable_serial_console", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "enable_serial_console", "action": "enable_serial_console"})

	if _, err := runWorkflowAction(rt, "set_vm_agent", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "enabled": "1"}, true); err != nil {
		return nil, wrapWorkflowStepError("set_vm_agent", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "enable_vm_agent_flag", "action": "set_vm_agent"})

	if _, err := runWorkflowAction(rt, "attach_cdrom_iso", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "iso": isoWithMedia, "slot": "ide2"}, true); err != nil {
		return nil, wrapWorkflowStepError("attach_cdrom_iso", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "attach_installer_iso", "action": "attach_cdrom_iso", "iso": isoWithMedia})

	if _, err := runWorkflowAction(rt, "set_net_boot_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "net0": workflowNet0, "boot": workflowBootInstall, "bootdisk": "scsi0"}, true); err != nil {
		return nil, wrapWorkflowStepError("set_net_boot_config(install)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "set_install_boot_order", "action": "set_net_boot_config", "boot": workflowBootInstall})

	if _, err := runWorkflowAction(rt, "vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "start", "desired-state": "running"}, true); err != nil {
		return nil, wrapWorkflowStepError("vm_power(start install)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "start_installer", "action": "vm_power", "mode": "start"})

	capturePayload, err := runWorkflowAction(rt, "serial_ws_capture_to_file", map[string]string{
		"node":            node,
		"vmid":            strconv.Itoa(targetVMID),
		"log-path":        paths.SerialLog,
		"append":          "0",
		"expect":          workflowInstallExpect,
		"timeout-seconds": "3600",
	}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("serial_ws_capture_to_file", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "wait_install_finish", "action": "serial_ws_capture_to_file", "log_path": paths.SerialLog, "diagnostics": capturePayload["diagnostics"]})

	statusAfterInstall, err := runWorkflowAction(rt, "get_vm_status", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("get_vm_status(after install)", targetVMID, err)
	}
	statusMap, _ := statusAfterInstall["result"].(map[string]any)
	status := strings.TrimSpace(asStringValue(statusMap["status"]))
	if status != "stopped" {
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("expected VM %d to be stopped after autoinstall poweroff, got status=%s", targetVMID, status))
	}
	steps = append(steps, map[string]any{"step": "verify_poweroff", "action": "get_vm_status", "status": status})

	if _, err := runWorkflowAction(rt, "attach_cdrom_iso", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "iso": "none,media=cdrom", "slot": "ide2"}, true); err != nil {
		return nil, wrapWorkflowStepError("detach_installer_iso", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "detach_installer_iso", "action": "attach_cdrom_iso"})

	if _, err := runWorkflowAction(rt, "set_net_boot_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "net0": workflowNet0, "boot": workflowBootTemplate, "bootdisk": "scsi0"}, true); err != nil {
		return nil, wrapWorkflowStepError("set_net_boot_config(template)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "set_disk_boot_order", "action": "set_net_boot_config", "boot": workflowBootTemplate})

	if _, err := runWorkflowAction(rt, "vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "start", "desired-state": "running"}, true); err != nil {
		return nil, wrapWorkflowStepError("vm_power(start verify)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "start_disk_boot_verification", "action": "vm_power", "mode": "start"})

	qgaPayload, qgaPolls, err := pollWorkflowQGAReady(rt, node, targetVMID, 6*time.Minute, 10*time.Second)
	if err != nil {
		return nil, wrapWorkflowStepError("agent_network_get_interfaces", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "verify_qga_ready", "action": "agent_network_get_interfaces", "poll_count": qgaPolls, "result": qgaPayload["result"]})

	if _, err := runWorkflowAction(rt, "vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "stop", "desired-state": "stopped"}, true); err != nil {
		return nil, wrapWorkflowStepError("vm_power(stop before template)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "stop_before_template", "action": "vm_power", "mode": "stop"})

	if _, err := runWorkflowAction(rt, "convert_vm_to_template", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, true); err != nil {
		return nil, wrapWorkflowStepError("convert_vm_to_template", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "convert_to_template", "action": "convert_vm_to_template"})

	configPayload, err := runWorkflowAction(rt, "get_vm_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, false)
	if err != nil {
		return nil, wrapWorkflowStepError("get_vm_config(verify template)", targetVMID, err)
	}
	configResult, _ := configPayload["result"].(map[string]any)
	agentConfig := asStringValue(configResult["agent"])
	steps = append(steps, map[string]any{"step": "verify_template_config", "action": "get_vm_config", "agent": agentConfig})

	if err := writeWorkflowTemplateIDFile(paths.TemplateIDFile, targetVMID); err != nil {
		return nil, err
	}
	steps = append(steps, map[string]any{"step": "write_template_id_file", "path": paths.TemplateIDFile})

	return map[string]any{
		"workflow": workflowUbuntu24WithAgentTemplate,
		"ok":       true,
		"scope":    rt.Opts.AuthScope,
		"request": map[string]any{
			"node":        node,
			"target_vmid": targetVMID,
		},
		"result": map[string]any{
			"node":             node,
			"template_vmid":    targetVMID,
			"template_name":    templateName,
			"template_id_path": paths.TemplateIDFile,
			"installer_iso":    installerISO,
			"serial_log_path":  paths.SerialLog,
			"agent_config":     agentConfig,
			"steps":            steps,
		},
		"diagnostics": map[string]any{"step_count": len(steps), "qga_poll_count": qgaPolls},
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

func pollWorkflowQGAReady(rt commandRuntime, node string, vmid int, timeout time.Duration, interval time.Duration) (map[string]any, int, error) {
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	polls := 0
	for {
		payload, err := runWorkflowAction(rt, "agent_network_get_interfaces", map[string]string{
			"node": node,
			"vmid": strconv.Itoa(vmid),
		}, false)
		if err == nil {
			polls++
			return payload, polls, nil
		}
		polls++
		if time.Now().After(deadline) {
			return nil, polls, apperr.Wrap(apperr.CodeNetwork, "qga did not become ready before timeout", err)
		}
		time.Sleep(interval)
	}
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

func ensureWorkflowTargetVMIDFree(rt commandRuntime, targetVMID int) error {
	target, err := getWorkflowTargetResource(rt, targetVMID)
	if err != nil {
		return err
	}
	if target != nil {
		return apperr.New(apperr.CodeInvalidArgs, fmt.Sprintf("target-vmid %d already exists", targetVMID))
	}
	return nil
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

func parseWorkflowTemplateFlag(v any) bool {
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

func resolveWorkflowPaths(targetVMID int) (workflowPaths, error) {
	root, err := resolveWorkspaceRoot()
	if err != nil {
		return workflowPaths{}, err
	}
	return workflowPaths{
		SourceISO:      filepath.Join(root, "build", "ubuntu-24.04.2-live-server-amd64.iso"),
		PrebuiltISO:    filepath.Join(root, "build", "ubuntu-24.04.2-shim-nocloud-serial-poweroff.iso"),
		BuildOutputISO: filepath.Join(root, "build", fmt.Sprintf("ubuntu-24-with-agent-template-%d.iso", targetVMID)),
		UserData:       filepath.Join(root, "applications", "proxmox-cli", "assets", "single-disk-nocloud", "user-data"),
		MetaData:       filepath.Join(root, "applications", "proxmox-cli", "assets", "single-disk-nocloud", "meta-data"),
		WorkDir:        filepath.Join(root, "build", "autoinstall-iso-work", fmt.Sprintf("workflow-template-%d", targetVMID)),
		SerialLog:      filepath.Join(root, "build", fmt.Sprintf("serial-bootstrap-template-%d.log", targetVMID)),
		TemplateIDFile: filepath.Join(root, "build", "ubuntu-24-with-agent.vm-template.id"),
	}, nil
}

func ensureWorkflowInputFiles(paths workflowPaths) error {
	if !isFile(paths.PrebuiltISO) && !isFile(paths.SourceISO) {
		return apperr.New(apperr.CodeConfig, "source iso not found: "+paths.SourceISO)
	}
	if !isFile(paths.UserData) {
		return apperr.New(apperr.CodeConfig, "user-data not found: "+paths.UserData)
	}
	if !isFile(paths.MetaData) {
		return apperr.New(apperr.CodeConfig, "meta-data not found: "+paths.MetaData)
	}
	return nil
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

func writeWorkflowTemplateIDFile(path string, vmid int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to create template id directory", err)
	}
	content := strconv.Itoa(vmid) + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to write template id file", err)
	}
	return nil
}

func wrapWorkflowStepError(step string, vmid int, err error) error {
	message := fmt.Sprintf("workflow step failed: %s; cleanup may be required for vmid=%d", step, vmid)
	return apperr.Wrap(apperr.CodeNetwork, message, err)
}

func isoVolIDWithMedia(volid string) string {
	trimmed := strings.TrimSpace(volid)
	if trimmed == "" {
		return trimmed
	}
	if strings.Contains(trimmed, ",media=cdrom") {
		return trimmed
	}
	return trimmed + ",media=cdrom"
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
