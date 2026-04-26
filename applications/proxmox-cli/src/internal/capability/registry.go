package capability

import (
	"context"
	"sort"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

type Handler func(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error)

type Meta struct {
	Async          bool
	Capability     string
	WaitSkipReason string
}

type registryEntry struct {
	handler Handler
	meta    Meta
}

type CapabilityGroup struct {
	Name    string
	Actions []string
}

const (
	waitSkipReadOnly         = "action is read-only"
	waitSkipSelfPolled       = "action is synchronous or self-polled"
	waitSkipSessionDriven    = "action is synchronous or session-driven"
	waitSkipSynchronous      = "action is synchronous"
	capabilityInventory      = "inventory"
	capabilityTask           = "task"
	capabilityVM             = "vm"
	capabilityGuest          = "guest"
	capabilityStorage        = "storage"
	capabilityConsole        = "console"
	capabilitySSH            = "ssh"
	capabilityAccess         = "access"
	capabilityISOBuilder     = "iso-builder"
	capabilityInstallReview  = "install-review"
	capabilitySerialFallback = "serial-fallback"
)

var operationRegistry = map[string]registryEntry{
	"list_nodes":                {handler: runListNodes, meta: Meta{Capability: capabilityInventory, WaitSkipReason: waitSkipReadOnly}},
	"list_cluster_resources":    {handler: runListClusterResources, meta: Meta{Capability: capabilityInventory, WaitSkipReason: waitSkipReadOnly}},
	"list_vms_by_node":          {handler: runListVMsByNode, meta: Meta{Capability: capabilityInventory, WaitSkipReason: waitSkipReadOnly}},
	"get_vm_config":             {handler: runGetVMConfig, meta: Meta{Capability: capabilityInventory, WaitSkipReason: waitSkipReadOnly}},
	"get_effective_permissions": {handler: runGetEffectivePermissions, meta: Meta{Capability: capabilityInventory, WaitSkipReason: waitSkipReadOnly}},
	"get_task_status":           {handler: runGetTaskStatus, meta: Meta{Capability: capabilityTask, WaitSkipReason: waitSkipReadOnly}},
	"get_next_vmid":             {handler: runGetNextVMID, meta: Meta{Capability: capabilityTask, WaitSkipReason: waitSkipReadOnly}},
	"get_vm_status":             {handler: runGetVMStatus, meta: Meta{Capability: capabilityInventory, WaitSkipReason: waitSkipReadOnly}},
	"list_tasks_by_vmid":        {handler: runListTasksByVMID, meta: Meta{Capability: capabilityTask, WaitSkipReason: waitSkipReadOnly}},

	"clone_template":         {handler: runCloneTemplate, meta: Meta{Capability: capabilityVM, Async: true}},
	"migrate_vm":             {handler: runMigrateVM, meta: Meta{Capability: capabilityVM, Async: true}},
	"convert_vm_to_template": {handler: runConvertVMToTemplate, meta: Meta{Capability: capabilityVM, Async: true}},
	"update_vm_config":       {handler: runUpdateVMConfig, meta: Meta{Capability: capabilityVM, WaitSkipReason: waitSkipSynchronous}},
	"vm_power":               {handler: runVMPower, meta: Meta{Capability: capabilityVM, Async: true}},
	"set_vm_agent":           {handler: runSetVMAgent, meta: Meta{Capability: capabilityVM, WaitSkipReason: waitSkipSynchronous}},
	"create_vm":              {handler: runCreateVM, meta: Meta{Capability: capabilityVM, Async: true}},
	"attach_cdrom_iso":       {handler: runAttachCDROMISO, meta: Meta{Capability: capabilityVM, WaitSkipReason: waitSkipSynchronous}},
	"set_net_boot_config":    {handler: runSetNetBootConfig, meta: Meta{Capability: capabilityVM, WaitSkipReason: waitSkipSynchronous}},
	"enable_serial_console":  {handler: runEnableSerialConsole, meta: Meta{Capability: capabilityVM, WaitSkipReason: waitSkipSynchronous}},
	"review_install_tasks":   {handler: runReviewInstallTasks, meta: Meta{Capability: capabilityInstallReview, WaitSkipReason: waitSkipSynchronous}},
	"sendkey":                {handler: runSendKey, meta: Meta{Capability: capabilitySerialFallback, WaitSkipReason: waitSkipSynchronous}},

	"agent_network_get_interfaces": {handler: runAgentNetworkGetInterfaces, meta: Meta{Capability: capabilityGuest, WaitSkipReason: waitSkipSelfPolled}},
	"agent_exec":                   {handler: runAgentExec, meta: Meta{Capability: capabilityGuest, WaitSkipReason: waitSkipSelfPolled}},
	"agent_exec_status":            {handler: runAgentExecStatus, meta: Meta{Capability: capabilityGuest, WaitSkipReason: waitSkipSelfPolled}},
	"storage_upload_guard":         {handler: runStorageUploadGuard, meta: Meta{Capability: capabilityStorage, WaitSkipReason: waitSkipSelfPolled}},
	"storage_upload_iso":           {handler: runStorageUploadISO, meta: Meta{Capability: capabilityStorage, WaitSkipReason: waitSkipSelfPolled}},
	"build_ubuntu_autoinstall_iso": {handler: runBuildUbuntuAutoinstallISOAdapter, meta: Meta{Capability: capabilityISOBuilder, WaitSkipReason: waitSkipSelfPolled}},

	"open_vm_termproxy":                 {handler: runOpenVMTermproxy, meta: Meta{Capability: capabilityConsole, WaitSkipReason: waitSkipSessionDriven}},
	"validate_k1_serial_readable":       {handler: runValidateK1SerialReadable, meta: Meta{Capability: capabilityConsole, WaitSkipReason: waitSkipSessionDriven}},
	"serial_ws_session_control":         {handler: runSerialWSSessionControl, meta: Meta{Capability: capabilityConsole, WaitSkipReason: waitSkipSessionDriven}},
	"validate_serial_output_criterion2": {handler: runValidateSerialOutputCriterion2, meta: Meta{Capability: capabilityConsole, WaitSkipReason: waitSkipSessionDriven}},
	"serial_ws_capture_to_file":         {handler: runSerialWSCaptureToFile, meta: Meta{Capability: capabilityConsole, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_check_service":                 {handler: runSSHCheckServiceAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_inject_pubkey_qga":             {handler: runSSHInjectPubKeyQGA, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_exec":                          {handler: runSSHExecAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_scp_transfer":                  {handler: runSSHScpTransferAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_print_connect_command":         {handler: runSSHPrintConnectCommandAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_tunnel_start":                  {handler: runSSHTunnelStartAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_tunnel_status":                 {handler: runSSHTunnelStatusAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},
	"ssh_tunnel_stop":                   {handler: runSSHTunnelStopAdapter, meta: Meta{Capability: capabilitySSH, WaitSkipReason: waitSkipSessionDriven}},

	"create_pve_user_with_root": {handler: runCreatePVEUserWithRoot, meta: Meta{Capability: capabilityAccess, WaitSkipReason: waitSkipSessionDriven}},
	"create_pool_with_root":     {handler: runCreatePoolWithRoot, meta: Meta{Capability: capabilityAccess, WaitSkipReason: waitSkipSessionDriven}},
	"get_user_acl_binding":      {handler: runGetUserACLBinding, meta: Meta{Capability: capabilityAccess, WaitSkipReason: waitSkipSessionDriven}},
	"grant_user_acl":            {handler: runGrantUserACL, meta: Meta{Capability: capabilityAccess, WaitSkipReason: waitSkipSessionDriven}},
	"revoke_user_acl":           {handler: runRevokeUserACL, meta: Meta{Capability: capabilityAccess, WaitSkipReason: waitSkipSessionDriven}},
}

func Dispatch(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, Meta, error) {
	entry, ok := operationRegistry[req.Name]
	if !ok {
		return nil, Meta{}, apperr.New(apperr.CodeInvalidArgs, "operation not implemented yet: "+req.Name)
	}
	result, err := entry.handler(ctx, client, req)
	if err != nil {
		return nil, entry.meta, err
	}
	return result, entry.meta, nil
}

func LookupMeta(name string) (Meta, bool) {
	entry, ok := operationRegistry[name]
	if !ok {
		return Meta{}, false
	}
	return entry.meta, true
}

func CapabilityGroups() []CapabilityGroup {
	order := []string{
		capabilityInventory,
		capabilityTask,
		capabilityVM,
		capabilityGuest,
		capabilityStorage,
		capabilityISOBuilder,
		capabilityConsole,
		capabilitySSH,
		capabilityAccess,
		capabilityInstallReview,
		capabilitySerialFallback,
	}
	indexed := map[string]int{}
	for idx, name := range order {
		indexed[name] = idx
	}
	buckets := map[string][]string{}
	for name, entry := range operationRegistry {
		capability := entry.meta.Capability
		if capability == "" {
			capability = "misc"
		}
		buckets[capability] = append(buckets[capability], name)
	}
	capabilities := make([]string, 0, len(buckets))
	for capability := range buckets {
		capabilities = append(capabilities, capability)
	}
	sort.Slice(capabilities, func(i, j int) bool {
		left, leftOK := indexed[capabilities[i]]
		right, rightOK := indexed[capabilities[j]]
		if leftOK && rightOK {
			return left < right
		}
		if leftOK {
			return true
		}
		if rightOK {
			return false
		}
		return capabilities[i] < capabilities[j]
	})
	groups := make([]CapabilityGroup, 0, len(capabilities))
	for _, capability := range capabilities {
		actions := buckets[capability]
		sort.Strings(actions)
		groups = append(groups, CapabilityGroup{Name: capability, Actions: actions})
	}
	return groups
}

func runBuildUbuntuAutoinstallISOAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = ctx
	_ = client
	return runBuildUbuntuAutoinstallISO(req)
}

func runSSHCheckServiceAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = client
	return runSSHCheckService(ctx, req)
}

func runSSHExecAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = client
	return runSSHExec(ctx, req)
}

func runSSHScpTransferAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = client
	return runSSHScpTransfer(ctx, req)
}

func runSSHPrintConnectCommandAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = ctx
	_ = client
	return runSSHPrintConnectCommand(req)
}

func runSSHTunnelStartAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = client
	return runSSHTunnelStart(ctx, req)
}

func runSSHTunnelStatusAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = ctx
	_ = client
	return runSSHTunnelStatus(req)
}

func runSSHTunnelStopAdapter(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	_ = ctx
	_ = client
	return runSSHTunnelStop(req)
}
