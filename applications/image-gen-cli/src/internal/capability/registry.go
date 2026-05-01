package capability

import (
	"context"
	"sort"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/imageapi"
)

type Request struct {
	Name string
	Args map[string]string
}

type Handler func(ctx context.Context, client *imageapi.Client, req Request) (map[string]any, error)

type registryEntry struct {
	handler  Handler
	readOnly bool
}

var operationRegistry = map[string]registryEntry{
	"generate_image": {handler: runGenerateImage},
}

func Dispatch(ctx context.Context, client *imageapi.Client, req Request) (map[string]any, error) {
	entry, ok := operationRegistry[req.Name]
	if !ok {
		return nil, apperr.New(apperr.CodeInvalidArgs, "operation not implemented yet: "+req.Name)
	}
	return entry.handler(ctx, client, req)
}

func Names() []string {
	names := make([]string, 0, len(operationRegistry))
	for name := range operationRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Describe(name string) (map[string]any, bool) {
	entry, ok := operationRegistry[name]
	if !ok {
		return nil, false
	}
	return map[string]any{"name": name, "read_only": entry.readOnly, "args": capabilityArgs(name)}, true
}

func capabilityArgs(name string) []map[string]any {
	if name != "generate_image" {
		return []map[string]any{}
	}
	return []map[string]any{
		{"name": "prompt", "required": true, "description": "Text prompt for image generation."},
		{"name": "stream", "required": false, "description": "Use streaming mode (true/false)."},
		{"name": "store", "required": false, "description": "Request response storage for chaining."},
		{"name": "previous_response_id", "required": false, "description": "Previous response id for chaining."},
		{"name": "size", "required": false, "description": "Image size (1024x1024, 1024x1536, 1536x1024, auto)."},
		{"name": "quality", "required": false, "description": "Image quality (auto, high, medium, low)."},
		{"name": "output_format", "required": false, "description": "Output format (png, jpeg, webp)."},
		{"name": "output_compression", "required": false, "description": "Output compression for jpeg/webp (0-100)."},
		{"name": "background", "required": false, "description": "Background mode (auto, opaque, transparent)."},
	}
}
