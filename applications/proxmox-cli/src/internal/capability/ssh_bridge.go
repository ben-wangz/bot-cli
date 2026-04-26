package capability

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	sshcap "github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/capability/ssh"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runSSHCheckService(ctx context.Context, req Request) (map[string]any, error) {
	return sshcap.CheckService(ctx, sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHExec(ctx context.Context, req Request) (map[string]any, error) {
	return sshcap.Exec(ctx, sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHScpTransfer(ctx context.Context, req Request) (map[string]any, error) {
	return sshcap.ScpTransfer(ctx, sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHPrintConnectCommand(req Request) (map[string]any, error) {
	return sshcap.PrintConnectCommand(sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHTunnelStart(ctx context.Context, req Request) (map[string]any, error) {
	return sshcap.TunnelStart(ctx, sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHTunnelStatus(req Request) (map[string]any, error) {
	return sshcap.TunnelStatus(sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHTunnelStop(req Request) (map[string]any, error) {
	return sshcap.TunnelStop(sshcap.Request{Name: req.Name, Args: req.Args, Scope: req.Scope})
}

func runSSHInjectPubKeyQGA(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	username, err := RequiredString(req.Args, "username")
	if err != nil {
		return nil, err
	}
	pubKey, source, err := sshcap.ResolvePublicKey(req.Args)
	if err != nil {
		return nil, err
	}
	fingerprint, err := sshcap.PublicKeyFingerprint(pubKey)
	if err != nil {
		return nil, err
	}
	script := sshcap.BuildInjectPubKeyScript(username, pubKey)
	agentReq := Request{
		Name: "agent_exec",
		Args: map[string]string{
			"node":            node,
			"vmid":            strconv.Itoa(vmid),
			"command":         "true",
			"shell":           "1",
			"script":          script,
			"timeout-seconds": "45",
		},
		Scope: req.Scope,
	}
	agentResult, err := runAgentExec(ctx, client, agentReq)
	if err != nil {
		return nil, err
	}
	exitCode, exited := sshcap.ExtractAgentExecExit(agentResult)
	if !exited || exitCode != 0 {
		return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("qga key injection command failed with exit code %d", exitCode))
	}
	home := "/home/" + username
	if username == "root" {
		home = "/root"
	}
	request := map[string]any{"node": node, "vmid": vmid, "username": username, "key_source": source}
	result := map[string]any{
		"username":             username,
		"authorized_keys_path": filepath.ToSlash(filepath.Join(home, ".ssh", "authorized_keys")),
		"fingerprint":          fingerprint,
		"qga_exit_code":        exitCode,
	}
	return buildResult(req, request, result, map[string]any{"pubkey_source": source}), nil
}
