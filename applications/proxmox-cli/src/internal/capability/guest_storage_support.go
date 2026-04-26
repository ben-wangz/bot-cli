package capability

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

func storageVolumeExists(ctx context.Context, client *pveapi.Client, node string, storage string, expectedVolID string) (bool, error) {
	path := fmt.Sprintf("/nodes/%s/storage/%s/content", url.PathEscape(node), url.PathEscape(storage))
	query := url.Values{}
	query.Set("content", "iso")
	data, err := client.GetData(ctx, path, query)
	if err != nil {
		return false, err
	}
	list, ok := data.([]any)
	if !ok {
		return false, nil
	}
	for _, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(asString(entry["volid"])) == expectedVolID {
			return true, nil
		}
	}
	return false, nil
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

func parseOptionalBoolArg(args map[string]string, key string) (bool, error) {
	raw := strings.TrimSpace(args[key])
	if raw == "" {
		return false, nil
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, apperr.New(apperr.CodeInvalidArgs, key+" must be one of 1|0|true|false")
	}
}

func writeSeedFile(seedPath string, filename string, content string) error {
	path := filepath.Join(seedPath, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return apperr.Wrap(apperr.CodeConfig, "failed to write seed file "+filename, err)
	}
	return nil
}

func prepareNoCloudFiles(treeDir string) (string, string, error) {
	noCloudDir := filepath.Join(treeDir, "nocloud")
	if err := os.MkdirAll(noCloudDir, 0o755); err != nil {
		return "", "", apperr.Wrap(apperr.CodeConfig, "failed to create nocloud directory", err)
	}
	userDataPath := filepath.Join(noCloudDir, "user-data")
	metaDataPath := filepath.Join(noCloudDir, "meta-data")
	assetUserDataPath, assetMetaDataPath, err := resolveUbuntu2404AutoinstallAssets()
	if err != nil {
		return "", "", err
	}
	if err := copyFile(assetUserDataPath, userDataPath); err != nil {
		return "", "", err
	}
	if err := copyFile(assetMetaDataPath, metaDataPath); err != nil {
		return "", "", err
	}
	return userDataPath, metaDataPath, nil
}

func resolveUbuntu2404AutoinstallAssets() (string, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", "", apperr.Wrap(apperr.CodeConfig, "failed to resolve working directory", err)
	}
	current := workingDir
	for {
		assetDir := filepath.Join(current, "applications", "proxmox-cli", "assets", "ubuntu-24.04")
		userDataPath := filepath.Join(assetDir, "user-data")
		metaDataPath := filepath.Join(assetDir, "meta-data")
		if isRegularFile(userDataPath) && isRegularFile(metaDataPath) {
			return userDataPath, metaDataPath, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", "", apperr.New(apperr.CodeConfig, "required assets not found: applications/proxmox-cli/assets/ubuntu-24.04/{user-data,meta-data}")
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

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
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
