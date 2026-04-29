package cli

import (
	"fmt"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/capability"
)

func rootHelp() string {
	return strings.TrimSpace(`aria2-cli

Usage:
  aria2-cli [global options] <command> [args]

Commands:
  capability   Run a single atomic capability
  workflow     Run a multi-step workflow
  help         Show help

Global options:
  --rpc-endpoint <url>   JSON-RPC endpoint (default http://127.0.0.1:6800/jsonrpc)
  --rpc-secret <secret>  RPC secret token (or env ARIA2_RPC_SECRET)
  --timeout <seconds>    Request timeout in seconds
  --output <format>      Output format (json)
` + "\n")
}

func capabilityHelp() string {
	list := capability.Names()
	return fmt.Sprintf("capability usage:\n  aria2-cli capability <name> [--key value]\n\nimplemented capabilities:\n  %s\n", strings.Join(list, "\n  "))
}

func workflowHelp() string {
	return `workflow usage:
  aria2-cli workflow <name> [--key value]

implemented workflows:
  queue_add_and_wait
  pause_resume_chain
  cleanup_completed
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
