package workflow

import (
	"fmt"
	"strings"
)

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
		MemoryMB:         provisionMemoryMB,
		Cores:            provisionCPUCores,
		DiskSizeGB:       provisionDiskSizeGB,
		Net0:             provisionNet0,
		InstallBootOrder: provisionBootInstall,
		RuntimeBootOrder: provisionBootTemplate,
		BootDisk:         "scsi0",
	}
	if v := strings.TrimSpace(args["template-name"]); v != "" {
		cfg.TemplateName = v
	}
	if v, ok, err := OptionalPositiveInt(args, "memory"); err != nil {
		return provisionTemplateConfig{}, err
	} else if ok {
		cfg.MemoryMB = v
	}
	if v, ok, err := OptionalPositiveInt(args, "cores"); err != nil {
		return provisionTemplateConfig{}, err
	} else if ok {
		cfg.Cores = v
	}
	if v, ok, err := OptionalPositiveInt(args, "disk-size-gb"); err != nil {
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

func getWorkflowTargetResource(deps ProvisionTemplateDeps, targetVMID int) (map[string]any, error) {
	listPayload, err := deps.RunStep("list_cluster_resources", map[string]string{"type": "vm"}, false)
	if err != nil {
		return nil, err
	}
	list, _ := listPayload["result"].([]any)
	for _, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vmid, ok := AnyToInt(entry["vmid"])
		if ok && vmid == targetVMID {
			return entry, nil
		}
	}
	return nil, nil
}

func asStringValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}
