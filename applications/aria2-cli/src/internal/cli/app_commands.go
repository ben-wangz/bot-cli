package cli

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

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
	name := args[0]
	parsedArgs, err := capability.ParseArgs(args[1:])
	if err != nil {
		return nil, err
	}
	payload, err := capability.Dispatch(context.Background(), rt.Client, capability.Request{Name: name, Args: parsedArgs})
	if err != nil {
		return nil, err
	}
	if !rt.Opts.Wait || capability.IsReadOnly(name) {
		return payload, nil
	}
	applyWaitDiagnostics(context.Background(), rt, name, payload)
	return payload, nil
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

func applyWaitDiagnostics(ctx context.Context, rt commandRuntime, capabilityName string, payload map[string]any) {
	gids := extractResultGIDs(payload["result"])
	if len(gids) == 0 {
		mergeDiagnostics(payload, map[string]any{"wait_skipped": "no gid found in capability result"})
		return
	}
	deadline := time.Now().Add(rt.Opts.Timeout)
	if rt.Opts.WaitTimeout > 0 {
		deadline = time.Now().Add(rt.Opts.WaitTimeout)
	}
	if rt.Opts.WaitTimeout < 10*time.Second {
		deadline = time.Now().Add(10 * time.Second)
	}
	waits := make([]map[string]any, 0, len(gids))
	for _, gid := range gids {
		item := waitForGID(ctx, rt, capabilityName, gid, deadline)
		waits = append(waits, item)
	}
	mergeDiagnostics(payload, map[string]any{"wait": waits})
}

func waitForGID(ctx context.Context, rt commandRuntime, capabilityName, gid string, deadline time.Time) map[string]any {
	state := map[string]any{"gid": gid, "capability": capabilityName, "reached": false, "status": ""}
	for time.Now().Before(deadline) {
		res, err := rt.Client.Call(ctx, "aria2.tellStatus", []any{gid})
		if err != nil {
			if capabilityName == "remove" || capabilityName == "remove_all" {
				state["reached"] = true
				state["status"] = "removed"
				return state
			}
			state["error"] = err.Error()
			time.Sleep(waitInterval(rt.Opts.WaitInterval))
			continue
		}
		status := extractStatus(res)
		state["status"] = status
		if reachedTargetStatus(capabilityName, status) {
			state["reached"] = true
			return state
		}
		time.Sleep(waitInterval(rt.Opts.WaitInterval))
	}
	state["timed_out"] = true
	return state
}

func waitInterval(interval time.Duration) time.Duration {
	if interval <= 0 {
		return 400 * time.Millisecond
	}
	if interval < 100*time.Millisecond {
		return 100 * time.Millisecond
	}
	return interval
}

func extractStatus(raw json.RawMessage) string {
	parsed := map[string]any{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ""
	}
	status, _ := parsed["status"].(string)
	return strings.TrimSpace(status)
}

func reachedTargetStatus(capabilityName, status string) bool {
	s := strings.TrimSpace(status)
	switch capabilityName {
	case "add_uri", "add_magnet", "add_torrent", "add_metalink":
		return s == "complete" || s == "error" || s == "removed"
	case "pause", "pause_all":
		return s == "paused"
	case "resume", "resume_all":
		return s == "active" || s == "waiting" || s == "complete"
	case "remove", "remove_all":
		return s == "removed"
	default:
		return s != ""
	}
}

func extractResultGIDs(result any) []string {
	ids := []string{}
	if direct, ok := result.(string); ok {
		direct = strings.TrimSpace(direct)
		if direct != "" {
			return []string{direct}
		}
	}
	obj, ok := result.(map[string]any)
	if !ok {
		return ids
	}
	if gid, ok := obj["gid"].(string); ok && strings.TrimSpace(gid) != "" {
		ids = append(ids, strings.TrimSpace(gid))
	}
	if arr, ok := obj["removed_gids"].([]any); ok {
		for _, item := range arr {
			if gid, ok := item.(string); ok && strings.TrimSpace(gid) != "" {
				ids = append(ids, strings.TrimSpace(gid))
			}
		}
	}
	return ids
}

func mergeDiagnostics(payload map[string]any, extra map[string]any) {
	diag, _ := payload["diagnostics"].(map[string]any)
	if diag == nil {
		diag = map[string]any{}
	}
	for k, v := range extra {
		diag[k] = v
	}
	payload["diagnostics"] = diag
}
