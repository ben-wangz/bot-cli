package action

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func ExecutePhase3(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	switch req.Name {
	case "agent_network_get_interfaces":
		return runAgentNetworkGetInterfaces(ctx, client, req)
	case "agent_exec":
		return runAgentExec(ctx, client, req)
	case "agent_exec_status":
		return runAgentExecStatus(ctx, client, req)
	case "dump_cloudinit":
		return runDumpCloudinit(ctx, client, req)
	case "storage_upload_guard":
		return runStorageUploadGuard(ctx, client, req)
	case "storage_upload_snippet":
		return runStorageUploadSnippet(ctx, client, req)
	case "storage_upload_iso":
		return runStorageUploadISO(ctx, client, req)
	case "build_ubuntu_autoinstall_iso":
		return runBuildUbuntuAutoinstallISO(req)
	case "render_and_serve_seed":
		return runRenderAndServeSeed(req)
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported action in phase 3: "+req.Name)
	}
}

func runAgentNetworkGetInterfaces(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, qgaUnavailableError("agent_network_get_interfaces", err)
	}
	payload := unwrapResultField(data)
	ipv4 := collectIPv4(payload)
	return buildResult(req, map[string]any{"node": node, "vmid": vmid}, payload, map[string]any{"ipv4_addresses": ipv4, "ipv4_count": len(ipv4)}), nil
}

func runAgentExec(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	command := strings.TrimSpace(req.Args["command"])
	if command == "" {
		command = strings.TrimSpace(req.Args["cmd"])
	}
	if command == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "missing required action arg --command")
	}
	timeoutSeconds := 30
	if req.Args["timeout-seconds"] != "" {
		parsedTimeout, parseErr := RequiredInt(req.Args, "timeout-seconds")
		if parseErr != nil {
			return nil, parseErr
		}
		timeoutSeconds = parsedTimeout
	}
	pollMillis := 1000
	if req.Args["poll-interval-ms"] != "" {
		parsedPoll, parseErr := RequiredInt(req.Args, "poll-interval-ms")
		if parseErr != nil {
			return nil, parseErr
		}
		pollMillis = parsedPoll
	}
	form := mapArgsToForm(req.Args, "node", "vmid", "command", "cmd", "timeout-seconds", "poll-interval-ms")
	form.Set("command", command)
	path := fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec", url.PathEscape(node), vmid)
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, qgaUnavailableError("agent_exec", err)
	}
	pid := extractExecPID(data)
	if pid == "" {
		return nil, apperr.New(apperr.CodeNetwork, "agent_exec response does not contain pid")
	}
	status, polls, err := pollAgentExecStatus(ctx, client, node, vmid, pid, time.Duration(timeoutSeconds)*time.Second, time.Duration(pollMillis)*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid, "command": command}, map[string]any{"pid": pid, "status": status}, map[string]any{"poll_count": polls, "timeout_seconds": timeoutSeconds}), nil
}

func runAgentExecStatus(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	pid, err := RequiredString(req.Args, "pid")
	if err != nil {
		return nil, err
	}
	status, err := getAgentExecStatus(ctx, client, node, vmid, pid)
	if err != nil {
		return nil, err
	}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid, "pid": pid}, status, map[string]any{"exited": toBool(status["exited"]), "exitcode": status["exitcode"]}), nil
}

func runDumpCloudinit(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	typeName := strings.TrimSpace(req.Args["type"])
	if typeName == "" {
		typeName = "user"
	}
	if !isOneOf(typeName, "user", "network", "meta") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "type must be one of user|network|meta")
	}
	query := url.Values{}
	query.Set("type", typeName)
	path := fmt.Sprintf("/nodes/%s/qemu/%d/cloudinit/dump", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, query)
	if err != nil {
		return nil, err
	}
	content := asString(data)
	result := map[string]any{"type": typeName, "content": content}
	return buildResult(req, map[string]any{"node": node, "vmid": vmid, "type": typeName}, result, map[string]any{"content_length": len(content)}), nil
}

func runStorageUploadGuard(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	storage, err := RequiredString(req.Args, "storage")
	if err != nil {
		return nil, err
	}
	contentType := strings.TrimSpace(req.Args["content-type"])
	if contentType == "" {
		contentType = "snippets"
	}
	config, err := getStorageConfig(ctx, client, node, storage)
	if err != nil {
		return nil, err
	}
	supported := splitCSV(asString(config["content"]))
	allowed := containsString(supported, contentType)
	hint := ""
	if !allowed && contentType == "snippets" {
		hint = "target storage does not allow snippets; enable snippets content type or choose another storage"
	}
	result := map[string]any{
		"node":                    node,
		"storage":                 storage,
		"content_type":            contentType,
		"upload_allowed":          allowed,
		"supported_content_types": supported,
		"hint":                    hint,
	}
	diagnostics := map[string]any{"guard": "allowed"}
	if !allowed {
		diagnostics["guard"] = "blocked"
	}
	return buildResult(req, map[string]any{"node": node, "storage": storage, "content-type": contentType}, result, diagnostics), nil
}

func runStorageUploadSnippet(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	storage, err := RequiredString(req.Args, "storage")
	if err != nil {
		return nil, err
	}
	sourcePath, err := RequiredString(req.Args, "source-path")
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(sourcePath); statErr != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "source-path is not readable", statErr)
	}
	filename := strings.TrimSpace(req.Args["filename"])
	if filename == "" {
		filename = filepath.Base(sourcePath)
	}
	config, err := getStorageConfig(ctx, client, node, storage)
	if err != nil {
		return nil, err
	}
	supported := splitCSV(asString(config["content"]))
	if !containsString(supported, "snippets") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "target storage does not allow snippets; run storage_upload_guard first")
	}
	path := fmt.Sprintf("/nodes/%s/storage/%s/upload", url.PathEscape(node), url.PathEscape(storage))
	fields := map[string]string{
		"content": "snippets",
	}
	data, err := client.PostMultipartFile(ctx, path, fields, "filename", sourcePath, filename)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"node":       node,
		"storage":    storage,
		"filename":   filename,
		"volid":      asString(data),
		"content":    "snippets",
		"sourcePath": sourcePath,
	}
	return buildResult(req, map[string]any{"node": node, "storage": storage, "source-path": sourcePath, "filename": filename}, result, map[string]any{}), nil
}

func runStorageUploadISO(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	storage, err := RequiredString(req.Args, "storage")
	if err != nil {
		return nil, err
	}
	sourcePath, err := RequiredString(req.Args, "source-path")
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(sourcePath); statErr != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "source-path is not readable", statErr)
	}
	config, err := getStorageConfig(ctx, client, node, storage)
	if err != nil {
		return nil, err
	}
	supported := splitCSV(asString(config["content"]))
	if !containsString(supported, "iso") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "target storage does not allow iso uploads")
	}
	path := fmt.Sprintf("/nodes/%s/storage/%s/upload", url.PathEscape(node), url.PathEscape(storage))
	fields := map[string]string{"content": "iso"}
	filename := strings.TrimSpace(req.Args["filename"])
	uploadFilename := filepath.Base(sourcePath)
	if filename != "" {
		uploadFilename = filename
	}
	data, err := client.PostMultipartFile(ctx, path, fields, "filename", sourcePath, uploadFilename)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"node":        node,
		"storage":     storage,
		"filename":    uploadFilename,
		"content":     "iso",
		"source_path": sourcePath,
		"volid":       asString(data),
	}
	if filename != "" {
		result["requested_filename"] = filename
	}
	return buildResult(req, map[string]any{"node": node, "storage": storage, "source-path": sourcePath}, result, map[string]any{}), nil
}

func runBuildUbuntuAutoinstallISO(req Request) (map[string]any, error) {
	sourceISO, err := RequiredString(req.Args, "source-iso")
	if err != nil {
		return nil, err
	}
	outputISO, err := RequiredString(req.Args, "output-iso")
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(sourceISO); statErr != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "source-iso is not readable", statErr)
	}
	if os.Geteuid() != 0 {
		return nil, apperr.New(apperr.CodeAuth, "build_ubuntu_autoinstall_iso requires root privileges for loop mount")
	}
	volumeID := strings.TrimSpace(req.Args["volume-id"])
	if volumeID == "" {
		volumeID = "Ubuntu-24.04-Autoinstall"
	}
	kernelCmdline := strings.TrimSpace(req.Args["kernel-cmdline"])
	if kernelCmdline == "" {
		kernelCmdline = "autoinstall ds=nocloud\\;s=/cdrom/nocloud/ console=tty1 console=ttyS0,115200n8"
	}
	workDir := strings.TrimSpace(req.Args["work-dir"])
	if workDir == "" {
		workDir = filepath.Join("build", "autoinstall-iso-work", fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	absoluteWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to resolve work-dir", err)
	}
	mountDir := filepath.Join(absoluteWorkDir, "mount")
	treeDir := filepath.Join(absoluteWorkDir, "tree")
	if err := os.MkdirAll(mountDir, 0o755); err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create mount directory", err)
	}
	if err := os.MkdirAll(treeDir, 0o755); err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create tree directory", err)
	}
	if _, err := runLocalCommand(context.Background(), "mount", "-o", "loop,ro", sourceISO, mountDir); err != nil {
		return nil, err
	}
	mounted := true
	defer func() {
		if mounted {
			_, _ = runLocalCommand(context.Background(), "umount", mountDir)
		}
	}()
	if _, err := runLocalCommand(context.Background(), "cp", "-a", mountDir+"/.", treeDir+"/"); err != nil {
		return nil, err
	}
	if _, err := runLocalCommand(context.Background(), "umount", mountDir); err != nil {
		return nil, err
	}
	mounted = false
	userDataPath, metaDataPath, err := prepareNoCloudFiles(req.Args, treeDir)
	if err != nil {
		return nil, err
	}
	modified, err := patchUbuntuBootConfigs(treeDir, kernelCmdline)
	if err != nil {
		return nil, err
	}
	if len(modified) == 0 {
		return nil, apperr.New(apperr.CodeConfig, "did not find boot config files to inject kernel cmdline")
	}
	absoluteOutputISO, err := filepath.Abs(outputISO)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to resolve output-iso", err)
	}
	if err := os.MkdirAll(filepath.Dir(absoluteOutputISO), 0o755); err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create output directory", err)
	}
	mkisofsArgs, err := buildUbuntuISO(treeDir, absoluteOutputISO, volumeID)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"source_iso":            sourceISO,
		"output_iso":            absoluteOutputISO,
		"work_dir":              absoluteWorkDir,
		"kernel_cmdline":        kernelCmdline,
		"modified_boot_configs": modified,
		"user_data_path":        userDataPath,
		"meta_data_path":        metaDataPath,
		"mkisofs_args":          mkisofsArgs,
	}
	return buildResult(req, map[string]any{"source-iso": sourceISO, "output-iso": absoluteOutputISO}, result, map[string]any{"modified_file_count": len(modified)}), nil
}

func runRenderAndServeSeed(req Request) (map[string]any, error) {
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	seedDir := strings.TrimSpace(req.Args["seed-dir"])
	if seedDir == "" {
		seedDir = "build/seed"
	}
	seedName := strings.TrimSpace(req.Args["seed-name"])
	if seedName == "" {
		seedName = fmt.Sprintf("vm-%d", vmid)
	}
	hostname := strings.TrimSpace(req.Args["hostname"])
	if hostname == "" {
		hostname = seedName
	}
	host := strings.TrimSpace(req.Args["host"])
	if host == "" {
		host = "127.0.0.1"
	}
	port := 8088
	if strings.TrimSpace(req.Args["port"]) != "" {
		parsedPort, parseErr := RequiredInt(req.Args, "port")
		if parseErr != nil {
			return nil, parseErr
		}
		if parsedPort > 65535 {
			return nil, apperr.New(apperr.CodeInvalidArgs, "port must be <= 65535")
		}
		port = parsedPort
	}
	absSeedDir, err := filepath.Abs(seedDir)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to resolve seed-dir", err)
	}
	seedPath := filepath.Join(absSeedDir, seedName)
	if err := os.MkdirAll(seedPath, 0o755); err != nil {
		return nil, apperr.Wrap(apperr.CodeConfig, "failed to create seed directory", err)
	}
	metaData := strings.TrimSpace(req.Args["meta-data"])
	if metaData == "" {
		metaData = fmt.Sprintf("instance-id: iid-%s\nlocal-hostname: %s\n", seedName, hostname)
	}
	userData := strings.TrimSpace(req.Args["user-data"])
	if userData == "" {
		userData = fmt.Sprintf("#cloud-config\nhostname: %s\nmanage_etc_hosts: true\n", hostname)
	}
	networkConfig := strings.TrimSpace(req.Args["network-config"])
	if networkConfig == "" {
		networkConfig = "version: 2\nethernets:\n  ens18:\n    dhcp4: true\n"
	}
	if err := writeSeedFile(seedPath, "meta-data", metaData); err != nil {
		return nil, err
	}
	if err := writeSeedFile(seedPath, "user-data", userData); err != nil {
		return nil, err
	}
	if err := writeSeedFile(seedPath, "network-config", networkConfig); err != nil {
		return nil, err
	}
	seedURL := fmt.Sprintf("http://%s:%d/%s/", host, port, url.PathEscape(seedName))
	serveCommand := fmt.Sprintf("python3 -m http.server %d --bind %s --directory %s", port, host, absSeedDir)
	result := map[string]any{
		"seed_path":     seedPath,
		"seed_url":      seedURL,
		"serve_address": fmt.Sprintf("%s:%d", host, port),
		"files": map[string]any{
			"meta-data":      filepath.Join(seedPath, "meta-data"),
			"user-data":      filepath.Join(seedPath, "user-data"),
			"network-config": filepath.Join(seedPath, "network-config"),
		},
		"serve_command": serveCommand,
	}
	diagnostics := map[string]any{"served": false, "note": "seed files rendered; start local static server with serve_command"}
	return buildResult(req, map[string]any{"vmid": vmid, "seed-dir": seedDir, "seed-name": seedName}, result, diagnostics), nil
}

func getAgentExecStatus(ctx context.Context, client *pveapi.Client, node string, vmid int, pid string) (map[string]any, error) {
	query := url.Values{}
	query.Set("pid", pid)
	path := fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec-status", url.PathEscape(node), vmid)
	data, err := client.GetData(ctx, path, query)
	if err != nil {
		return nil, qgaUnavailableError("agent_exec_status", err)
	}
	status, ok := unwrapResultField(data).(map[string]any)
	if !ok {
		return nil, apperr.New(apperr.CodeNetwork, "agent_exec_status response is not an object")
	}
	return status, nil
}

func pollAgentExecStatus(ctx context.Context, client *pveapi.Client, node string, vmid int, pid string, timeout time.Duration, interval time.Duration) (map[string]any, int, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if interval <= 0 {
		interval = time.Second
	}
	deadline := time.Now().Add(timeout)
	polls := 0
	for {
		if time.Now().After(deadline) {
			return nil, polls, apperr.New(apperr.CodeNetwork, "agent_exec timed out waiting for exit status")
		}
		status, err := getAgentExecStatus(ctx, client, node, vmid, pid)
		if err != nil {
			return nil, polls, err
		}
		polls++
		if toBool(status["exited"]) {
			return status, polls, nil
		}
		select {
		case <-ctx.Done():
			return nil, polls, apperr.Wrap(apperr.CodeNetwork, "agent_exec polling interrupted", ctx.Err())
		case <-time.After(interval):
		}
	}
}

func extractExecPID(data any) string {
	m, ok := unwrapResultField(data).(map[string]any)
	if !ok {
		return ""
	}
	if v := strings.TrimSpace(asString(m["pid"])); v != "" {
		return v
	}
	if v := strings.TrimSpace(asString(m["PID"])); v != "" {
		return v
	}
	return ""
}

func unwrapResultField(data any) any {
	m, ok := data.(map[string]any)
	if !ok {
		return data
	}
	v, ok := m["result"]
	if !ok {
		return data
	}
	return v
}

func collectIPv4(data any) []string {
	interfaces, ok := data.([]any)
	if !ok {
		return []string{}
	}
	addresses := []string{}
	seen := map[string]bool{}
	for _, item := range interfaces {
		iface, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entries, ok := iface["ip-addresses"].([]any)
		if !ok {
			continue
		}
		for _, addrRaw := range entries {
			addr, ok := addrRaw.(map[string]any)
			if !ok {
				continue
			}
			if strings.ToLower(asString(addr["ip-address-type"])) != "ipv4" {
				continue
			}
			value := strings.TrimSpace(asString(addr["ip-address"]))
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			addresses = append(addresses, value)
		}
	}
	return addresses
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func containsString(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

func getStorageConfig(ctx context.Context, client *pveapi.Client, node string, storage string) (map[string]any, error) {
	path := fmt.Sprintf("/nodes/%s/storage/%s", url.PathEscape(node), url.PathEscape(storage))
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	config := firstObject(unwrapResultField(data))
	if config != nil && strings.TrimSpace(asString(config["content"])) != "" {
		return config, nil
	}
	statusPath := fmt.Sprintf("/nodes/%s/storage/%s/status", url.PathEscape(node), url.PathEscape(storage))
	statusData, statusErr := client.GetData(ctx, statusPath, url.Values{})
	if statusErr == nil {
		config = firstObject(unwrapResultField(statusData))
		if config != nil && strings.TrimSpace(asString(config["content"])) != "" {
			return config, nil
		}
	}
	listPath := fmt.Sprintf("/nodes/%s/storage", url.PathEscape(node))
	listData, listErr := client.GetData(ctx, listPath, url.Values{})
	if listErr == nil {
		if list, ok := listData.([]any); ok {
			for _, entry := range list {
				row, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if strings.TrimSpace(asString(row["storage"])) == storage {
					return row, nil
				}
			}
		}
	}
	if config == nil {
		return nil, apperr.New(apperr.CodeNetwork, "storage response is not an object")
	}
	return config, nil
}

func firstObject(data any) map[string]any {
	if m, ok := data.(map[string]any); ok {
		return m
	}
	list, ok := data.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, _ := list[0].(map[string]any)
	return m
}

func qgaUnavailableError(actionName string, cause error) error {
	message := fmt.Sprintf("%s failed; qemu guest agent may be unavailable (not installed, not running, or VM not started)", actionName)
	return apperr.Wrap(apperr.CodeNetwork, message, cause)
}

func toBool(v any) bool {
	switch value := v.(type) {
	case bool:
		return value
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(value))
		return trimmed == "1" || trimmed == "true" || trimmed == "yes"
	case int:
		return value != 0
	case int64:
		return value != 0
	case float64:
		return value != 0
	default:
		return false
	}
}

func writeSeedFile(seedPath string, filename string, content string) error {
	path := filepath.Join(seedPath, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to write seed file "+filename, err)
	}
	return nil
}

func prepareNoCloudFiles(args map[string]string, treeDir string) (string, string, error) {
	noCloudDir := filepath.Join(treeDir, "nocloud")
	if err := os.MkdirAll(noCloudDir, 0o755); err != nil {
		return "", "", apperr.Wrap(apperr.CodeConfig, "failed to create nocloud directory", err)
	}
	userDataPath := filepath.Join(noCloudDir, "user-data")
	metaDataPath := filepath.Join(noCloudDir, "meta-data")
	providedUserData := strings.TrimSpace(args["user-data-path"])
	providedMetaData := strings.TrimSpace(args["meta-data-path"])
	if providedUserData != "" {
		if err := copyFile(providedUserData, userDataPath); err != nil {
			return "", "", err
		}
	} else {
		userData, err := defaultUbuntuAutoinstallUserData(args)
		if err != nil {
			return "", "", err
		}
		if err := os.WriteFile(userDataPath, []byte(userData), 0o644); err != nil {
			return "", "", apperr.Wrap(apperr.CodeConfig, "failed to write default user-data", err)
		}
	}
	if providedMetaData != "" {
		if err := copyFile(providedMetaData, metaDataPath); err != nil {
			return "", "", err
		}
	} else {
		hostname := strings.TrimSpace(args["hostname"])
		if hostname == "" {
			hostname = "ubuntu2404-auto"
		}
		metaData := fmt.Sprintf("instance-id: iid-%d\nlocal-hostname: %s\n", time.Now().UnixNano(), hostname)
		if err := os.WriteFile(metaDataPath, []byte(metaData), 0o644); err != nil {
			return "", "", apperr.Wrap(apperr.CodeConfig, "failed to write default meta-data", err)
		}
	}
	return userDataPath, metaDataPath, nil
}

func defaultUbuntuAutoinstallUserData(args map[string]string) (string, error) {
	hostname := strings.TrimSpace(args["hostname"])
	if hostname == "" {
		hostname = "ubuntu2404-auto"
	}
	username := strings.TrimSpace(args["username"])
	if username == "" {
		username = "cloud"
	}
	passwordHash := strings.TrimSpace(args["password-hash"])
	if passwordHash == "" {
		password := strings.TrimSpace(args["password"])
		if password == "" {
			return "", apperr.New(apperr.CodeInvalidArgs, "missing --password or --password-hash for default autoinstall user-data")
		}
		hashed, err := hashPasswordSHA512(password)
		if err != nil {
			return "", err
		}
		passwordHash = hashed
	}
	lateCmdConsole := "sed -i 's/^GRUB_CMDLINE_LINUX=.*/GRUB_CMDLINE_LINUX=\"console=tty1 console=ttyS0,115200n8\"/' /etc/default/grub"
	lateCmdDefault := "sed -i 's/^GRUB_CMDLINE_LINUX_DEFAULT=.*/GRUB_CMDLINE_LINUX_DEFAULT=\"console=tty1 console=ttyS0,115200n8\"/' /etc/default/grub"
	return strings.TrimSpace(fmt.Sprintf(`#cloud-config
autoinstall:
  version: 1
  identity:
    hostname: %s
    username: %s
    password: "%s"
  ssh:
    install-server: true
  packages:
    - qemu-guest-agent
  late-commands:
    - curtin in-target --target=/target -- sh -lc '%s'
    - curtin in-target --target=/target -- sh -lc '%s'
    - curtin in-target --target=/target -- systemctl enable serial-getty@ttyS0.service
    - curtin in-target --target=/target -- systemctl enable qemu-guest-agent.service
    - curtin in-target --target=/target -- update-grub
`, hostname, username, passwordHash, lateCmdConsole, lateCmdDefault)), nil
}

func hashPasswordSHA512(password string) (string, error) {
	output, err := runLocalCommand(context.Background(), "python3", "-c", "import crypt,sys; print(crypt.crypt(sys.argv[1], crypt.mksalt(crypt.METHOD_SHA512)))", password)
	if err != nil {
		return "", err
	}
	hash := strings.TrimSpace(output)
	if hash == "" {
		return "", apperr.New(apperr.CodeInternal, "generated password hash is empty")
	}
	return hash, nil
}

func copyFile(sourcePath string, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to open source file", err)
	}
	defer source.Close()
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to create destination directory", err)
	}
	destination, err := os.Create(destPath)
	if err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to create destination file", err)
	}
	defer destination.Close()
	if _, err := io.Copy(destination, source); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to copy file", err)
	}
	return nil
}

func patchUbuntuBootConfigs(treeDir string, kernelCmdline string) ([]string, error) {
	candidates := []string{
		filepath.Join(treeDir, "boot", "grub", "grub.cfg"),
		filepath.Join(treeDir, "boot", "grub", "loopback.cfg"),
		filepath.Join(treeDir, "isolinux", "txt.cfg"),
	}
	modified := []string{}
	for _, filePath := range candidates {
		if _, err := os.Stat(filePath); err != nil {
			continue
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeConfig, "failed to read boot config", err)
		}
		updated, changed := injectKernelCmdline(string(data), kernelCmdline)
		if !changed {
			continue
		}
		if err := os.WriteFile(filePath, []byte(updated), 0o644); err != nil {
			return nil, apperr.Wrap(apperr.CodeConfig, "failed to write boot config", err)
		}
		modified = append(modified, filePath)
	}
	return modified, nil
}

func injectKernelCmdline(content string, kernelCmdline string) (string, bool) {
	if strings.TrimSpace(kernelCmdline) == "" {
		return content, false
	}
	lines := strings.Split(content, "\n")
	changed := false
	matcher := regexp.MustCompile(`(?i)(\s+linux\s+|\s+append\s+)`)
	for i, line := range lines {
		if !strings.Contains(line, "/casper/vmlinuz") {
			continue
		}
		if !matcher.MatchString(line) {
			continue
		}
		if strings.Contains(line, kernelCmdline) {
			continue
		}
		idx := strings.Index(line, " ---")
		if idx < 0 {
			idx = strings.Index(line, "---")
		}
		if idx < 0 {
			continue
		}
		left := strings.TrimRight(line[:idx], " \t")
		right := line[idx:]
		lines[i] = left + " " + kernelCmdline + " " + strings.TrimLeft(right, " ")
		changed = true
	}
	if !changed {
		return content, false
	}
	return strings.Join(lines, "\n"), true
}

func buildUbuntuISO(treeDir string, outputISO string, volumeID string) ([]string, error) {
	biosBoot := filepath.Join(treeDir, "boot", "grub", "i386-pc", "eltorito.img")
	efiBoot := filepath.Join(treeDir, "boot", "grub", "efi.img")
	args := []string{"-r", "-V", volumeID, "-o", outputISO, "-J", "-joliet-long", "-l", "-cache-inodes"}
	if _, err := os.Stat(biosBoot); err == nil {
		args = append(args, "-b", "boot/grub/i386-pc/eltorito.img", "-c", "boot.catalog", "-no-emul-boot", "-boot-load-size", "4", "-boot-info-table")
	}
	if _, err := os.Stat(efiBoot); err == nil {
		args = append(args, "-eltorito-alt-boot", "-e", "boot/grub/efi.img", "-no-emul-boot")
	}
	args = append(args, treeDir)
	if _, err := runLocalCommand(context.Background(), "mkisofs", args...); err != nil {
		return nil, err
	}
	return args, nil
}

func runLocalCommand(ctx context.Context, name string, args ...string) (string, error) {
	commandPath, err := exec.LookPath(name)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeConfig, "required command not found: "+name, err)
	}
	command := exec.CommandContext(ctx, commandPath, args...)
	output, err := command.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		message := fmt.Sprintf("command failed: %s %s", name, strings.Join(args, " "))
		if text != "" {
			message += "; output=" + text
		}
		return text, apperr.Wrap(apperr.CodeNetwork, message, err)
	}
	return text, nil
}
