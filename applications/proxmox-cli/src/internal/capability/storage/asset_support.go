package storagecap

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

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
