package workflow

import (
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"strings"
)

type VerifyISOArtifactDeps struct {
	GetStorageContent func(node string, storage string) (any, error)
}

func EnsureUploadedISOArtifactExists(node string, storage string, volid string, deps VerifyISOArtifactDeps) error {
	if deps.GetStorageContent == nil {
		return apperr.New(apperr.CodeInternal, "workflow dependency missing: get storage content")
	}
	data, err := deps.GetStorageContent(node, storage)
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
