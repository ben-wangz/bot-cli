package cli

import (
	"context"
	"io"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/capability"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/workflow"
)

type commandRuntime struct {
	Opts   GlobalOptions
	Client *aria2rpc.Client
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
	name := args[0]
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	return capability.Dispatch(context.Background(), rt.Client, capability.Request{Name: name, Args: parsedArgs})
}

func runWorkflowCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if len(args) == 0 || hasHelp(args[:1]) {
		_, _ = io.WriteString(rt.Stdout, workflowHelp())
		return nil, nil
	}
	name := args[0]
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	return workflow.Execute(context.Background(), rt.Client, name, parsedArgs)
}
