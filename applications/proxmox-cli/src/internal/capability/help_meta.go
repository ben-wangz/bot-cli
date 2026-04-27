package capability

type HelpMeta struct {
	Summary      string   `json:"summary"`
	RequiredArgs []string `json:"required_args"`
	OptionalArgs []string `json:"optional_args"`
	Examples     []string `json:"examples"`
}

func LookupHelpMeta(name string) (HelpMeta, bool) {
	if meta, ok := capabilityHelpMetaCore[name]; ok {
		return meta, true
	}
	meta, ok := capabilityHelpMetaExtended[name]
	return meta, ok
}
