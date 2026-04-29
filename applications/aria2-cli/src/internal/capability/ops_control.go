package capability

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
)

func runPause(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	method := "aria2.pause"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forcePause"
	}
	res, err := client.Call(ctx, method, []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runPauseAll(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	method := "aria2.pauseAll"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forcePauseAll"
	}
	res, err := client.Call(ctx, method, []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runResume(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.unpause", []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runResumeAll(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.unpauseAll", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runRemove(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	method := "aria2.remove"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forceRemove"
	}
	res, err := client.Call(ctx, method, []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runRemoveAll(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	active, err := client.Call(ctx, "aria2.tellActive", []any{})
	if err != nil {
		return nil, err
	}
	waiting, err := client.Call(ctx, "aria2.tellWaiting", []any{0, 1000})
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(active, &items); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to decode active list", err)
	}
	var waitingItems []map[string]any
	if err := json.Unmarshal(waiting, &waitingItems); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to decode waiting list", err)
	}
	items = append(items, waitingItems...)
	removed := make([]string, 0, len(items))
	seen := map[string]bool{}
	method := "aria2.remove"
	if OptionalBool(req.Args, "force", false) {
		method = "aria2.forceRemove"
	}
	for _, item := range items {
		gid := strings.TrimSpace(stringValue(item["gid"]))
		if gid == "" || seen[gid] {
			continue
		}
		seen[gid] = true
		if _, callErr := client.Call(ctx, method, []any{gid}); callErr == nil {
			removed = append(removed, gid)
		}
	}
	return envelope(req, map[string]any{"removed_gids": removed, "count": len(removed)}, nil), nil
}

func runPurgeDownloadResult(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.purgeDownloadResult", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}
