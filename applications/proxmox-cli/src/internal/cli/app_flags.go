package cli

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func parseGlobalOptions(args []string) (GlobalOptions, []string, error) {
	opts := defaultGlobalOptions()
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--" {
			return opts, args[i+1:], nil
		}
		if !strings.HasPrefix(arg, "-") {
			return opts, args[i:], nil
		}
		if arg == "-h" || arg == "--help" {
			opts.Help = true
			i++
			continue
		}

		name, value, hasValue := splitFlag(arg)
		switch name {
		case "--api-base":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.APIBase = v
			i = next
		case "--auth-scope":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthScope = v
			i = next
		case "--auth-file":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthFile = v
			i = next
		case "--auth-user":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthUser = v
			i = next
		case "--auth-password":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthPass = v
			i = next
		case "--auth-token":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthToken = v
			i = next
		case "--auth-ticket":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthTicket = v
			i = next
		case "--auth-csrf-token":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.AuthCSRF = v
			i = next
		case "--timeout":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			parsed, parseErr := parseTimeout(v)
			if parseErr != nil {
				return opts, nil, parseErr
			}
			opts.Timeout = parsed
			i = next
		case "--output":
			v, next, err := takeStringFlag(value, hasValue, args, i)
			if err != nil {
				return opts, nil, err
			}
			opts.Output = strings.ToLower(v)
			i = next
		case "--insecure-tls":
			opts.InsecureTLS = true
			i++
		case "--wait":
			opts.Wait = true
			i++
		case "--dry-run":
			opts.DryRun = true
			i++
		case "--debug":
			opts.Debug = true
			i++
		default:
			return opts, nil, apperr.New(apperr.CodeInvalidArgs, "unknown global flag: "+name)
		}
	}
	return opts, []string{}, nil
}

func splitFlag(arg string) (string, string, bool) {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return arg, "", false
}

func takeStringFlag(inlineValue string, hasValue bool, args []string, index int) (string, int, error) {
	if hasValue {
		return inlineValue, index + 1, nil
	}
	if index+1 >= len(args) {
		return "", 0, apperr.New(apperr.CodeInvalidArgs, "missing value for flag "+args[index])
	}
	return args[index+1], index + 2, nil
}

func parseTimeout(value string) (time.Duration, error) {
	if strings.ContainsAny(value, "hms") {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return 0, apperr.Wrap(apperr.CodeInvalidArgs, "invalid timeout", err)
		}
		return parsed, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, apperr.Wrap(apperr.CodeInvalidArgs, "invalid timeout", err)
	}
	if seconds <= 0 {
		return 0, apperr.New(apperr.CodeInvalidArgs, "timeout must be positive")
	}
	return time.Duration(seconds) * time.Second, nil
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func tryPrintCommandHelp(args []string, stdout io.Writer) bool {
	if len(args) == 0 {
		return false
	}
	if !hasHelp(args[1:]) {
		return false
	}
	switch args[0] {
	case "capability":
		_, _ = io.WriteString(stdout, capabilityHelp())
		return true
	case "workflow":
		_, _ = io.WriteString(stdout, workflowHelp())
		return true
	case "console":
		_, _ = io.WriteString(stdout, consoleHelp())
		return true
	case "auth":
		_, _ = io.WriteString(stdout, authHelp())
		return true
	default:
		return false
	}
}
