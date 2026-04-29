package capability

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
)

func runAddURI(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	uri, err := RequiredString(req.Args, "uri")
	if err != nil {
		return nil, err
	}
	options, err := OptionalJSONObject(req.Args, "options")
	if err != nil {
		return nil, err
	}
	position, err := OptionalInt(req.Args, "position", -1)
	if err != nil {
		return nil, err
	}
	params := []any{[]string{uri}}
	if len(options) > 0 {
		params = append(params, options)
	}
	if position >= 0 {
		params = append(params, position)
	}
	res, err := client.Call(ctx, "aria2.addUri", params)
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runAddMagnet(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	uri, err := RequiredString(req.Args, "uri")
	if err != nil {
		return nil, err
	}
	req.Args["uri"] = uri
	return runAddURI(ctx, client, req)
}

func runAddTorrent(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	path, err := RequiredString(req.Args, "file")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to read torrent file", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	options, err := OptionalJSONObject(req.Args, "options")
	if err != nil {
		return nil, err
	}
	position, err := OptionalInt(req.Args, "position", -1)
	if err != nil {
		return nil, err
	}
	params := []any{encoded, []string{}}
	if len(options) > 0 {
		params = append(params, options)
	}
	if position >= 0 {
		if len(options) == 0 {
			params = append(params, map[string]any{})
		}
		params = append(params, position)
	}
	res, err := client.Call(ctx, "aria2.addTorrent", params)
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), map[string]any{"source_file": path}), nil
}

func runAddMetalink(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	path, err := RequiredString(req.Args, "file")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to read metalink file", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	options, err := OptionalJSONObject(req.Args, "options")
	if err != nil {
		return nil, err
	}
	position, err := OptionalInt(req.Args, "position", -1)
	if err != nil {
		return nil, err
	}
	params := []any{encoded}
	if len(options) > 0 {
		params = append(params, options)
	}
	if position >= 0 {
		if len(options) == 0 {
			params = append(params, map[string]any{})
		}
		params = append(params, position)
	}
	res, err := client.Call(ctx, "aria2.addMetalink", params)
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), map[string]any{"source_file": path}), nil
}
