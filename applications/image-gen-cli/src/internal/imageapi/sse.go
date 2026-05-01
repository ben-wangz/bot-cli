package imageapi

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
)

func parseSSEStream(r io.Reader) (GenerateResult, error) {
	result := GenerateResult{}
	reader := bufio.NewReader(r)
	dataLines := []string{}
	finalItem := map[string]any{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return GenerateResult{}, apperr.Wrap(apperr.CodeRPC, "failed to read streaming response", err)
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(trimmed, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(trimmed, "data:")))
		}
		if strings.TrimSpace(trimmed) == "" {
			parsed, done, parseErr := handleSSEDataBlock(dataLines, &result, finalItem)
			dataLines = dataLines[:0]
			if parseErr != nil {
				return GenerateResult{}, parseErr
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
		parsed, done, parseErr := handleSSEDataBlock(dataLines, &result, finalItem)
		if parseErr != nil {
			return GenerateResult{}, parseErr
		}
		if done {
			return parsed, nil
		}
	}
	return GenerateResult{}, apperr.New(apperr.CodeRPC, "streaming completed without final image result")
}

func handleSSEDataBlock(dataLines []string, agg *GenerateResult, finalItem map[string]any) (GenerateResult, bool, error) {
	if len(dataLines) == 0 {
		return GenerateResult{}, false, nil
	}
	data := strings.TrimSpace(strings.Join(dataLines, "\n"))
	if data == "" || data == "[DONE]" {
		return GenerateResult{}, false, nil
	}
	obj := map[string]any{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return GenerateResult{}, false, nil
	}
	typeName := asString(obj["type"])
	if typeName == "response.image_generation_call.partial_image" {
		if asString(obj["partial_image_b64"]) != "" {
			agg.PreviewCount++
		}
		return GenerateResult{}, false, nil
	}
	if typeName == "response.output_item.done" {
		item, _ := obj["item"].(map[string]any)
		if asString(item["type"]) == "image_generation_call" && asString(item["result"]) != "" {
			for k, v := range item {
				finalItem[k] = v
			}
		}
		return GenerateResult{}, false, nil
	}
	if typeName == "response.completed" {
		responseObj, _ := obj["response"].(map[string]any)
		parsed, err := extractFinalResult(responseObj)
		if err != nil && len(finalItem) > 0 {
			fallback := map[string]any{"id": asString(responseObj["id"]), "output": []any{finalItem}}
			parsed, err = extractFinalResult(fallback)
		}
		if err != nil {
			return GenerateResult{}, false, nil
		}
		parsed.PreviewCount = agg.PreviewCount
		return parsed, true, nil
	}
	if typeName == "response.failed" {
		return GenerateResult{}, false, apperr.New(apperr.CodeRPC, "upstream reported response.failed")
	}
	return GenerateResult{}, false, nil
}
