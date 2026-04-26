package cli

import (
	"context"
	"fmt"
	"net/url"

	workflowpkg "github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/workflow"
)

func executeWorkflow(rt commandRuntime, name string, args map[string]string) (map[string]any, error) {
	return workflowpkg.Execute(name, args, workflowpkg.ExecuteDeps{
		AuthScope: rt.Opts.AuthScope,
		RunStep: func(name string, stepArgs map[string]string, wait bool) (map[string]any, error) {
			return workflowpkg.RunCommandStep(name, stepArgs, wait, workflowpkg.CommandStepDeps{
				Stderr: rt.Stderr,
				ExecCommand: func(args []string, wait bool) (map[string]any, error) {
					return execWorkflowCapabilityCommand(rt, args, wait)
				},
			})
		},
		EnsureUploadedISOArtifactExists: func(node string, storage string, volid string) error {
			return workflowpkg.EnsureUploadedISOArtifactExists(node, storage, volid, workflowpkg.VerifyISOArtifactDeps{
				GetStorageContent: func(node string, storage string) (any, error) {
					query := url.Values{}
					query.Set("content", "iso")
					path := fmt.Sprintf("/nodes/%s/storage/%s/content", url.PathEscape(node), url.PathEscape(storage))
					return rt.Client.GetData(context.Background(), path, query)
				},
			})
		},
	})
}

func execWorkflowCapabilityCommand(rt commandRuntime, args []string, wait bool) (map[string]any, error) {
	stepRT := rt
	stepRT.Opts.Wait = wait
	return runCapabilityCommand(stepRT, args)
}
