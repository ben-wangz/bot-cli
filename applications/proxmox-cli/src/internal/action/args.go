package action

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func ParseArgs(args []string) (map[string]string, error) {
	result := map[string]string{}
	for i := 0; i < len(args); i++ {
		name, inline, hasInline, err := splitActionArg(args[i])
		if err != nil {
			return nil, err
		}
		if hasInline {
			result[name] = inline
			continue
		}
		if i+1 >= len(args) {
			return nil, apperr.New(apperr.CodeInvalidArgs, "missing value for action arg --"+name)
		}
		result[name] = args[i+1]
		i++
	}
	return result, nil
}

func splitActionArg(token string) (name string, value string, hasValue bool, err error) {
	if !strings.HasPrefix(token, "--") {
		return "", "", false, apperr.New(apperr.CodeInvalidArgs, "action args must be --key value or --key=value")
	}
	trimmed := strings.TrimPrefix(token, "--")
	if trimmed == "" {
		return "", "", false, apperr.New(apperr.CodeInvalidArgs, "invalid empty action arg")
	}
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true, nil
	}
	return trimmed, "", false, nil
}

func RequiredNode(args map[string]string) (string, error) {
	node := strings.TrimSpace(args["node"])
	if node == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --node")
	}
	for _, r := range node {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", apperr.New(apperr.CodeInvalidArgs, "node contains invalid character")
	}
	return node, nil
}

func RequiredVMID(args map[string]string) (int, error) {
	vmidRaw := strings.TrimSpace(args["vmid"])
	if vmidRaw == "" {
		return 0, apperr.New(apperr.CodeInvalidArgs, "missing required action arg --vmid")
	}
	vmid, err := strconv.Atoi(vmidRaw)
	if err != nil || vmid <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, "vmid must be a positive integer")
	}
	return vmid, nil
}

func RequiredUPID(args map[string]string) (string, error) {
	upid := strings.TrimSpace(args["upid"])
	if upid == "" {
		return "", apperr.New(apperr.CodeInvalidArgs, "missing required action arg --upid")
	}
	if !strings.HasPrefix(strings.ToUpper(upid), "UPID:") {
		return "", apperr.New(apperr.CodeInvalidArgs, "upid must start with UPID:")
	}
	return upid, nil
}

func IsPhase1Action(name string) bool {
	switch name {
	case "list_nodes", "list_cluster_resources", "list_vms_by_node", "get_vm_config", "get_effective_permissions",
		"get_task_status", "get_next_vmid", "get_vm_status", "list_tasks_by_vmid":
		return true
	default:
		return false
	}
}
