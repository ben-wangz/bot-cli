package capability

var capabilityHelpMetaCore = map[string]HelpMeta{
	"list_nodes": {
		Summary:  "List all cluster nodes.",
		Examples: []string{"proxmox-cli capability list_nodes"},
	},
	"list_cluster_resources": {
		Summary:      "List cluster resources, optionally by type.",
		OptionalArgs: []string{"type"},
		Examples:     []string{"proxmox-cli capability list_cluster_resources --type vm"},
	},
	"list_vms_by_node": {
		Summary:      "List VMs on one node.",
		RequiredArgs: []string{"node"},
		Examples:     []string{"proxmox-cli capability list_vms_by_node --node pve1"},
	},
	"get_vm_config": {
		Summary:      "Get VM config.",
		RequiredArgs: []string{"node", "vmid"},
		Examples:     []string{"proxmox-cli capability get_vm_config --node pve1 --vmid 1200"},
	},
	"get_effective_permissions": {
		Summary:      "Resolve effective permissions at a path.",
		OptionalArgs: []string{"path"},
		Examples:     []string{"proxmox-cli capability get_effective_permissions --path /pool/dev-pool"},
	},
	"get_task_status": {
		Summary:      "Get async task status by UPID.",
		RequiredArgs: []string{"node", "upid"},
		Examples:     []string{"proxmox-cli capability get_task_status --node pve1 --upid UPID:pve1:..."},
	},
	"get_next_vmid": {
		Summary:  "Get next available VMID inside allowed operation range.",
		Examples: []string{"proxmox-cli capability get_next_vmid"},
	},
	"get_vm_status": {
		Summary:      "Get current VM runtime status.",
		RequiredArgs: []string{"node", "vmid"},
		Examples:     []string{"proxmox-cli capability get_vm_status --node pve1 --vmid 1200"},
	},
	"list_tasks_by_vmid": {
		Summary:      "List tasks for a VM on a node.",
		RequiredArgs: []string{"node", "vmid"},
		OptionalArgs: []string{"source"},
		Examples:     []string{"proxmox-cli capability list_tasks_by_vmid --node pve1 --vmid 1200 --source active"},
	},
	"clone_template": {
		Summary:      "Clone template VM into a new VM.",
		RequiredArgs: []string{"node", "source-vmid", "target-vmid"},
		OptionalArgs: []string{"full", "name", "target", "pool"},
		Examples:     []string{"proxmox-cli capability clone_template --node pve1 --source-vmid 1001 --target-vmid 1200 --pool dev-pool"},
	},
	"migrate_vm": {
		Summary:      "Migrate VM to another node.",
		RequiredArgs: []string{"node", "vmid", "target"},
		OptionalArgs: []string{"online", "with-local-disks"},
		Examples:     []string{"proxmox-cli capability migrate_vm --node pve1 --vmid 1200 --target pve2 --online 1"},
	},
	"convert_vm_to_template": {
		Summary:      "Convert VM into a template.",
		RequiredArgs: []string{"node", "vmid"},
		Examples:     []string{"proxmox-cli capability convert_vm_to_template --node pve1 --vmid 1200"},
	},
	"update_vm_config": {
		Summary:      "Update VM config keys.",
		RequiredArgs: []string{"node", "vmid"},
		OptionalArgs: []string{"any qemu config key"},
		Examples:     []string{"proxmox-cli capability update_vm_config --node pve1 --vmid 1200 --memory 4096 --cores 2"},
	},
	"vm_power": {
		Summary:      "Execute VM power operation.",
		RequiredArgs: []string{"node", "vmid", "mode"},
		OptionalArgs: []string{"desired-state"},
		Examples:     []string{"proxmox-cli capability vm_power --node pve1 --vmid 1200 --mode start"},
	},
	"destroy_vm": {
		Summary:      "Destroy a VM safely within allowed range.",
		RequiredArgs: []string{"node", "vmid"},
		OptionalArgs: []string{"if-missing", "purge", "destroy-unreferenced-disks"},
		Examples:     []string{"proxmox-cli capability destroy_vm --node pve1 --vmid 1200 --if-missing ok"},
	},
	"set_vm_agent": {
		Summary:      "Enable or disable QEMU guest agent.",
		RequiredArgs: []string{"node", "vmid"},
		OptionalArgs: []string{"enabled"},
		Examples:     []string{"proxmox-cli capability set_vm_agent --node pve1 --vmid 1200 --enabled 1"},
	},
	"create_vm": {
		Summary:      "Create VM from explicit config args.",
		RequiredArgs: []string{"node", "vmid", "vm config args"},
		OptionalArgs: []string{"if-exists", "pool"},
		Examples:     []string{"proxmox-cli capability create_vm --node pve1 --vmid 1200 --name vm-1200 --memory 4096 --cores 2"},
	},
	"attach_cdrom_iso": {
		Summary:      "Attach ISO to CDROM slot.",
		RequiredArgs: []string{"node", "vmid", "iso"},
		OptionalArgs: []string{"slot", "media"},
		Examples:     []string{"proxmox-cli capability attach_cdrom_iso --node pve1 --vmid 1200 --iso local:iso/u24.iso"},
	},
	"set_net_boot_config": {
		Summary:      "Configure net boot and NIC config.",
		RequiredArgs: []string{"node", "vmid", "net0", "boot"},
		OptionalArgs: []string{"bootdisk"},
		Examples:     []string{"proxmox-cli capability set_net_boot_config --node pve1 --vmid 1200 --net0 virtio,bridge=vmbr0 --boot order=net0"},
	},
	"enable_serial_console": {
		Summary:      "Enable serial console on VM.",
		RequiredArgs: []string{"node", "vmid"},
		Examples:     []string{"proxmox-cli capability enable_serial_console --node pve1 --vmid 1200"},
	},
	"review_install_tasks": {
		Summary:      "Inspect active installer tasks for VM.",
		RequiredArgs: []string{"node", "vmid"},
		Examples:     []string{"proxmox-cli capability review_install_tasks --node pve1 --vmid 1200"},
	},
	"sendkey": {
		Summary:      "Send key sequence to VM console.",
		RequiredArgs: []string{"node", "vmid", "key"},
		OptionalArgs: []string{"skiplock"},
		Examples:     []string{"proxmox-cli capability sendkey --node pve1 --vmid 1200 --key ret"},
	},
	"agent_network_get_interfaces": {
		Summary:      "List guest interfaces via QGA.",
		RequiredArgs: []string{"node", "vmid"},
		Examples:     []string{"proxmox-cli capability agent_network_get_interfaces --node pve1 --vmid 1200"},
	},
	"agent_exec": {
		Summary:      "Execute command via QGA.",
		RequiredArgs: []string{"node", "vmid", "command"},
		OptionalArgs: []string{"shell", "script", "shell-bin", "timeout-seconds", "poll-interval-ms", "no-wait", "input-data"},
		Examples:     []string{"proxmox-cli capability agent_exec --node pve1 --vmid 1200 --command hostname"},
	},
	"agent_exec_status": {
		Summary:      "Fetch QGA exec status by pid.",
		RequiredArgs: []string{"node", "vmid", "pid"},
		Examples:     []string{"proxmox-cli capability agent_exec_status --node pve1 --vmid 1200 --pid 1234"},
	},
	"storage_upload_guard": {
		Summary:      "Check storage content-type compatibility.",
		RequiredArgs: []string{"node", "storage"},
		OptionalArgs: []string{"content-type"},
		Examples:     []string{"proxmox-cli capability storage_upload_guard --node pve1 --storage local --content-type iso"},
	},
	"storage_upload_iso": {
		Summary:      "Upload ISO file to target storage.",
		RequiredArgs: []string{"node", "storage", "source-path"},
		OptionalArgs: []string{"filename", "if-exists"},
		Examples:     []string{"proxmox-cli --timeout 20m capability storage_upload_iso --node pve1 --storage local --source-path ./artifact.iso"},
	},
	"build_ubuntu_autoinstall_iso": {
		Summary:      "Build Ubuntu autoinstall ISO from bundled assets.",
		RequiredArgs: []string{"source-iso", "output-iso"},
		OptionalArgs: []string{"volume-id", "kernel-cmdline", "work-dir"},
		Examples:     []string{"proxmox-cli capability build_ubuntu_autoinstall_iso --source-iso ./ubuntu.iso --output-iso ./build/autoinstall.iso"},
	},
}
