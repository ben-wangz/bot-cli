package vmcap

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func RunUpdateVMConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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

func RunAttachCDROMISO(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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
	media := strings.TrimSpace(req.Args["media"])
	slotValue := iso
	if media != "" && !strings.Contains(strings.ToLower(slotValue), "media=") {
		slotValue = slotValue + ",media=" + media
	}
	form := url.Values{}
	form.Set(ideSlot, slotValue)
	path := fmt.Sprintf("/nodes/%s/qemu/%d/config", url.PathEscape(node), vmid)
	data, err := client.PutFormData(ctx, path, form)
	if err != nil {
		return nil, err
	}
	request := map[string]any{"node": node, "vmid": vmid, "slot": ideSlot, "iso": iso}
	if media != "" {
		request["media"] = media
	}
	result := writeResult(req, request, data)
	if media != "" {
		mergeDiagnosticsMap(result, map[string]any{"slot_value": slotValue})
	}
	return result, nil
}

func RunSetNetBootConfig(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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

func RunEnableSerialConsole(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
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
