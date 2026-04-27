package sshcap

import (
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func rejectUnsupportedPassword(args map[string]string, capabilityName string) error {
	if _, provided := args["password"]; !provided {
		return nil
	}
	name := strings.TrimSpace(capabilityName)
	if name == "" {
		name = "ssh capability"
	}
	return apperr.New(apperr.CodeInvalidArgs, "--password is not supported for "+name+" in batch/key mode; use --identity-file")
}
