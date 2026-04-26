package action

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

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
	ifExists := strings.TrimSpace(strings.ToLower(req.Args["if-exists"]))
	if ifExists == "" {
		ifExists = "replace"
	}
	if !isOneOf(ifExists, "replace", "skip") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "if-exists must be one of replace|skip")
	}
	expectedVolID := fmt.Sprintf("%s:iso/%s", storage, uploadFilename)
	if ifExists == "skip" {
		exists, existsErr := storageVolumeExists(ctx, client, node, storage, expectedVolID)
		if existsErr != nil {
			return nil, existsErr
		}
		if exists {
			result := map[string]any{
				"node":           node,
				"storage":        storage,
				"filename":       uploadFilename,
				"content":        "iso",
				"source_path":    sourcePath,
				"volid":          expectedVolID,
				"uploaded":       false,
				"skipped_upload": true,
			}
			diagnostics := map[string]any{"if_exists": ifExists, "wait_skipped": "existing iso reused"}
			return buildResult(req, map[string]any{"node": node, "storage": storage, "source-path": sourcePath, "if-exists": ifExists}, result, diagnostics), nil
		}
	}
	data, err := client.PostMultipartFile(ctx, path, fields, "filename", sourcePath, uploadFilename)
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(asString(data))
	volID := expectedVolID
	uploadUPID := ""
	var uploadTaskStatus map[string]any
	if strings.HasPrefix(strings.ToUpper(raw), "UPID:") {
		uploadUPID = raw
	} else if strings.Contains(raw, ":") && strings.Contains(raw, "/") {
		volID = raw
	}
	if uploadUPID != "" {
		waitNode := strings.TrimSpace(parseUPIDNode(uploadUPID))
		if waitNode == "" {
			waitNode = node
		}
		status, waitErr := WaitTask(ctx, client, waitNode, uploadUPID, WaitOptions{Timeout: 20 * time.Minute, Interval: 2 * time.Second})
		if waitErr != nil {
			return nil, apperr.Wrap(apperr.CodeNetwork, "storage upload task did not complete successfully", waitErr)
		}
		uploadTaskStatus = status
	}
	result := map[string]any{
		"node":        node,
		"storage":     storage,
		"filename":    uploadFilename,
		"content":     "iso",
		"source_path": sourcePath,
		"volid":       volID,
		"uploaded":    true,
	}
	if uploadUPID != "" {
		result["upload_upid"] = uploadUPID
		if waitNode := strings.TrimSpace(parseUPIDNode(uploadUPID)); waitNode != "" {
			result["upload_task_node"] = waitNode
		}
		result["upload_task"] = uploadTaskStatus
	}
	if filename != "" {
		result["requested_filename"] = filename
	}
	diagnostics := map[string]any{"if_exists": ifExists}
	if uploadUPID != "" {
		diagnostics["upload_task_upid"] = uploadUPID
		diagnostics["upload_waited"] = true
	}
	return buildResult(req, map[string]any{"node": node, "storage": storage, "source-path": sourcePath, "if-exists": ifExists}, result, diagnostics), nil
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
	userDataPath, metaDataPath, err := prepareNoCloudFiles(treeDir)
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
