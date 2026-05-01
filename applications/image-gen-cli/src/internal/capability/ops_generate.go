package capability

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/imageapi"
)

func runGenerateImage(ctx context.Context, client *imageapi.Client, req Request) (map[string]any, error) {
	prompt, err := RequiredString(req.Args, "prompt")
	if err != nil {
		return nil, err
	}
	stream := OptionalBool(req.Args, "stream", true)
	store := OptionalBool(req.Args, "store", false)
	compression, err := OptionalInt(req.Args, "output_compression", 85)
	if err != nil {
		return nil, err
	}
	params := imageapi.GenerateParams{
		Prompt:             prompt,
		Model:              defaultString(req.Args["model"], "gpt-5.5"),
		ImageModel:         defaultString(req.Args["image_model"], "gpt-image-2"),
		Stream:             stream,
		Store:              store,
		PreviousResponseID: strings.TrimSpace(req.Args["previous_response_id"]),
		Size:               defaultString(req.Args["size"], "1024x1024"),
		Quality:            defaultString(req.Args["quality"], "auto"),
		OutputFormat:       defaultString(req.Args["output_format"], "png"),
		OutputCompression:  compression,
		Background:         defaultString(req.Args["background"], "auto"),
	}
	if err := validateParams(params); err != nil {
		return nil, err
	}
	result, err := client.Generate(ctx, params)
	if err != nil {
		return nil, err
	}
	filePath, err := saveImage(result.FinalImageBase64, result.OutputFormat)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok": true,
		"request": map[string]any{
			"capability": "generate_image",
			"args":       req.Args,
		},
		"result": map[string]any{
			"output_file":   filePath,
			"response_id":   result.ResponseID,
			"output_format": result.OutputFormat,
			"stream":        stream,
		},
		"diagnostics": map[string]any{
			"preview_count": result.PreviewCount,
		},
	}, nil
}

func validateParams(p imageapi.GenerateParams) error {
	if !inSet(p.Size, []string{"1024x1024", "1024x1536", "1536x1024", "auto"}) {
		return apperr.New(apperr.CodeInvalidArgs, "size must be one of: 1024x1024, 1024x1536, 1536x1024, auto")
	}
	if !inSet(p.Quality, []string{"auto", "high", "medium", "low"}) {
		return apperr.New(apperr.CodeInvalidArgs, "quality must be one of: auto, high, medium, low")
	}
	if !inSet(p.OutputFormat, []string{"png", "jpeg", "webp"}) {
		return apperr.New(apperr.CodeInvalidArgs, "output_format must be one of: png, jpeg, webp")
	}
	if !inSet(p.Background, []string{"auto", "opaque", "transparent"}) {
		return apperr.New(apperr.CodeInvalidArgs, "background must be one of: auto, opaque, transparent")
	}
	if p.OutputCompression < 0 || p.OutputCompression > 100 {
		return apperr.New(apperr.CodeInvalidArgs, "output_compression must be between 0 and 100")
	}
	if p.OutputFormat == "png" && p.OutputCompression != 85 {
		return apperr.New(apperr.CodeInvalidArgs, "output_compression is only valid for jpeg or webp")
	}
	return nil
}

func inSet(target string, allowed []string) bool {
	for _, item := range allowed {
		if strings.TrimSpace(target) == item {
			return true
		}
	}
	return false
}

func defaultString(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func saveImage(imageBase64 string, outputFormat string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeRPC, "failed to decode image base64", err)
	}
	fileName := "image-gen-" + time.Now().UTC().Format("20060102-150405") + "." + outputFormat
	path := filepath.Join(".", fileName)
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to write output file", err)
	}
	return path, nil
}
