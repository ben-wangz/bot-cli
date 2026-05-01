package cli

import (
	"context"
	"io"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/capability"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/imageapi"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/workflow"
)

type commandRuntime struct {
	Opts   GlobalOptions
	Client *imageapi.Client
	Stdout io.Writer
}

func dispatchCommand(rt commandRuntime, args []string) error {
	command := args[0]
	commandArgs := []string{}
	if len(args) > 1 {
		commandArgs = args[1:]
	}
	var payload map[string]any
	var err error
	switch command {
	case "capability":
		payload, err = runCapabilityCommand(rt, commandArgs)
	case "workflow":
		payload, err = runWorkflowCommand(rt, commandArgs)
	case "help", "--help", "-h":
		_, _ = io.WriteString(rt.Stdout, rootHelp())
		return nil
	default:
		return apperr.New(apperr.CodeInvalidArgs, "unknown command: "+command)
	}
	if err != nil {
		return err
	}
	if payload == nil {
		return nil
	}
	return output.Render(rt.Stdout, rt.Opts.Output, payload)
}

func runCapabilityCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if len(args) == 0 || hasHelp(args[:1]) {
		_, _ = io.WriteString(rt.Stdout, capabilityHelp())
		return nil, nil
	}
	if args[0] == "describe" {
		if len(args) == 1 {
			items := make([]map[string]any, 0)
			for _, name := range capability.Names() {
				detail, _ := capability.Describe(name)
				items = append(items, detail)
			}
			return map[string]any{"ok": true, "request": map[string]any{"capability": "describe"}, "result": items, "diagnostics": map[string]any{}}, nil
		}
		detail, ok := capability.Describe(args[1])
		if !ok {
			return nil, apperr.New(apperr.CodeInvalidArgs, "capability not implemented yet: "+args[1])
		}
		return map[string]any{"ok": true, "request": map[string]any{"capability": "describe", "name": args[1]}, "result": detail, "diagnostics": map[string]any{}}, nil
	}
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	return capability.Dispatch(context.Background(), rt.Client, capability.Request{Name: args[0], Args: parsedArgs})
}

func runWorkflowCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if len(args) == 0 || hasHelp(args[:1]) {
		_, _ = io.WriteString(rt.Stdout, workflowHelp())
		return nil, nil
	}
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	return workflow.Execute(context.Background(), rt.Client, args[0], parsedArgs)
}
