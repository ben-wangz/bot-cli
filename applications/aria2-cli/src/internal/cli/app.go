package cli

import (
	"os"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/output"
)

type GlobalOptions struct {
	RPCEndpoint string
	RPCSecret   string
	Timeout     time.Duration
	Output      string
	Help        bool
}

func defaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		RPCEndpoint: "http://127.0.0.1:6800/jsonrpc",
		RPCSecret:   strings.TrimSpace(os.Getenv("ARIA2_RPC_SECRET")),
		Timeout:     30 * time.Second,
		Output:      output.FormatJSON,
	}
}
