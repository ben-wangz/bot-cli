package capability

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

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
	originalCommand := command
	useShell, err := parseOptionalBoolArg(req.Args, "shell")
	if err != nil {
		return nil, err
	}
	noWait, err := parseOptionalBoolArg(req.Args, "no-wait")
	if err != nil {
		return nil, err
	}
	inputData := strings.TrimSpace(req.Args["input-data"])
	if useShell {
		shellBin := strings.TrimSpace(req.Args["shell-bin"])
		if shellBin == "" {
			shellBin = "/bin/sh"
		}
		script := strings.TrimSpace(req.Args["script"])
		if script == "" {
			script = originalCommand
		}
		if script == "" {
			return nil, apperr.New(apperr.CodeInvalidArgs, "--shell requires --command or --script")
		}
		if !strings.HasSuffix(script, "\n") {
			script += "\n"
		}
		inputData = script
		command = shellBin
	}
	timeoutSeconds := 30
	pollMillis := 1000
	if !noWait {
		if req.Args["timeout-seconds"] != "" {
			parsedTimeout, parseErr := RequiredInt(req.Args, "timeout-seconds")
			if parseErr != nil {
				return nil, parseErr
			}
			timeoutSeconds = parsedTimeout
		}
		if req.Args["poll-interval-ms"] != "" {
			parsedPoll, parseErr := RequiredInt(req.Args, "poll-interval-ms")
			if parseErr != nil {
				return nil, parseErr
			}
			pollMillis = parsedPoll
		}
	}
	form := url.Values{}
	form.Set("command", command)
	if inputData != "" {
		form.Set("input-data", inputData)
	}
	path := fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec", url.PathEscape(node), vmid)
	data, err := client.PostFormData(ctx, path, form)
	if err != nil {
		return nil, qgaUnavailableError("agent_exec", err)
	}
	pid := extractExecPID(data)
	if pid == "" {
		return nil, apperr.New(apperr.CodeNetwork, "agent_exec response does not contain pid")
	}
	request := map[string]any{"node": node, "vmid": vmid, "command": originalCommand, "exec_command": command}
	if useShell {
		request["shell"] = true
	}
	if noWait {
		request["no_wait"] = true
	}
	if inputData != "" {
		request["input_data_len"] = len(inputData)
	}
	if noWait {
		result := map[string]any{"pid": pid, "exec_command": command, "shell": useShell, "no_wait": true}
		diagnostics := map[string]any{"no_wait": true, "wait_skipped": "agent_exec status polling disabled by --no-wait", "shell": useShell, "input_data_len": len(inputData)}
		return buildResult(req, request, result, diagnostics), nil
	}
	status, polls, err := pollAgentExecStatus(ctx, client, node, vmid, pid, time.Duration(timeoutSeconds)*time.Second, time.Duration(pollMillis)*time.Millisecond)
	if err != nil {
		return nil, err
	}
	result := map[string]any{"pid": pid, "status": status, "exec_command": command, "shell": useShell, "no_wait": false}
	diagnostics := map[string]any{"poll_count": polls, "timeout_seconds": timeoutSeconds, "no_wait": false, "shell": useShell, "input_data_len": len(inputData)}
	return buildResult(req, request, result, diagnostics), nil
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
