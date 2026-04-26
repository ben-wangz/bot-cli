package workflow

import "github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"

const (
	workflowBootstrapUserPoolACL      = "bootstrap-bot-user-pool-acl"
	workflowProvisionTemplateArtifact = "provision-template-from-artifact"
)

type ExecuteDeps struct {
	AuthScope                       string
	RunStep                         func(name string, stepArgs map[string]string, wait bool) (map[string]any, error)
	EnsureUploadedISOArtifactExists func(node string, storage string, volid string) error
}

func Execute(name string, args map[string]string, deps ExecuteDeps) (map[string]any, error) {
	switch name {
	case workflowBootstrapUserPoolACL:
		return RunBootstrapUserPoolACL(args, BootstrapUserPoolACLDeps{
			AuthScope: deps.AuthScope,
			RunStep:   deps.RunStep,
		})
	case workflowProvisionTemplateArtifact:
		return RunProvisionTemplateFromArtifact(args, ProvisionTemplateDeps{
			AuthScope:                       deps.AuthScope,
			RunStep:                         deps.RunStep,
			EnsureUploadedISOArtifactExists: deps.EnsureUploadedISOArtifactExists,
		})
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow not implemented yet: "+name)
	}
}
