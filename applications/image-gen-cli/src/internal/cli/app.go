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
	Method     string
	Timeout    time.Duration
	Output     string
	OutputDir  string
	OutputName string
	Help       bool
}

func defaultGlobalOptions() GlobalOptions {
	opts := GlobalOptions{
		APIBaseURL: strings.TrimSpace(os.Getenv("IMAGE_API_BASE_URL")),
		APIKey:     strings.TrimSpace(os.Getenv("IMAGE_API_KEY")),
		Method:     strings.TrimSpace(os.Getenv("IMAGE_METHOD")),
		Timeout:    300 * time.Second,
		Output:     output.FormatJSON,
		OutputDir:  strings.TrimSpace(os.Getenv("IMAGE_OUTPUT_DIR")),
		OutputName: strings.TrimSpace(os.Getenv("IMAGE_OUTPUT_NAME")),
	}
	if opts.Method == "" {
		opts.Method = "direct"
	}
	return opts
}

func isValidMethod(method string) bool {
	return method == "direct" || method == "tools"
}
