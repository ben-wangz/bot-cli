package capability

import (
	"context"
	"sort"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/ttsapi"
)

type Handler func(ctx context.Context, client *ttsapi.Client, req Request) (map[string]any, error)
type OfflineHandler func(req Request) (map[string]any, error)

type registryEntry struct {
	handler        Handler
	offlineHandler OfflineHandler
	readOnly       bool
}

var operationRegistry = map[string]registryEntry{
	"generate_speech":  {handler: runGenerateSpeech, readOnly: false},
	"suggest_voices":   {offlineHandler: runSuggestVoices, readOnly: true},
	"file_to_data_uri": {offlineHandler: runFileToDataURI, readOnly: true},
}

func Dispatch(ctx context.Context, client *ttsapi.Client, req Request) (map[string]any, error) {
	entry, ok := operationRegistry[req.Name]
	if !ok {
		return nil, apperr.New(apperr.CodeInvalidArgs, "operation not implemented yet: "+req.Name)
	}
	if entry.offlineHandler != nil {
		return entry.offlineHandler(req)
	}
	if client == nil {
		return nil, apperr.New(apperr.CodeConfig, "api client is required for this capability")
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
	meta := describeMeta(name)
	meta["read_only"] = entry.readOnly
	return meta, true
}

func describeMeta(name string) map[string]any {
	switch name {
	case "generate_speech":
		return map[string]any{
			"name":    "generate_speech",
			"summary": "Synthesize speech audio using MiMo-V2.5-TTS models.",
			"args": []map[string]any{
				{"name": "assistant_text", "required": true, "description": "Text to be synthesized and spoken (what to say)."},
				{"name": "model", "required": false, "description": "Model id (default: mimo-v2.5-tts)."},
				{"name": "user_text", "required": false, "description": "Style/context instruction (how to say it). Required by mimo-v2.5-tts-voicedesign."},
				{"name": "audio_format", "required": false, "description": "Requested output audio format (default: wav)."},
				{"name": "builtin_voice", "required": false, "description": "Built-in voice ID for mimo-v2.5-tts only. Not valid for voicedesign/voiceclone models."},
				{"name": "clone_voice_data_uri", "required": false, "description": "Voice clone input for mimo-v2.5-tts-voiceclone only. Data URI / Data URL format (RFC 2397): data:{mime};base64,{audio}."},
				{"name": "stream", "required": false, "description": "Enable streaming mode (true/false). audio_format=pcm16 is recommended for chunk stitching."},
				{"name": "output_dir", "required": false, "description": "Directory to save generated audio file."},
				{"name": "output_name", "required": false, "description": "Output file name override."},
			},
			"model_rules": []map[string]any{
				{
					"model":                "mimo-v2.5-tts",
					"assistant_text":       "required",
					"user_text":            "optional",
					"builtin_voice":        "optional",
					"clone_voice_data_uri": "not_allowed",
					"notes":                "Regular built-in voice synthesis model.",
				},
				{
					"model":                "mimo-v2.5-tts-voicedesign",
					"assistant_text":       "required",
					"user_text":            "required",
					"builtin_voice":        "not_allowed",
					"clone_voice_data_uri": "not_allowed",
					"notes":                "user_text carries voice identity/style description.",
				},
				{
					"model":                "mimo-v2.5-tts-voiceclone",
					"assistant_text":       "required",
					"user_text":            "optional",
					"builtin_voice":        "not_allowed",
					"clone_voice_data_uri": "required",
					"notes":                "clone_voice_data_uri should be Data URI / Data URL (RFC 2397).",
				},
			},
			"examples": []string{
				"tts-cli capability generate_speech --assistant-text 'Hello world' --model mimo-v2.5-tts --builtin-voice Chloe --audio-format wav",
				"tts-cli capability generate_speech --model mimo-v2.5-tts-voicedesign --assistant-text 'This is a sample.' --user-text 'Young female voice, warm and calm.'",
				"tts-cli capability generate_speech --model mimo-v2.5-tts-voiceclone --assistant-text 'Demo' --clone-voice-data-uri 'data:audio/mpeg;base64,...'",
			},
		}
	case "suggest_voices":
		return map[string]any{
			"name":    "suggest_voices",
			"summary": "Show suggested model and voice mapping without enforcing hard validation.",
			"args": []map[string]any{
				{"name": "model", "required": false, "description": "If set, return suggestions for one model only; otherwise return all models."},
			},
			"examples": []string{
				"tts-cli capability suggest_voices",
				"tts-cli capability suggest_voices --model mimo-v2.5-tts",
			},
		}
	case "file_to_data_uri":
		return map[string]any{
			"name":    "file_to_data_uri",
			"summary": "Convert a local audio file to Base64 Data URI with MIME detection from file content.",
			"args": []map[string]any{
				{"name": "file_path", "required": true, "description": "Local audio file path."},
				{"name": "include_data_uri", "required": false, "description": "If true, include full data_uri in result. Default false to avoid oversized output."},
			},
			"examples": []string{
				"tts-cli capability file_to_data_uri --file-path ../tests/artofwar_01_sun_64kb_trimmed_32k.mp3",
				"tts-cli capability file_to_data_uri --file-path ../tests/artofwar_01_sun_64kb_trimmed_32k.mp3 --include-data-uri true",
			},
		}
	default:
		return map[string]any{"name": name, "summary": "Unknown capability", "args": []map[string]any{}}
	}
}
