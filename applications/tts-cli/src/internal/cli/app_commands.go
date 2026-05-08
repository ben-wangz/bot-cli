package cli

import (
	"context"
	"io"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/capability"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/ttsapi"
)

type commandRuntime struct {
	Opts   GlobalOptions
	Client *ttsapi.Client
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
	case "--help", "-h":
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
		return runDescribeCapability(args[1:])
	}
	if args[0] == "list" {
		return map[string]any{
			"ok": true,
			"request": map[string]any{"capability": "list"},
			"result": map[string]any{"capabilities": capability.Names()},
			"diagnostics": map[string]any{},
		}, nil
	}
	if hasHelp(args[1:]) {
		detail, ok := capabilityDetailHelp(args[0])
		if !ok {
			return nil, apperr.New(apperr.CodeInvalidArgs, "capability not implemented yet: "+args[0])
		}
		_, _ = io.WriteString(rt.Stdout, detail)
		return nil, nil
	}
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(rt.Opts.OutputDir) != "" && strings.TrimSpace(parsedArgs["output_dir"]) == "" {
		parsedArgs["output_dir"] = strings.TrimSpace(rt.Opts.OutputDir)
	}
	if strings.TrimSpace(rt.Opts.OutputName) != "" && strings.TrimSpace(parsedArgs["output_name"]) == "" {
		parsedArgs["output_name"] = strings.TrimSpace(rt.Opts.OutputName)
	}
	return capability.Dispatch(context.Background(), rt.Client, capability.Request{Name: args[0], Args: parsedArgs})
}

func runDescribeCapability(args []string) (map[string]any, error) {
	if len(args) == 0 {
		items := make([]map[string]any, 0)
		for _, name := range capability.Names() {
			detail, _ := capability.Describe(name)
			items = append(items, shapeDescribeResult(name, detail))
		}
		return map[string]any{
			"ok": true,
			"request": map[string]any{"capability": "describe"},
			"result": items,
			"diagnostics": map[string]any{},
		}, nil
	}
	detail, ok := capability.Describe(args[0])
	if !ok {
		return nil, apperr.New(apperr.CodeInvalidArgs, "capability not implemented yet: "+args[0])
	}
	return map[string]any{
		"ok": true,
		"request": map[string]any{"capability": "describe", "name": args[0]},
		"result": shapeDescribeResult(args[0], detail),
		"diagnostics": map[string]any{},
	}, nil
}

func shapeDescribeResult(name string, detail map[string]any) map[string]any {
	required := make([]map[string]any, 0)
	optional := make([]map[string]any, 0)
	if args, ok := detail["args"].([]map[string]any); ok {
		for _, arg := range args {
			item := map[string]any{
				"name":        asString(arg["name"]),
				"description": asString(arg["description"]),
			}
			if req, _ := arg["required"].(bool); req {
				required = append(required, item)
			} else {
				optional = append(optional, item)
			}
		}
	}
	result := map[string]any{
		"command":    "capability",
		"action":     "describe",
		"capability": name,
		"summary":    asString(detail["summary"]),
		"args": map[string]any{
			"required": required,
			"optional": optional,
		},
		"examples": detail["examples"],
		"help": map[string]any{
			"usage": "tts-cli capability " + name + " [--key value]",
			"detail_help": "tts-cli capability " + name + " --help",
		},
	}
	if readOnly, ok := detail["read_only"].(bool); ok {
		result["read_only"] = readOnly
	}
	if modelRules, ok := detail["model_rules"]; ok {
		result["model_rules"] = modelRules
	}
	return result
}
