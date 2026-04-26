package cli

import (
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/auth"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/output"
)

type GlobalOptions struct {
	APIBase     string
	AuthScope   string
	AuthFile    string
	AuthUser    string
	AuthPass    string
	AuthToken   string
	AuthTicket  string
	AuthCSRF    string
	InsecureTLS bool
	Timeout     time.Duration
	Wait        bool
	DryRun      bool
	Output      string
	Debug       bool
	Help        bool
}

func defaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		APIBase:   "https://127.0.0.1:8006/api2/json",
		AuthScope: auth.ScopeUser,
		Timeout:   30 * time.Second,
		Output:    output.FormatJSON,
	}
}
