package workflow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/capability"
)

func Execute(ctx context.Context, client *aria2rpc.Client, name string, args map[string]string) (map[string]any, error) {
	switch name {
	case "queue_add_and_wait":
		return queueAddAndWait(ctx, client, args)
	case "pause_resume_chain":
		return pauseResumeChain(ctx, client, args)
	case "cleanup_completed":
		return cleanupCompleted(ctx, client, args)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow not implemented yet: "+name)
	}
}

func queueAddAndWait(ctx context.Context, client *aria2rpc.Client, args map[string]string) (map[string]any, error) {
	uri, err := capability.RequiredString(args, "uri")
	if err != nil {
		return nil, err
	}
	timeoutSec, err := capability.OptionalInt(args, "wait_timeout", 300)
	if err != nil {
		return nil, err
	}
	intervalSec, err := capability.OptionalInt(args, "wait_interval", 2)
	if err != nil {
		return nil, err
	}
	addResultRaw, err := client.Call(ctx, "aria2.addUri", []any{[]string{uri}})
	if err != nil {
		return nil, err
	}
	var gid string
	if err := json.Unmarshal(addResultRaw, &gid); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to parse addUri result", err)
	}
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	status := ""
	last := map[string]any{}
	for time.Now().Before(deadline) {
		res, callErr := client.Call(ctx, "aria2.tellStatus", []any{gid})
		if callErr != nil {
			return nil, callErr
		}
		if err := json.Unmarshal(res, &last); err != nil {
			return nil, apperr.Wrap(apperr.CodeInternal, "failed to parse tellStatus", err)
		}
		if s, ok := last["status"].(string); ok {
			status = s
			if s == "complete" || s == "error" || s == "removed" {
				break
			}
		}
		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
	return map[string]any{
		"ok":          true,
		"request":     map[string]any{"workflow": "queue_add_and_wait", "args": args},
		"result":      map[string]any{"gid": gid, "final_status": status, "download": last},
		"diagnostics": map[string]any{"timed_out": status == ""},
	}, nil
}

func pauseResumeChain(ctx context.Context, client *aria2rpc.Client, args map[string]string) (map[string]any, error) {
	gid, err := capability.RequiredString(args, "gid")
	if err != nil {
		return nil, err
	}
	if _, err := client.Call(ctx, "aria2.pause", []any{gid}); err != nil {
		return nil, err
	}
	if _, err := client.Call(ctx, "aria2.unpause", []any{gid}); err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellStatus", []any{gid})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":          true,
		"request":     map[string]any{"workflow": "pause_resume_chain", "args": args},
		"result":      map[string]any{"gid": gid, "status": json.RawMessage(res)},
		"diagnostics": map[string]any{},
	}, nil
}

func cleanupCompleted(ctx context.Context, client *aria2rpc.Client, args map[string]string) (map[string]any, error) {
	offset, err := capability.OptionalInt(args, "offset", 0)
	if err != nil {
		return nil, err
	}
	limit, err := capability.OptionalInt(args, "limit", 200)
	if err != nil {
		return nil, err
	}
	stoppedRaw, err := client.Call(ctx, "aria2.tellStopped", []any{offset, limit})
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(stoppedRaw, &items); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to parse stopped list", err)
	}
	purged := []string{}
	for _, item := range items {
		if gid, ok := item["gid"].(string); ok && gid != "" {
			if _, callErr := client.Call(ctx, "aria2.removeDownloadResult", []any{gid}); callErr == nil {
				purged = append(purged, gid)
			}
		}
	}
	return map[string]any{
		"ok":          true,
		"request":     map[string]any{"workflow": "cleanup_completed", "args": args},
		"result":      map[string]any{"purged_gids": purged, "count": len(purged)},
		"diagnostics": map[string]any{},
	}, nil
}
