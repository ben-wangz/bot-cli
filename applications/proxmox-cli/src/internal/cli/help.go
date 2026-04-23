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

Safety env vars (for VMID-scoped actions):
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
  clone_template --node <node> --source-vmid <id> --target-vmid <id> [--name <name>] [--target <node>] [--full 0|1 default=0]
  migrate_vm --node <node> --vmid <id> --target <node>
  convert_vm_to_template --node <node> --vmid <id>
  update_vm_config --node <node> --vmid <id> --<config-key> <value>
  vm_power --node <node> --vmid <id> --mode <start|stop|shutdown|reboot|reset> [--desired-state running|stopped]
  set_vm_agent --node <node> --vmid <id> [--enabled 1|0]
  create_vm --node <node> --vmid <id> --name <name> [--memory <mb>] [--cores <n>] [--if-exists fail|reuse]
  attach_cdrom_iso --node <node> --vmid <id> --iso <storage:iso/file.iso> [--slot ide2] [--media cdrom]
  set_net_boot_config --node <node> --vmid <id> --net0 <value> --boot <value>
  start_installer_and_console_ticket --node <node> --vmid <id>
  enable_serial_console --node <node> --vmid <id>
  review_install_tasks --node <node> --vmid <id>
  sendkey --node <node> --vmid <id> --key <key>

Phase 3 implemented actions:
  agent_network_get_interfaces --node <node> --vmid <id>
  agent_exec --node <node> --vmid <id> --command <command> [--shell 1|0] [--shell-bin /bin/sh] [--script <shell script>] [--input-data <stdin>] [--no-wait 1|0] [--timeout-seconds 30]
  agent_exec_status --node <node> --vmid <id> --pid <pid>
  dump_cloudinit --node <node> --vmid <id> [--type user|network|meta]
  storage_upload_guard --node <node> --storage <storage> [--content-type snippets]
  storage_upload_snippet --node <node> --storage <storage> --source-path <file> [--filename <name>]
  storage_upload_iso --node <node> --storage <storage> --source-path <file.iso> [--filename <name.iso>] [--if-exists replace|skip]
  build_ubuntu_autoinstall_iso --source-iso <ubuntu.iso> --output-iso <custom.iso> [--kernel-cmdline <cmdline>] [--username cloud] [--password <plain>] [--password-hash <hash>] [--hostname <name>] [--work-dir build/autoinstall-iso-work/<id>]
  render_and_serve_seed --vmid <id> [--seed-dir build/seed] [--seed-name vm-<id>] [--host 127.0.0.1] [--port 8088]

Phase 4 implemented actions:
  start_vnc_proxy --node <node> --vmid <id> [--websocket 1|0]
  connect_vnc_websocket --node <node> --vmid <id> [--port <port>] [--ticket <ticket>] [--probe-seconds 2]
  open_vm_termproxy --node <node> --vmid <id> [--serial serial0]
  validate_k1_serial_readable --node <node> --vmid <id> [--script <multi-line>] [--expect <text>] [--timeout-seconds 20]
  serial_ws_session_control --node <node> --vmid <id> [--script <multi-line>] [--expect <text>] [--timeout-seconds 60]
  validate_serial_output_criterion2 --node <node> --vmid <id> [--log-path <file>] [--append 1|0 default=1] [--script <multi-line>] [--expect <text>] [--timeout-seconds 120]
  serial_ws_capture_to_file --node <node> --vmid <id> --log-path <file> [--append 1|0 default=1] [--script <multi-line>] [--expect <text>] [--timeout-seconds 120]

serial_ws_capture_to_file runbook (required for install diagnosis):
  1) Serial output is not persisted by Proxmox; re-run the whole install flow before capture if previous run is gone.
  2) Always run capture in background with long timeout (for example nohup + pid file), then watch log growth and stop early when enough evidence is collected.
  3) Reconnect sessions must append to the same log file; do not overwrite. Keep --append at default value 1.
  4) If log only shows "OKstarting serial terminal on interface serial0", treat it as a console redirection failure and re-check ISO kernel cmdline has serial console enabled.

Phase 5 implemented actions:
  node_termproxy_shell_exec --node <node> [--cmd login] [--cmd-opt <csv>] [--cmd-opts <nul-terminated>] [--script <multi-line>] [--expect <text>] [--timeout-seconds 60]

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
  proxmox-cli workflow ubuntu24-with-agent-template --node eva002 --target-vmid 1201

Implemented workflows:
  ubuntu24-with-agent-template --node <node> --target-vmid <id>

ubuntu24-with-agent-template result:
  - reuse prebuilt installer ISO when available, otherwise build autoinstall ISO
  - create fresh VM from scratch, run unattended install, verify qga readiness
  - convert VM to template and write template VMID to build/ubuntu-24-with-agent.vm-template.id
  - prints each underlying action call to stderr for troubleshooting (sensitive args redacted)
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
