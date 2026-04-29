package capability

import (
	"context"
	"encoding/json"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
)

func runTellStatus(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	gid, err := RequiredString(req.Args, "gid")
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellStatus", []any{gid})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runListActive(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.tellActive", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runListWaiting(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	offset, err := OptionalInt(req.Args, "offset", 0)
	if err != nil {
		return nil, err
	}
	limit, err := OptionalInt(req.Args, "limit", 100)
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellWaiting", []any{offset, limit})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runListStopped(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	offset, err := OptionalInt(req.Args, "offset", 0)
	if err != nil {
		return nil, err
	}
	limit, err := OptionalInt(req.Args, "limit", 100)
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.tellStopped", []any{offset, limit})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runGetGlobalStat(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.getGlobalStat", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runRPCCall(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	method, err := RequiredString(req.Args, "method")
	if err != nil {
		return nil, err
	}
	params, err := OptionalJSONArray(req.Args, "params")
	if err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}
