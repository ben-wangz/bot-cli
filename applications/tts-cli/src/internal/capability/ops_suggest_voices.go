package capability

import "strings"

func runSuggestVoices(req Request) (map[string]any, error) {
	model := strings.TrimSpace(req.Args["model"])
	models := make([]map[string]any, 0, len(supportedModels))
	rules := make([]map[string]any, 0, len(supportedModels))
	for _, m := range supportedModels {
		if model != "" && model != m.ModelID {
			continue
		}
		models = append(models, map[string]any{"id": m.ModelID, "summary": m.Summary})
		rules = append(rules, map[string]any{
			"model":                 m.ModelID,
			"builtin_voice_rule":    m.BuiltinVoiceRule,
			"clone_voice_data_rule": m.CloneVoiceDataRule,
			"user_text_rule":        m.UserTextRule,
		})
	}
	builtin := make([]map[string]any, 0, len(builtInVoices))
	for _, v := range builtInVoices {
		builtin = append(builtin, map[string]any{"name": v.Name, "id": v.ID, "language": v.Language, "gender": v.Gender})
	}
	return map[string]any{
		"ok": true,
		"request": map[string]any{
			"capability": "suggest_voices",
			"args":       req.Args,
		},
		"result": map[string]any{
			"models":                 models,
			"voice_rules":            rules,
			"builtin_voices":         builtin,
			"clone_voice_data_uri":   "Data URI / Data URL (RFC 2397): data:{mime};base64,{audio}",
			"allowed_clone_mime_types": cloneVoiceDataURIMIMEWhitelist,
		},
		"diagnostics": map[string]any{},
	}, nil
}
