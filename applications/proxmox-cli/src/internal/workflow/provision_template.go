package workflow

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const (
	provisionTemplateWorkflowName = "provision-template-from-artifact"
	provisionDiskStorage          = "local-lvm"
	provisionDiskSizeGB           = 32
	provisionMemoryMB             = 8192
	provisionCPUCores             = 2
	provisionNet0                 = "virtio,bridge=vmbr0"
	provisionBootInstall          = "order=ide2"
	provisionBootTemplate         = "order=scsi0"
)

type ProvisionTemplateDeps struct {
	AuthScope                       string
	RunStep                         func(name string, stepArgs map[string]string, wait bool) (map[string]any, error)
	EnsureUploadedISOArtifactExists func(node string, storage string, volid string) error
}

func RunProvisionTemplateFromArtifact(args map[string]string, deps ProvisionTemplateDeps) (map[string]any, error) {
	if deps.RunStep == nil {
		return nil, apperr.New(apperr.CodeInternal, "workflow dependency missing: run step")
	}
	if deps.EnsureUploadedISOArtifactExists == nil {
		return nil, apperr.New(apperr.CodeInternal, "workflow dependency missing: ensure uploaded iso artifact exists")
	}
	if err := EnsureAllowedArgs(args, "node", "target-vmid", "artifact-iso", "install-timeout-seconds", "template-name", "memory", "cores", "disk-size-gb", "net0", "install-boot-order", "runtime-boot-order", "bootdisk", "resume-from", "pool"); err != nil {
		return nil, err
	}
	node, err := RequiredString(args, "node")
	if err != nil {
		return nil, err
	}
	targetVMID, err := ParseVMID(args, "target-vmid", true)
	if err != nil {
		return nil, err
	}
	artifactISO, err := RequiredString(args, "artifact-iso")
	if err != nil {
		return nil, err
	}
	storageID, _, err := ParseArtifactISOVolID(artifactISO)
	if err != nil {
		return nil, err
	}
	installTimeoutSeconds, err := RequiredPositiveInt(args, "install-timeout-seconds")
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
		if _, err := deps.RunStep("list_nodes", map[string]string{}, false); err != nil {
			return nil, WrapStepError("list_nodes", targetVMID, err)
		}
		targetResource, err := getWorkflowTargetResource(deps, targetVMID)
		if err != nil {
			return nil, err
		}
		if targetResource != nil {
			return nil, apperr.New(apperr.CodeInvalidArgs, fmt.Sprintf("target-vmid %d already exists", targetVMID))
		}
	}
	if err := deps.EnsureUploadedISOArtifactExists(node, storageID, artifactISO); err != nil {
		return nil, err
	}

	cfg, err := resolveProvisionTemplateConfig(args, targetVMID)
	if err != nil {
		return nil, err
	}
	serialLogPath, err := ResolveProvisionSerialLogPath(targetVMID)
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
			"scsi0":     fmt.Sprintf("%s:%d", provisionDiskStorage, cfg.DiskSizeGB),
			"if-exists": "fail",
		}
		if poolID != "" {
			createVMArgs["pool"] = poolID
		}
		if _, err := deps.RunStep("create_vm", createVMArgs, true); err != nil {
			return nil, WrapStepError("create_vm", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "create_vm"})
		if _, err := deps.RunStep("enable_serial_console", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, true); err != nil {
			return nil, WrapStepError("enable_serial_console", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "enable_serial_console"})
		if _, err := deps.RunStep("set_vm_agent", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "enabled": "1"}, true); err != nil {
			return nil, WrapStepError("set_vm_agent", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "set_vm_agent", "enabled": 1})
		if _, err := deps.RunStep("attach_cdrom_iso", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "iso": artifactISO, "slot": "ide2", "media": "cdrom"}, true); err != nil {
			return nil, WrapStepError("attach_cdrom_iso(install)", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "attach_cdrom_iso", "iso": artifactISO})
		if _, err := deps.RunStep("set_net_boot_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "net0": cfg.Net0, "boot": cfg.InstallBootOrder, "bootdisk": cfg.BootDisk}, true); err != nil {
			return nil, WrapStepError("set_net_boot_config(install)", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "set_net_boot_config_install", "boot": cfg.InstallBootOrder})
		if _, err := deps.RunStep("vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "start", "desired-state": "running"}, true); err != nil {
			return nil, WrapStepError("vm_power(start install)", targetVMID, err)
		}
		steps = append(steps, map[string]any{"step": "vm_power_start_install"})
	} else {
		steps = append(steps, map[string]any{"step": "resume_from_serial_wait", "warning": "resume-from serial_wait is not fully validated; use with caution"})
	}

	appendMode := "0"
	if resumeFrom == "serial_wait" {
		appendMode = "1"
	}
	capturePayload, err := deps.RunStep("serial_ws_capture_to_file", map[string]string{
		"node":            node,
		"vmid":            strconv.Itoa(targetVMID),
		"log-path":        serialLogPath,
		"append":          appendMode,
		"timeout-seconds": strconv.Itoa(installTimeoutSeconds),
	}, false)
	if err != nil {
		return nil, WrapStepError("serial_ws_capture_to_file", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "serial_ws_capture_to_file", "log_path": serialLogPath, "timeout_seconds": installTimeoutSeconds, "append": appendMode == "1"})
	captureResult, _ := capturePayload["result"].(map[string]any)
	closedByRemote := AnyToBool(captureResult["closed_by_remote"])
	timedOut := !closedByRemote

	statusPayload, err := deps.RunStep("get_vm_status", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, false)
	if err != nil {
		return nil, WrapStepError("get_vm_status(after capture)", targetVMID, err)
	}
	statusResult, _ := statusPayload["result"].(map[string]any)
	vmStatus := strings.TrimSpace(asStringValue(statusResult["status"]))
	steps = append(steps, map[string]any{"step": "get_vm_status_after_capture", "status": vmStatus})
	if timedOut && vmStatus != "stopped" {
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("serial wait timed out and vm did not stop; serial_log_path=%s vm_status=%s next_step_hint=check serial log and current vm status", serialLogPath, vmStatus))
	}

	if _, err := deps.RunStep("attach_cdrom_iso", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "iso": "none,media=cdrom", "slot": "ide2"}, true); err != nil {
		return nil, WrapStepError("attach_cdrom_iso(detach)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "detach_cdrom_iso"})
	if _, err := deps.RunStep("set_net_boot_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "net0": cfg.Net0, "boot": cfg.RuntimeBootOrder, "bootdisk": cfg.BootDisk}, true); err != nil {
		return nil, WrapStepError("set_net_boot_config(runtime)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "set_net_boot_config_runtime", "boot": cfg.RuntimeBootOrder})
	if _, err := deps.RunStep("vm_power", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID), "mode": "stop", "desired-state": "stopped"}, true); err != nil {
		return nil, WrapStepError("vm_power(stop)", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "vm_power_stop"})
	if _, err := deps.RunStep("convert_vm_to_template", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, true); err != nil {
		return nil, WrapStepError("convert_vm_to_template", targetVMID, err)
	}
	steps = append(steps, map[string]any{"step": "convert_vm_to_template"})
	configPayload, err := deps.RunStep("get_vm_config", map[string]string{"node": node, "vmid": strconv.Itoa(targetVMID)}, false)
	if err != nil {
		return nil, WrapStepError("get_vm_config", targetVMID, err)
	}
	configResult, _ := configPayload["result"].(map[string]any)
	steps = append(steps, map[string]any{"step": "get_vm_config", "agent": asStringValue(configResult["agent"])})

	return map[string]any{
		"workflow": provisionTemplateWorkflowName,
		"ok":       true,
		"scope":    deps.AuthScope,
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
