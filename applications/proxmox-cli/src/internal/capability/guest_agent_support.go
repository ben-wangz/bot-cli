package capability

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/taskwait"
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
	var latest map[string]any
	polls, err := taskwait.Poll(ctx, taskwait.PollOptions{
		Timeout:            timeout,
		Interval:           interval,
		TimeoutMessage:     "agent_exec timed out waiting for exit status",
		InterruptedMessage: "agent_exec polling interrupted",
	}, func(pollCtx context.Context) (bool, error) {
		status, err := getAgentExecStatus(pollCtx, client, node, vmid, pid)
		if err != nil {
			return false, err
		}
		latest = status
		return toBool(status["exited"]), nil
	})
	if err != nil {
		return nil, polls, err
	}
	return latest, polls, nil
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
