package cli

func rootHelp() string {
	return `proxmox-cli - Agent-facing CLI for Proxmox operations

Usage:
  proxmox-cli [global options] <command> [args]

Commands:
  action      Execute one action (A01-A43 roadmap)
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
  --timeout <seconds|duration>     Request timeout (default: 30s)
  --wait                           Wait for asynchronous task completion
  --dry-run                        Print execution plan without mutation
  --output <json|yaml|table>       Output format (default: json)
  --debug                          Enable debug logs with redaction
  -h, --help                       Show help

Safety env vars (for mutating actions):
  PVE_ALLOWED_VMID_MIN             Allowed VMID lower bound (default: 1001)
  PVE_ALLOWED_VMID_MAX             Allowed VMID upper bound (default: 2000)

Examples:
  proxmox-cli --auth-file ./auth.json auth inspect
  proxmox-cli --api-base https://pve:8006/api2/json action list_nodes
  proxmox-cli --output table workflow ubuntu24-serial-autoinstall
`
}

func actionHelp() string {
	return `proxmox-cli action - execute one action

Usage:
  proxmox-cli [global options] action <name> [action args]

Examples:
  proxmox-cli action list_nodes
  proxmox-cli action list_vms_by_node --node pve1
  proxmox-cli action get_vm_status --node pve1 --vmid 120

Phase 1 implemented actions:
  list_nodes
  list_cluster_resources [--type vm|storage|node]
  list_vms_by_node --node <node>
  get_vm_config --node <node> --vmid <vmid>
  get_effective_permissions [--path /vms]
  get_task_status --node <node> --upid <UPID>
  get_next_vmid
  get_vm_status --node <node> --vmid <vmid>
  list_tasks_by_vmid --node <node> --vmid <vmid> [--source active]

Phase 2 implemented actions:
  clone_template --node <node> --source-vmid <id> --target-vmid <id> [--name <name>] [--target <node>]
  migrate_vm --node <node> --vmid <id> --target <node>
  convert_vm_to_template --node <node> --vmid <id>
  update_vm_config --node <node> --vmid <id> --<config-key> <value>
  vm_power --node <node> --vmid <id> --mode <start|stop|shutdown|reboot|reset>
  set_vm_agent --node <node> --vmid <id> [--enabled 1|0]
  create_vm --node <node> --vmid <id> --name <name> [--memory <mb>] [--cores <n>]
  attach_cdrom_iso --node <node> --vmid <id> --iso <storage:iso/file.iso> [--slot ide2]
  set_net_boot_config --node <node> --vmid <id> --net0 <value> --boot <value>
  start_installer_and_console_ticket --node <node> --vmid <id>
  enable_serial_console --node <node> --vmid <id>
  review_install_tasks --node <node> --vmid <id>
  sendkey --node <node> --vmid <id> --key <key>

Phase roadmap:
  Phase 1: read/task actions
  Phase 2: VM lifecycle/config actions
  Phase 3: cloud-init/qga actions
  Phase 4: console/websocket actions
  Phase 5: privilege ladder/root actions
  Phase 6: policy/cleanup actions
`
}

func workflowHelp() string {
	return `proxmox-cli workflow - execute composed workflow

Usage:
  proxmox-cli [global options] workflow <name> [args]

Examples:
  proxmox-cli workflow ubuntu24-serial-autoinstall
  proxmox-cli --dry-run workflow ubuntu24-serial-autoinstall --vmid 120
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
