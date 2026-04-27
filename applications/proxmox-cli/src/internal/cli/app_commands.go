package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/auth"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/taskwait"
)

type commandRuntime struct {
	Opts    GlobalOptions
	Creds   auth.Credentials
	Sources auth.Sources
	Client  *pveapi.Client
	Stdout  io.Writer
	Stderr  io.Writer
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
	case "console":
		payload, err = runConsoleCommand(rt, commandArgs)
	case "auth":
		payload, err = runAuthCommand(rt, commandArgs)
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
		return runCapabilityDescribe(rt, args[1:])
	}
	if hasHelp(args[1:]) {
		detail, ok := capabilityDetailHelp(args[0])
		if !ok {
			return nil, apperr.New(apperr.CodeInvalidArgs, "capability not implemented yet: "+args[0])
		}
		_, _ = io.WriteString(rt.Stdout, detail)
		return nil, nil
	}
	name := args[0]
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	if rt.Opts.DryRun {
		return map[string]any{
			"capability": name,
			"ok":         true,
			"scope":      rt.Opts.AuthScope,
			"dry_run":    true,
			"request":    parsedArgs,
		}, nil
	}
	result, meta, err := executeCapability(rt, name, parsedArgs)
	if err != nil {
		return nil, err
	}
	applyCapabilityMeta(result, meta)
	if !rt.Opts.Wait {
		return result, nil
	}
	return applyCapabilityWait(rt, result, meta)
}

func executeCapability(rt commandRuntime, name string, parsedArgs map[string]string) (map[string]any, capability.Meta, error) {
	req := capability.Request{Name: name, Args: parsedArgs, Scope: rt.Opts.AuthScope}
	result, meta, err := capability.Dispatch(context.Background(), rt.Client, req)
	if err != nil {
		return nil, capability.Meta{}, err
	}
	return result, meta, nil
}

func applyCapabilityWait(rt commandRuntime, result map[string]any, meta capability.Meta) (map[string]any, error) {
	if hasWaitSkipDiagnostic(result) {
		return result, nil
	}
	if !meta.Async {
		reason := strings.TrimSpace(meta.WaitSkipReason)
		if reason == "" {
			reason = "capability is synchronous"
		}
		mergeDiagnostics(result, map[string]any{"wait_skipped": reason})
		return result, nil
	}
	node, upid := getWaitTarget(result)
	if node == "" || upid == "" {
		return nil, apperr.New(apperr.CodeInternal, "wait requested but async capability did not provide node/upid")
	}
	waitResult, waitErr := taskwait.WaitTask(context.Background(), rt.Client, node, upid, taskwait.WaitOptions{Timeout: waitTimeout(rt.Opts.Timeout), Interval: 2 * time.Second})
	if waitErr != nil {
		return nil, waitErr
	}
	mergeDiagnostics(result, map[string]any{"wait_status": waitResult})
	return result, nil
}

func mergeDiagnostics(result map[string]any, extra map[string]any) {
	diagnostics := map[string]any{}
	if current, ok := result["diagnostics"].(map[string]any); ok {
		for k, v := range current {
			diagnostics[k] = v
		}
	}
	for k, v := range extra {
		diagnostics[k] = v
	}
	result["diagnostics"] = diagnostics
}

func waitTimeout(timeout time.Duration) time.Duration {
	if timeout < 5*time.Minute {
		return 5 * time.Minute
	}
	return timeout
}

func getWaitTarget(result map[string]any) (string, string) {
	request, _ := result["request"].(map[string]any)
	node := asStringValue(request["node"])
	resultData, _ := result["result"].(map[string]any)
	upid := asStringValue(resultData["upid"])
	return node, upid
}

func asStringValue(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func hasWaitSkipDiagnostic(result map[string]any) bool {
	diagnostics, ok := result["diagnostics"].(map[string]any)
	if !ok {
		return false
	}
	message := strings.TrimSpace(asStringValue(diagnostics["wait_skipped"]))
	return message != ""
}

func applyCapabilityMeta(result map[string]any, meta capability.Meta) {
	if strings.TrimSpace(meta.Capability) == "" {
		return
	}
	mergeDiagnostics(result, map[string]any{"capability": meta.Capability})
}

func runWorkflowCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, workflowHelp())
		return nil, nil
	}
	if len(args) == 0 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow name is required")
	}
	name := args[0]
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	return executeWorkflow(rt, name, parsedArgs)
}

func runConsoleCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, consoleHelp())
		return nil, nil
	}
	subcommand := ""
	rest := []string{}
	if len(args) > 0 {
		subcommand = args[0]
		rest = args[1:]
	}
	if subcommand == "" {
		subcommand = "script"
	}
	return map[string]any{
		"command":    "console",
		"subcommand": subcommand,
		"args":       rest,
		"status":     "skeleton-ready",
	}, nil
}

func runAuthCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, authHelp())
		return nil, nil
	}
	subcommand := "inspect"
	if len(args) > 0 {
		subcommand = args[0]
	}
	switch subcommand {
	case "inspect":
		return map[string]any{
			"command":        "auth",
			"subcommand":     "inspect",
			"auth_scope":     rt.Opts.AuthScope,
			"credentials":    rt.Creds.SafeSummary(),
			"source_summary": rt.Sources.Summary(),
		}, nil
	case "example-auth-file":
		return map[string]any{
			"auth_file_example": auth.ExampleAuthFile(),
		}, nil
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unknown auth subcommand: "+subcommand)
	}
}
