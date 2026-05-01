package workflow

import (
	"context"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/capability"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/imageapi"
)

func Execute(ctx context.Context, client *imageapi.Client, name string, args map[string]string) (map[string]any, error) {
	switch name {
	case "minimal_chain":
		return minimalChain(ctx, client, args)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow not implemented yet: "+name)
	}
}

func minimalChain(ctx context.Context, client *imageapi.Client, args map[string]string) (map[string]any, error) {
	firstPrompt, err := capability.RequiredString(args, "first_prompt")
	if err != nil {
		return nil, err
	}
	secondPrompt, err := capability.RequiredString(args, "second_prompt")
	if err != nil {
		return nil, err
	}
	firstArgs := map[string]string{"prompt": firstPrompt, "stream": "true"}
	first, err := capability.Dispatch(ctx, client, capability.Request{Name: "generate_image", Args: firstArgs})
	if err != nil {
		return nil, err
	}
	resultObj, _ := first["result"].(map[string]any)
	responseID, _ := resultObj["response_id"].(string)
	secondArgs := map[string]string{"prompt": secondPrompt, "stream": "false", "store": "true", "previous_response_id": responseID}
	second, err := capability.Dispatch(ctx, client, capability.Request{Name: "generate_image", Args: secondArgs})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok": true,
		"request": map[string]any{
			"workflow": "minimal_chain",
			"args":     args,
		},
		"result": map[string]any{
			"first":  first["result"],
			"second": second["result"],
		},
		"diagnostics": map[string]any{},
	}, nil
}
