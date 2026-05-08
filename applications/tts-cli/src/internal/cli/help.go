package cli

import (
	"fmt"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/capability"
)

func rootHelp() string {
	return strings.TrimSpace(`tts-cli - Agent-facing TTS capability bridge for bot-cli

Usage:
  tts-cli [global options] <command> [args]

Commands:
  capability   Execute one capability

Global options:
  --api-base-url <url>  API base URL (default: https://api.xiaomimimo.com/v1 or env TTS_API_BASE_URL)
  --api-key <token>     API key (or env TTS_API_KEY)
  --timeout <seconds>   Request timeout in seconds (default: 300)
  --output-dir <path>   Output directory (or env TTS_OUTPUT_DIR)
  --output-name <name>  Output file name (or env TTS_OUTPUT_NAME)
  --output <format>     Output format (json)
  --help, -h            Show help
` + "\n")
}

func capabilityHelp() string {
	list := capability.Names()
	return fmt.Sprintf("capability usage:\n  tts-cli capability <name> [--key value]\n  tts-cli capability <name> --help\n  tts-cli capability describe [<name>]\n\nimplemented capabilities:\n  %s\n", strings.Join(list, "\n  "))
}

func capabilityDetailHelp(name string) (string, bool) {
	detail, ok := capability.Describe(name)
	if !ok {
		return "", false
	}
	var sb strings.Builder
	sb.WriteString("tts-cli capability ")
	sb.WriteString(name)
	sb.WriteString("\n\n")
	if summary, _ := detail["summary"].(string); strings.TrimSpace(summary) != "" {
		sb.WriteString(summary)
		sb.WriteString("\n\n")
	}
	sb.WriteString("Usage:\n")
	sb.WriteString("  tts-cli [global options] capability ")
	sb.WriteString(name)
	sb.WriteString(" [args]\n\n")
	sb.WriteString("Args:\n")
	args, _ := detail["args"].([]map[string]any)
	if len(args) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, arg := range args {
			required, _ := arg["required"].(bool)
			label := "optional"
			if required {
				label = "required"
			}
			sb.WriteString("  --")
			sb.WriteString(asString(arg["name"]))
			sb.WriteString(" [")
			sb.WriteString(label)
			sb.WriteString("] ")
			sb.WriteString(asString(arg["description"]))
			sb.WriteString("\n")
		}
	}
	if examples, ok := detail["examples"].([]string); ok && len(examples) > 0 {
		sb.WriteString("\nExamples:\n")
		for _, sample := range examples {
			sb.WriteString("  ")
			sb.WriteString(sample)
			sb.WriteString("\n")
		}
	}
	return sb.String(), true
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
