package workflow

import (
	"io"
	"sort"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

type CommandStepDeps struct {
	Stderr      io.Writer
	ExecCommand func(args []string, wait bool) (map[string]any, error)
}

func RunCommandStep(name string, stepArgs map[string]string, wait bool, deps CommandStepDeps) (map[string]any, error) {
	if deps.ExecCommand == nil {
		return nil, apperr.New(apperr.CodeInternal, "workflow dependency missing: exec command")
	}
	LogStep(deps.Stderr, "start", name, stepArgs)
	args := make([]string, 0, 1+len(stepArgs)*2)
	args = append(args, name)
	keys := make([]string, 0, len(stepArgs))
	for key := range stepArgs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "--"+key, stepArgs[key])
	}
	payload, err := deps.ExecCommand(args, wait)
	if err != nil {
		LogStepError(deps.Stderr, name, err)
		return nil, err
	}
	LogStep(deps.Stderr, "done", name, map[string]string{"ok": "true"})
	return payload, nil
}
