package capability

var capabilityHelpMetaExtended = map[string]HelpMeta{
	"ssh_check_service": {
		Summary:      "Check SSH reachability/auth in batch mode (key-based).",
		RequiredArgs: []string{"host", "user"},
		OptionalArgs: []string{"port", "identity-file", "connect-timeout-seconds", "extra-args"},
		Examples:     []string{"proxmox-cli capability ssh_check_service --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519"},
	},
	"ssh_inject_pubkey_qga": {
		Summary:      "Inject public key via QGA and require existing guest user.",
		RequiredArgs: []string{"node", "vmid", "username", "pub-key-file|pub-key"},
		Examples:     []string{"proxmox-cli capability ssh_inject_pubkey_qga --node pve1 --vmid 1200 --username ubuntu --pub-key-file ~/.ssh/id_ed25519.pub"},
	},
	"ssh_exec": {
		Summary:      "Execute remote SSH command in batch mode (key-based).",
		RequiredArgs: []string{"host", "user", "command"},
		OptionalArgs: []string{"port", "identity-file", "timeout-seconds", "connect-timeout-seconds", "extra-args"},
		Examples:     []string{"proxmox-cli capability ssh_exec --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519 --command 'uname -a'"},
	},
	"ssh_scp_transfer": {
		Summary:      "Transfer files through SCP in batch mode (key-based).",
		RequiredArgs: []string{"direction", "host", "user", "local-path", "remote-path"},
		OptionalArgs: []string{"port", "identity-file", "recursive", "connect-timeout-seconds", "extra-args"},
		Examples:     []string{"proxmox-cli capability ssh_scp_transfer --direction upload --host 10.0.0.10 --user ubuntu --identity-file ~/.ssh/id_ed25519 --local-path ./file --remote-path /tmp/file"},
	},
	"ssh_print_connect_command": {
		Summary:      "Print suggested ssh connect command.",
		RequiredArgs: []string{"host", "user"},
		OptionalArgs: []string{"port", "identity-file", "connect-timeout-seconds", "extra-args"},
		Examples:     []string{"proxmox-cli capability ssh_print_connect_command --host 10.0.0.10 --user ubuntu"},
	},
	"ssh_tunnel_start": {
		Summary:      "Start SSH local tunnel process.",
		RequiredArgs: []string{"host", "user", "local-port", "remote-host", "remote-port"},
		OptionalArgs: []string{"pid-file", "log-file", "port", "identity-file", "connect-timeout-seconds", "extra-args"},
		Examples:     []string{"proxmox-cli capability ssh_tunnel_start --host 10.0.0.10 --user ubuntu --local-port 5900 --remote-host 127.0.0.1 --remote-port 22 --pid-file build/tunnel.pid --log-file build/tunnel.log"},
	},
	"ssh_tunnel_status": {
		Summary:      "Check SSH tunnel process status.",
		RequiredArgs: []string{"pid-file"},
		Examples:     []string{"proxmox-cli capability ssh_tunnel_status --pid-file build/tunnel.pid"},
	},
	"ssh_tunnel_stop": {
		Summary:      "Stop SSH tunnel process.",
		RequiredArgs: []string{"pid-file"},
		Examples:     []string{"proxmox-cli capability ssh_tunnel_stop --pid-file build/tunnel.pid"},
	},
	"open_vm_termproxy": {
		Summary:      "Open VM termproxy endpoint.",
		RequiredArgs: []string{"node", "vmid"},
	},
	"validate_k1_serial_readable": {
		Summary:      "Validate serial readability criterion K1.",
		RequiredArgs: []string{"node", "vmid"},
	},
	"serial_ws_session_control": {
		Summary:      "Control serial websocket session lifecycle.",
		RequiredArgs: []string{"action"},
	},
	"validate_serial_output_criterion2": {
		Summary:      "Validate serial output criterion 2.",
		RequiredArgs: []string{"contains"},
	},
	"serial_ws_capture_to_file": {
		Summary:      "Capture serial websocket output to file.",
		RequiredArgs: []string{"output-file"},
	},
	"create_pve_user_with_root": {
		Summary:      "Create/reuse PVE user using root scope.",
		RequiredArgs: []string{"userid"},
		OptionalArgs: []string{"password", "if-exists", "comment", "email", "firstname", "lastname", "enable", "expire", "keys"},
	},
	"create_pool_with_root": {
		Summary:      "Create/reuse pool using root scope.",
		RequiredArgs: []string{"poolid"},
		OptionalArgs: []string{"if-exists", "comment"},
	},
	"get_user_acl_binding": {
		Summary:      "List ACL bindings for user.",
		RequiredArgs: []string{"userid"},
		OptionalArgs: []string{"path", "role"},
	},
	"grant_user_acl": {
		Summary:      "Grant ACL binding for user.",
		RequiredArgs: []string{"userid", "path", "role"},
		OptionalArgs: []string{"propagate"},
	},
	"revoke_user_acl": {
		Summary:      "Revoke ACL binding for user.",
		RequiredArgs: []string{"userid", "path", "role"},
	},
}
