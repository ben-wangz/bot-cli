package capability

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/directapi"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/toolsapi"
)

func runGenerateImage(ctx context.Context, toolsClient *toolsapi.Client, directClient *directapi.Client, req Request) (map[string]any, error) {
	prompt, err := RequiredString(req.Args, "prompt")
	if err != nil {
		return nil, err
	}
	method := defaultString(req.Args["method"], "direct")
	if !inSet(method, []string{"direct", "tools"}) {
		return nil, apperr.New(apperr.CodeInvalidArgs, "method must be one of: direct, tools")
	}
	stream := OptionalBool(req.Args, "stream", true)
	store := OptionalBool(req.Args, "store", false)
	compression, err := OptionalInt(req.Args, "output_compression", -1)
	if err != nil {
		return nil, err
	}
	baseParams := sharedParams{
		Prompt:             prompt,
		Method:             method,
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
	if err := validateParams(baseParams, req.Args); err != nil {
		return nil, err
	}
	result := generatedResult{}
	if method == "tools" {
		toolResult, genErr := toolsClient.Generate(ctx, toolsapi.GenerateParams{
			Prompt:             baseParams.Prompt,
			Model:              baseParams.Model,
			ImageModel:         baseParams.ImageModel,
			Stream:             baseParams.Stream,
			Store:              baseParams.Store,
			PreviousResponseID: baseParams.PreviousResponseID,
			Size:               baseParams.Size,
			Quality:            baseParams.Quality,
			OutputFormat:       baseParams.OutputFormat,
			OutputCompression:  baseParams.OutputCompression,
			Background:         baseParams.Background,
		})
		if genErr != nil {
			if strings.TrimSpace(baseParams.PreviousResponseID) != "" && isChainingUnsupported(genErr) {
				return nil, apperr.New(apperr.CodeRPC, "upstream does not support previous_response_id/store chaining; retry without chaining")
			}
			return nil, genErr
		}
		result = generatedResult(toolResult)
	} else {
		directResult, genErr := directClient.Generate(ctx, directapi.GenerateParams{
			Prompt:            baseParams.Prompt,
			ImageModel:        baseParams.ImageModel,
			Stream:            baseParams.Stream,
			Size:              baseParams.Size,
			Quality:           baseParams.Quality,
			OutputFormat:      baseParams.OutputFormat,
			OutputCompression: baseParams.OutputCompression,
			Background:        baseParams.Background,
		})
		if genErr != nil {
			return nil, genErr
		}
		result = generatedResult(directResult)
	}
	outputDir := strings.TrimSpace(req.Args["output_dir"])
	outputName := strings.TrimSpace(req.Args["output_name"])
	if outputDir == "" {
		outputDir = strings.TrimSpace(req.Args["global_output_dir"])
	}
	if outputName == "" {
		outputName = strings.TrimSpace(req.Args["global_output_name"])
	}
	filePath, err := saveImage(result.FinalImageBase64, result.OutputFormat, outputDir, outputName)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok": true,
		"request": map[string]any{
			"capability": "generate_image",
			"args":       req.Args,
			"method":     method,
		},
		"result": map[string]any{
			"output_file":   filePath,
			"response_id":   result.ResponseID,
			"output_format": result.OutputFormat,
			"stream":        stream,
		},
		"diagnostics": map[string]any{
			"preview_count": result.PreviewCount,
			"stream":        stream,
			"deprecated_fields": []string{"preview_count"},
		},
	}, nil
}

type sharedParams struct {
	Prompt             string
	Method             string
	Model              string
	ImageModel         string
	Stream             bool
	Store              bool
	PreviousResponseID string
	Size               string
	Quality            string
	OutputFormat       string
	OutputCompression  int
	Background         string
}

type generatedResult struct {
	ResponseID       string
	FinalImageBase64 string
	OutputFormat     string
	PreviewCount     int
}

func validateParams(p sharedParams, args map[string]string) error {
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
	if p.OutputCompression >= 0 && (p.OutputCompression < 0 || p.OutputCompression > 100) {
		return apperr.New(apperr.CodeInvalidArgs, "output_compression must be between 0 and 100")
	}
	if p.OutputCompression >= 0 && p.OutputFormat == "png" {
		return apperr.New(apperr.CodeInvalidArgs, "output_compression is only valid for jpeg or webp")
	}
	if p.Method == "direct" {
		if strings.TrimSpace(args["previous_response_id"]) != "" {
			return apperr.New(apperr.CodeInvalidArgs, "previous_response_id requires --method tools")
		}
		if strings.TrimSpace(args["store"]) != "" {
			return apperr.New(apperr.CodeInvalidArgs, "store requires --method tools")
		}
		if strings.TrimSpace(args["model"]) != "" {
			return apperr.New(apperr.CodeInvalidArgs, "model requires --method tools; use --image-model for direct method")
		}
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

func isChainingUnsupported(err error) bool {
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "previous_response_id") || strings.Contains(msg, "store") || strings.Contains(msg, "not support") || strings.Contains(msg, "unsupported")
}

func saveImage(imageBase64 string, outputFormat string, outputDir string, outputName string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeRPC, "failed to decode image base64", err)
	}
	if outputDir == "" {
		outputDir = "."
	}
	fileName := strings.TrimSpace(outputName)
	if fileName == "" {
		fileName = "image-gen-" + time.Now().UTC().Format("20060102-150405") + "." + outputFormat
	}
	if !strings.Contains(fileName, ".") {
		fileName = fileName + "." + outputFormat
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to prepare output directory", err)
	}
	path := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to write output file", err)
	}
	return path, nil
}
