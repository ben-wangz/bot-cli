package storagecap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

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
