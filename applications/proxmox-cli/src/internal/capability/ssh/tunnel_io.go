package sshcap

import (
	"errors"
	"os"
	"strings"
)

func cleanupFiles(pidFile string, cleanup map[string]any) {
	if err := os.Remove(pidFile); err == nil || errors.Is(err, os.ErrNotExist) {
		cleanup["removed_pid_file"] = true
	}
	metaPath := pidFile + ".meta.json"
	if err := os.Remove(metaPath); err == nil || errors.Is(err, os.ErrNotExist) {
		cleanup["removed_meta_file"] = true
	}
}

func readFileTail(path string, max int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(data))
	return tailText(text, max)
}
