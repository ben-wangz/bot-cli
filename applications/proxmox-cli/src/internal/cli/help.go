package cli

func rootHelp() string {
	return `proxmox-cli - Agent-facing CLI for Proxmox operations

Usage:
  proxmox-cli [global options] <command> [args]

Commands:
  action      Execute one action (A01-A53 roadmap)
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
  clone_template --node <node> --source-vmid <id> --target-vmid <id> [--name <name>] [--target <node>] [--full 0|1 default=0] [--pool <poolid>]
  migrate_vm --node <node> --vmid <id> --target <node>
  convert_vm_to_template --node <node> --vmid <id>
  update_vm_config --node <node> --vmid <id> --<config-key> <value>
  vm_power --node <node> --vmid <id> --mode <start|stop|shutdown|reboot|reset> [--desired-state running|stopped]
  set_vm_agent --node <node> --vmid <id> [--enabled 1|0]
  create_vm --node <node> --vmid <id> --name <name> [--memory <mb>] [--cores <n>] [--if-exists fail|reuse] [--pool <poolid>]
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
  storage_upload_iso --node <node> --storage <storage> --source-path <file.iso> [--filename <name.iso>] [--if-exists replace|skip] (waits for upload task completion; hint: run storage_upload_guard first)
  build_ubuntu_autoinstall_iso --source-iso <ubuntu.iso> --output-iso <custom.iso> [--kernel-cmdline <cmdline>] [--work-dir build/autoinstall-iso-work/<id>]
  render_and_serve_seed --vmid <id> [--seed-dir build/seed] [--seed-name vm-<id>] [--host 127.0.0.1] [--port 8088]

Phase 4 implemented actions:
  start_vnc_proxy --node <node> --vmid <id> [--websocket 1|0]
  connect_vnc_websocket --node <node> --vmid <id> [--port <port>] [--ticket <ticket>] [--probe-seconds 2]
  open_vm_termproxy --node <node> --vmid <id> [--serial serial0]
  validate_k1_serial_readable --node <node> --vmid <id> [--script <multi-line>] [--expect <text>] [--timeout-seconds 20]
  serial_ws_session_control --node <node> --vmid <id> [--script <multi-line>] [--expect <text>] [--timeout-seconds 60]
  validate_serial_output_criterion2 --node <node> --vmid <id> [--log-path <file>] [--append 1|0 default=1] [--script <multi-line>] [--expect <text>] [--timeout-seconds 120]
  serial_ws_capture_to_file --node <node> --vmid <id> --log-path <file> [--append 1|0 default=1] [--script <multi-line>] [--expect <text>] [--timeout-seconds 120]
  ssh_check_service --host <ip> [--port 22] --user <user> [--identity-file <key>] [--connect-timeout-seconds 5]
  ssh_inject_pubkey_qga --node <node> --vmid <id> --username <user> (--pub-key-file <file> | --pub-key <key>)
  ssh_exec --host <ip> [--port 22] --user <user> [--identity-file <key>] --command <cmd> [--timeout-seconds 30]
  ssh_scp_transfer --direction upload|download --host <ip> [--port 22] --user <user> [--identity-file <key>] --local-path <path> --remote-path <path> [--recursive 1|0]
  ssh_print_connect_command --host <ip> [--port 22] --user <user> [--identity-file <key>] [--extra-args "..."]
  ssh_tunnel_start --host <ip> [--port 22] --user <user> [--identity-file <key>] --local-port <port> --remote-host <host> --remote-port <port> [--pid-file <file>] [--log-file <file>]
  ssh_tunnel_status --pid-file <file>
  ssh_tunnel_stop --pid-file <file>

serial_ws_capture_to_file runbook (required for install diagnosis):
  1) Serial output is not persisted by Proxmox; re-run the whole install flow before capture if previous run is gone.
  2) Always run capture in background with long timeout (for example nohup + pid file), then watch log growth and stop early when enough evidence is collected.
  3) Reconnect sessions must append to the same log file; do not overwrite. Keep --append at default value 1.
  4) If log only shows "OKstarting serial terminal on interface serial0", treat it as a console redirection failure and re-check ISO kernel cmdline has serial console enabled.

Phase 5 implemented actions:
  create_pve_user_with_root --userid <user@realm> [--password <password>] [--comment <text>] [--email <mail>] [--firstname <name>] [--lastname <name>] [--enable 1|0] [--expire <epoch>] [--if-exists fail|reuse]
  create_pool_with_root --poolid <poolid> [--comment <text>] [--if-exists fail|reuse]
  get_user_acl_binding --userid <user@realm> [--path <acl-path>] [--role <role>]
  grant_user_acl --userid <user@realm> --path <acl-path> --role <role> [--propagate 1|0]
  revoke_user_acl --userid <user@realm> --path <acl-path> --role <role>
  node_termproxy_shell_exec --node <node> [--cmd login] [--cmd-opt <csv>] [--cmd-opts <nul-terminated>] [--script <multi-line>] [--expect <text>] [--timeout-seconds 60]

Phase 5 scope note:
  - Recommended direction is root-assisted bootstrap only (create pool/user + get/grant/revoke ACL).
  - ACL changes should use revoke+grant composition; no dedicated update action is required.
  - node_termproxy_shell_exec remains available as legacy compatibility path.

Phase roadmap:
  Phase 1: read/task actions
  Phase 2: VM lifecycle/config actions
  Phase 3: cloud-init/qga actions
  Phase 4: console/websocket + ssh control-plane actions
  Phase 5: privilege ladder/root actions
`
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
