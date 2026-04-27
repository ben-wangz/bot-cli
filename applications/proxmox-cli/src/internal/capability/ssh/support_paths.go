package sshcap

import (
	"fmt"
	"path/filepath"
	"strings"
)

func resolveTunnelPaths(args map[string]string, t target, localPort int, remoteHost string, remotePort int) (string, string) {
	pidFile := strings.TrimSpace(args["pid-file"])
	logFile := strings.TrimSpace(args["log-file"])
	name := fmt.Sprintf("%s-%d-%s-%d", sanitizeName(t.Host), localPort, sanitizeName(remoteHost), remotePort)
	if pidFile == "" {
		pidFile = filepath.Join("build", "ssh-tunnels", name+".pid")
	}
	if logFile == "" {
		logFile = filepath.Join("build", "ssh-tunnels", name+".log")
	}
	return pidFile, logFile
}
