package cli

import (
	"fmt"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability"
)

func rootHelp() string {
	return `proxmox-cli - Agent-facing CLI for Proxmox operations

Usage:
  proxmox-cli [global options] <command> [args]

Commands:
  capability  Execute one capability
  workflow    Execute composed workflow
  console     Serial console script/interactive helpers
  auth        Auth inspect and diagnostics

Global options:
  --api-base <url>                 Proxmox API base URL
  --auth-scope <user|root-token|root>
  --auth-file <path>               JSON auth file
  --auth-user <value>              Auth identity
  --auth-password <value>          Password credential
  --auth-token <value>             API token credential
  --auth-ticket <value>            Session ticket credential
  --auth-csrf-token <value>        CSRF token for ticket mode
  --insecure-tls                   Skip TLS cert verification
  --timeout <seconds|duration>     Request timeout (default: 30s; set a longer timeout explicitly for ISO upload and other long-running operations)
  --wait                           Wait for asynchronous task completion
  --dry-run                        Print execution plan without mutation
  --output <json|yaml|table>       Output format (default: json)
  --debug                          Enable debug logs with redaction
  -h, --help                       Show help

Safety env vars (for VMID-scoped capabilities):
  PVE_ALLOWED_VMID_MIN             Allowed VMID lower bound (default: 1001)
  PVE_ALLOWED_VMID_MAX             Allowed VMID upper bound (default: 2000)

Examples:
  proxmox-cli --auth-file ./auth.json auth inspect
  proxmox-cli --api-base https://pve:8006/api2/json capability list_nodes
  proxmox-cli --output table workflow ubuntu24-serial-autoinstall
`
}

func capabilityHelp() string {
	var sb strings.Builder
	sb.WriteString(`proxmox-cli capability - execute one capability

Usage:
  proxmox-cli [global options] capability <name> [args]
  proxmox-cli [global options] capability <name> --help
  proxmox-cli [global options] capability describe <name>

Examples:
  proxmox-cli capability list_nodes
  proxmox-cli capability list_vms_by_node --node pve1
  proxmox-cli capability get_vm_status --node pve1 --vmid 120

Implemented capabilities:
`)
	for _, group := range capability.CapabilityGroups() {
		sb.WriteString("  ")
		sb.WriteString(group.Name)
		sb.WriteString(":\n")
		for _, name := range group.Capabilities {
			sb.WriteString("    - ")
			sb.WriteString(name)
			sb.WriteString("\n")
		}
	}
	sb.WriteString(`

Notes:
  - Use workflow commands for composed end-to-end paths.
  - Use capability <name> --help for required/optional args and examples.
  - --wait applies task-wait only for async capabilities; others are synchronous or self-polled.
`)
	return sb.String()
}

func capabilityDetailHelp(name string) (string, bool) {
	entry, ok := capability.LookupHelpMeta(name)
	if !ok {
		return "", false
	}
	meta, _ := capability.LookupMeta(name)
	var sb strings.Builder
	sb.WriteString("proxmox-cli capability ")
	sb.WriteString(name)
	sb.WriteString("\n\n")
	if strings.TrimSpace(entry.Summary) != "" {
		sb.WriteString(entry.Summary)
		sb.WriteString("\n\n")
	}
	sb.WriteString("Usage:\n")
	sb.WriteString("  proxmox-cli [global options] capability ")
	sb.WriteString(name)
	sb.WriteString(" [args]\n\n")
	sb.WriteString("Args:\n")
	sb.WriteString("  Required: ")
	if len(entry.RequiredArgs) == 0 {
		sb.WriteString("(none)")
	} else {
		sb.WriteString("--")
		sb.WriteString(strings.Join(entry.RequiredArgs, ", --"))
	}
	sb.WriteString("\n")
	sb.WriteString("  Optional: ")
	if len(entry.OptionalArgs) == 0 {
		sb.WriteString("(none)")
	} else {
		sb.WriteString("--")
		sb.WriteString(strings.Join(entry.OptionalArgs, ", --"))
	}
	sb.WriteString("\n\n")
	sb.WriteString("Execution:\n")
	sb.WriteString("  Async task: ")
	sb.WriteString(fmt.Sprintf("%t", meta.Async))
	sb.WriteString("\n")
	if strings.TrimSpace(meta.WaitSkipReason) != "" {
		sb.WriteString("  Wait note: ")
		sb.WriteString(meta.WaitSkipReason)
		sb.WriteString("\n")
	}
	if len(entry.Examples) > 0 {
		sb.WriteString("\nExamples:\n")
		for _, sample := range entry.Examples {
			sb.WriteString("  ")
			sb.WriteString(sample)
			sb.WriteString("\n")
		}
	}
	return sb.String(), true
}

func workflowHelp() string {
	return `proxmox-cli workflow - execute composed workflow

Usage:
  proxmox-cli [global options] workflow <name> [args]

Examples:
  proxmox-cli workflow ubuntu24-serial-autoinstall
  proxmox-cli --dry-run workflow ubuntu24-serial-autoinstall --vmid 120

Implemented workflows:
  bootstrap-bot-user-pool-acl --userid <user@realm> --poolid <poolid> [--password <plain>] [--if-exists fail|reuse] [--sdn-acl-path </sdn/...>] [--sdn-role <role>]
  provision-template-from-artifact --node <node> --target-vmid <id> --artifact-iso <storage:iso/file.iso> --install-timeout-seconds <n> [--resume-from none|serial_wait] [--pool <poolid>]

bootstrap-bot-user-pool-acl result:
  - creates or reuses user and pool with root-assisted bootstrap actions
  - grants ACLs: /pool/<poolid>=PVEAdmin, /=PVEAuditor, /storage=PVEDatastoreAdmin
  - grants SDN ACL (default /sdn/zones/localnetwork with PVEAdmin for SDN.Use)
  - returns workflow-standard JSON result (no local env file mutation)

provision-template-from-artifact result:
  - consumes prebuilt + preuploaded ISO artifact only (no upload in workflow)
  - optionally pins create_vm into a target pool via --pool
  - enforces vmid-not-exists check and mandatory cdrom attach/detach path
  - supports resume mode from serial wait step (--resume-from serial_wait, use with caution)
`
}

func consoleHelp() string {
	return `proxmox-cli console - serial and termproxy helpers

Usage:
  proxmox-cli [global options] console <script|interactive> [args]

Examples:
  proxmox-cli console script --vmid 120 --expect "login:"
  proxmox-cli console interactive --vmid 120
`
}

func authHelp() string {
	return `proxmox-cli auth - inspect auth source and scope

Usage:
  proxmox-cli [global options] auth <inspect|example-auth-file>

Examples:
  proxmox-cli --auth-file ./auth.json auth inspect
  proxmox-cli auth example-auth-file
`
}
