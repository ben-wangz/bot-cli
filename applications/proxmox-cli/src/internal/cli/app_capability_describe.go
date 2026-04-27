package cli

import (
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability"
)

func runCapabilityDescribe(rt commandRuntime, args []string) (map[string]any, error) {
	if len(args) == 0 || hasHelp(args[:1]) {
		return map[string]any{
			"command": "capability",
			"action":  "describe",
			"usage":   "proxmox-cli capability describe <name>",
		}, nil
	}
	name := strings.TrimSpace(args[0])
	if name == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "capability name is required")
	}
	meta, ok := capability.LookupMeta(name)
	if !ok {
		return nil, apperr.New(apperr.CodeInvalidArgs, "capability not implemented yet: "+name)
	}
	helpMeta, _ := capability.LookupHelpMeta(name)
	return map[string]any{
		"capability": name,
		"group":      meta.Capability,
		"async":      meta.Async,
		"wait": map[string]any{
			"skip_reason": meta.WaitSkipReason,
		},
		"help": helpMeta,
	}, nil
}
