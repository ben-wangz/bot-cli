package cli

import (
	"os"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/output"
)

type GlobalOptions struct {
	APIBaseURL string
	APIKey     string
	Timeout    time.Duration
	Output     string
	OutputDir  string
	OutputName string
	Help       bool
}

func defaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		APIBaseURL: strings.TrimSpace(envOrDefault("TTS_API_BASE_URL", "https://api.xiaomimimo.com/v1")),
		APIKey:     strings.TrimSpace(os.Getenv("TTS_API_KEY")),
		Timeout:    300 * time.Second,
		Output:     output.FormatJSON,
		OutputDir:  strings.TrimSpace(os.Getenv("TTS_OUTPUT_DIR")),
		OutputName: strings.TrimSpace(os.Getenv("TTS_OUTPUT_NAME")),
	}
}

func envOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
