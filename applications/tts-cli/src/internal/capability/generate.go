package capability

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/ttsapi"
)

func runGenerateSpeech(ctx context.Context, client *ttsapi.Client, req Request) (map[string]any, error) {
	startedAt := time.Now()
	assistantText, err := RequiredString(req.Args, "assistant_text")
	if err != nil {
		return nil, err
	}
	model := defaultString(req.Args["model"], "mimo-v2.5-tts")
	userText := strings.TrimSpace(req.Args["user_text"])
	audioFormat := defaultString(req.Args["audio_format"], "wav")
	builtinVoice := strings.TrimSpace(req.Args["builtin_voice"])
	cloneVoiceDataURI := strings.TrimSpace(req.Args["clone_voice_data_uri"])
	stream := OptionalBool(req.Args, "stream", false)
	if err := validateGenerateParams(model, userText, builtinVoice, cloneVoiceDataURI); err != nil {
		return nil, err
	}
	params := ttsapi.GenerateParams{
		Model:             model,
		AssistantText:     assistantText,
		UserText:          userText,
		AudioFormat:       audioFormat,
		BuiltinVoice:      builtinVoice,
		CloneVoiceDataURI: cloneVoiceDataURI,
		Stream:            stream,
	}
	result, err := client.Generate(ctx, params)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.AudioFormat) == "" {
		result.AudioFormat = audioFormat
	}
	outputDir := strings.TrimSpace(req.Args["output_dir"])
	outputName := strings.TrimSpace(req.Args["output_name"])
	if outputDir == "" {
		outputDir = strings.TrimSpace(req.Args["global_output_dir"])
	}
	if outputName == "" {
		outputName = strings.TrimSpace(req.Args["global_output_name"])
	}
	filePath, err := saveAudio(result.AudioBytes, result.AudioFormat, outputDir, outputName)
	if err != nil {
		return nil, err
	}
	warnings := buildWarnings(model, builtinVoice, cloneVoiceDataURI, stream, audioFormat)
	warnings = append(warnings, result.Warnings...)
	elapsedMS := time.Since(startedAt).Milliseconds()
	return map[string]any{
		"ok": true,
		"request": map[string]any{"capability": "generate_speech", "args": req.Args},
		"result": map[string]any{
			"output_file":   filePath,
			"response_id":   result.ResponseID,
			"audio_format":  result.AudioFormat,
			"stream":        stream,
			"bytes":         result.AudioByteSize,
			"chunk_count":   result.ChunkCount,
		},
		"diagnostics": map[string]any{
			"warnings":      warnings,
			"elapsed_ms":    elapsedMS,
			"audio_bytes":   result.AudioByteSize,
			"stream_chunks": result.ChunkCount,
		},
	}, nil
}

func validateGenerateParams(model string, userText string, builtinVoice string, cloneVoiceDataURI string) error {
	if !isSupportedModel(model) {
		return apperr.New(apperr.CodeInvalidArgs, "model must be one of: mimo-v2.5-tts, mimo-v2.5-tts-voicedesign, mimo-v2.5-tts-voiceclone")
	}
	if model != "mimo-v2.5-tts" && strings.TrimSpace(builtinVoice) != "" {
		return apperr.New(apperr.CodeInvalidArgs, "builtin_voice is only valid for mimo-v2.5-tts")
	}
	if model == "mimo-v2.5-tts-voicedesign" && strings.TrimSpace(userText) == "" {
		return apperr.New(apperr.CodeInvalidArgs, "user_text is required for mimo-v2.5-tts-voicedesign")
	}
	if model != "mimo-v2.5-tts-voiceclone" && strings.TrimSpace(cloneVoiceDataURI) != "" {
		return apperr.New(apperr.CodeInvalidArgs, "clone_voice_data_uri is only valid for mimo-v2.5-tts-voiceclone")
	}
	if model == "mimo-v2.5-tts-voiceclone" && strings.TrimSpace(cloneVoiceDataURI) == "" {
		return apperr.New(apperr.CodeInvalidArgs, "clone_voice_data_uri is required for mimo-v2.5-tts-voiceclone")
	}
	if strings.TrimSpace(builtinVoice) != "" && strings.TrimSpace(cloneVoiceDataURI) != "" {
		return apperr.New(apperr.CodeInvalidArgs, "builtin_voice and clone_voice_data_uri cannot both be set")
	}
	return nil
}

func isSupportedModel(model string) bool {
	for _, m := range supportedModels {
		if m.ModelID == strings.TrimSpace(model) {
			return true
		}
	}
	return false
}

func saveAudio(audioBytes []byte, audioFormat string, outputDir string, outputName string) (string, error) {
	if outputDir == "" {
		outputDir = "."
	}
	fileName := strings.TrimSpace(outputName)
	if fileName == "" {
		fileName = "tts-" + time.Now().UTC().Format("20060102-150405") + "." + normalizeExt(audioFormat)
	}
	if !strings.Contains(fileName, ".") {
		fileName += "." + normalizeExt(audioFormat)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to prepare output directory", err)
	}
	path := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(path, audioBytes, 0o644); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "failed to write output file", err)
	}
	return path, nil
}

func normalizeExt(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	if f == "" {
		return "wav"
	}
	return f
}

func buildWarnings(model string, builtinVoice string, cloneVoiceDataURI string, stream bool, audioFormat string) []string {
	warnings := make([]string, 0)
	if stream && strings.ToLower(strings.TrimSpace(audioFormat)) != "pcm16" {
		warnings = append(warnings, "streaming is usually used with audio_format=pcm16 for chunk stitching")
	}
	if model == "mimo-v2.5-tts" && strings.TrimSpace(builtinVoice) != "" && !isKnownBuiltInVoice(builtinVoice) {
		warnings = append(warnings, "builtin_voice is not in known suggestions; request was forwarded upstream")
	}
	if model == "mimo-v2.5-tts-voiceclone" && strings.TrimSpace(cloneVoiceDataURI) != "" {
		if !hasAllowedDataURIPrefix(cloneVoiceDataURI) {
			warnings = append(warnings, "clone_voice_data_uri MIME is outside suggested whitelist; request was forwarded upstream")
		}
		if !looksLikeDataURI(cloneVoiceDataURI) {
			warnings = append(warnings, "clone_voice_data_uri should follow Data URI / Data URL format (RFC 2397)")
		}
	}
	return warnings
}

func isKnownBuiltInVoice(voiceID string) bool {
	v := strings.TrimSpace(voiceID)
	for _, voice := range builtInVoices {
		if voice.ID == v {
			return true
		}
	}
	return false
}

func hasAllowedDataURIPrefix(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	for _, mime := range cloneVoiceDataURIMIMEWhitelist {
		prefix := "data:" + strings.ToLower(mime) + ";base64,"
		if strings.HasPrefix(v, prefix) {
			return true
		}
	}
	return false
}

func looksLikeDataURI(value string) bool {
	v := strings.TrimSpace(value)
	if !strings.HasPrefix(strings.ToLower(v), "data:") {
		return false
	}
	parts := strings.SplitN(v, ",", 2)
	if len(parts) != 2 {
		return false
	}
	header := strings.ToLower(parts[0])
	if !strings.Contains(header, ";base64") {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
	return err == nil
}

func defaultString(v string, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}
