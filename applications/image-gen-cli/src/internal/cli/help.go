package cli

import (
	"fmt"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/capability"
)

func rootHelp() string {
	return strings.TrimSpace(`image-gen-cli

Usage:
  image-gen-cli [global options] <command> [args]

Commands:
  capability   Run a single atomic capability
  workflow     Run a multi-step workflow
  help         Show help

Global options:
  --api-base-url <url>  API base url (or env IMAGE_API_BASE_URL)
  --api-key <token>     API key (or env IMAGE_API_KEY)
  --timeout <seconds>   Request timeout in seconds
  --output <format>     Output format (json)
` + "\n")
}

func capabilityHelp() string {
	list := capability.Names()
	return fmt.Sprintf("capability usage:\n  image-gen-cli capability <name> [--key value]\n  image-gen-cli capability describe [<name>]\n\nimplemented capabilities:\n  %s\n", strings.Join(list, "\n  "))
}

func workflowHelp() string {
	return `workflow usage:
  image-gen-cli workflow <name> [--key value]

implemented workflows:
  minimal_chain
`
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}
