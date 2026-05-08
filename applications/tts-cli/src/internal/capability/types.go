package capability

type Request struct {
	Name string
	Args map[string]string
}

type builtInVoice struct {
	Name     string
	ID       string
	Language string
	Gender   string
}

type modelRule struct {
	ModelID            string
	Summary            string
	BuiltinVoiceRule   string
	CloneVoiceDataRule string
	UserTextRule       string
}

var supportedModels = []modelRule{
	{
		ModelID:            "mimo-v2.5-tts",
		Summary:            "Built-in voice synthesis model for regular TTS generation.",
		BuiltinVoiceRule:   "Optional. Recommended to use built-in voice IDs.",
		CloneVoiceDataRule: "Not used.",
		UserTextRule:       "Optional style/context instruction.",
	},
	{
		ModelID:            "mimo-v2.5-tts-voicedesign",
		Summary:            "Voice design model driven by text description.",
		BuiltinVoiceRule:   "Not used.",
		CloneVoiceDataRule: "Not used.",
		UserTextRule:       "Required for voice style/identity description.",
	},
	{
		ModelID:            "mimo-v2.5-tts-voiceclone",
		Summary:            "Voice clone model using a reference audio sample.",
		BuiltinVoiceRule:   "Not used.",
		CloneVoiceDataRule: "Recommended: Data URI (RFC 2397): data:{mime};base64,{audio}",
		UserTextRule:       "Optional style/context instruction.",
	},
}

var builtInVoices = []builtInVoice{
	{Name: "MiMo Default", ID: "mimo_default", Language: "cluster-dependent", Gender: "-"},
	{Name: "冰糖", ID: "冰糖", Language: "Chinese", Gender: "Female"},
	{Name: "茉莉", ID: "茉莉", Language: "Chinese", Gender: "Female"},
	{Name: "苏打", ID: "苏打", Language: "Chinese", Gender: "Male"},
	{Name: "白桦", ID: "白桦", Language: "Chinese", Gender: "Male"},
	{Name: "Mia", ID: "Mia", Language: "English", Gender: "Female"},
	{Name: "Chloe", ID: "Chloe", Language: "English", Gender: "Female"},
	{Name: "Milo", ID: "Milo", Language: "English", Gender: "Male"},
	{Name: "Dean", ID: "Dean", Language: "English", Gender: "Male"},
}

var cloneVoiceDataURIMIMEWhitelist = []string{"audio/mpeg", "audio/mp3", "audio/wav"}
