package storagecap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func RunBuildUbuntuAutoinstallISO(req Request) (map[string]any, error) {
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
