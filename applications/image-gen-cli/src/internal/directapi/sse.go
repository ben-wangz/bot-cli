package directapi

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
)

func decodeStreamingResponse(r io.Reader) (GenerateResult, error) {
	result := GenerateResult{}
	reader := bufio.NewReader(r)
	dataLines := []string{}
	seenTypes := map[string]bool{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return GenerateResult{}, apperr.Wrap(apperr.CodeRPC, "failed to read direct streaming response", err)
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(trimmed, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(trimmed, "data:")))
		}
		if strings.TrimSpace(trimmed) == "" {
			parsed, done, parsedType, parseErr := handleDirectSSEDataBlock(dataLines, &result)
			dataLines = dataLines[:0]
			if parseErr != nil {
				return GenerateResult{}, parseErr
			}
			if parsedType != "" {
				seenTypes[parsedType] = true
			}
			if done {
				return parsed, nil
			}
		}
		if err == io.EOF {
			break
		}
	}
	if len(dataLines) > 0 {
		parsed, done, parsedType, parseErr := handleDirectSSEDataBlock(dataLines, &result)
		if parseErr != nil {
			return GenerateResult{}, parseErr
		}
		if parsedType != "" {
			seenTypes[parsedType] = true
		}
		if done {
			return parsed, nil
		}
	}
	return GenerateResult{}, apperr.New(apperr.CodeRPC, "direct streaming completed without final image result; seen events: "+strings.Join(sortedKeys(seenTypes), ","))
}

func handleDirectSSEDataBlock(dataLines []string, agg *GenerateResult) (GenerateResult, bool, string, error) {
	if len(dataLines) == 0 {
		return GenerateResult{}, false, "", nil
	}
	data := strings.TrimSpace(strings.Join(dataLines, "\n"))
	if data == "" || data == "[DONE]" {
		return GenerateResult{}, false, "", nil
	}
	obj := map[string]any{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return GenerateResult{}, false, "invalid_json_data", nil
	}
	typeName := asString(obj["type"])
	switch typeName {
	case "image_generation.partial_image":
		if asString(obj["b64_json"]) != "" {
			agg.PreviewCount++
		}
		return GenerateResult{}, false, typeName, nil
	case "image_generation.result", "image_generation.completed":
		imageBase64 := asString(obj["b64_json"])
		if imageBase64 == "" {
			return GenerateResult{}, false, typeName, nil
		}
		if _, err := base64.StdEncoding.DecodeString(imageBase64); err != nil {
			return GenerateResult{}, false, typeName, apperr.Wrap(apperr.CodeRPC, "final image result is not valid base64", err)
		}
		return GenerateResult{
			ResponseID:       asString(obj["id"]),
			FinalImageBase64: imageBase64,
			OutputFormat:     normalizedFormat(asString(obj["output_format"])),
			PreviewCount:     agg.PreviewCount,
		}, true, typeName, nil
	case "error":
		message := asString(obj["message"])
		if message == "" {
			message = "upstream reported error during direct streaming"
		}
		return GenerateResult{}, false, typeName, apperr.New(apperr.CodeRPC, message)
	case "":
		if imageBase64 := asString(obj["b64_json"]); imageBase64 != "" {
			if _, err := base64.StdEncoding.DecodeString(imageBase64); err != nil {
				return GenerateResult{}, false, "inline_result", apperr.Wrap(apperr.CodeRPC, "final image result is not valid base64", err)
			}
			return GenerateResult{
				ResponseID:       asString(obj["id"]),
				FinalImageBase64: imageBase64,
				OutputFormat:     normalizedFormat(asString(obj["output_format"])),
				PreviewCount:     agg.PreviewCount,
			}, true, "inline_result", nil
		}
		return GenerateResult{}, false, "missing_type", nil
	default:
		return GenerateResult{}, false, typeName, nil
	}
}

func sortedKeys(set map[string]bool) []string {
	if len(set) == 0 {
		return []string{"none"}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
