package cli

import (
	"os"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/output"
)

type GlobalOptions struct {
	APIBaseURL string
	APIKey     string
	Timeout    time.Duration
	Output     string
	Help       bool
}

func defaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		APIBaseURL: strings.TrimSpace(os.Getenv("IMAGE_API_BASE_URL")),
		APIKey:     strings.TrimSpace(os.Getenv("IMAGE_API_KEY")),
		Timeout:    60 * time.Second,
		Output:     output.FormatJSON,
	}
}
