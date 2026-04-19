package cli

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/auth"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/redact"
)

type GlobalOptions struct {
	APIBase     string
	AuthScope   string
	AuthFile    string
	AuthUser    string
	AuthPass    string
	AuthToken   string
	AuthTicket  string
	AuthCSRF    string
	InsecureTLS bool
	Timeout     time.Duration
	Wait        bool
	DryRun      bool
	Output      string
	Debug       bool
	Help        bool
}

func defaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		APIBase:   "https://127.0.0.1:8006/api2/json",
		AuthScope: auth.ScopeUser,
		Timeout:   30 * time.Second,
		Output:    output.FormatJSON,
	}
}

func Run(args []string, stdout, stderr io.Writer) int {
	opts, tail, err := parseGlobalOptions(args)
	if err != nil {
		return printError(err, stderr)
	}

	if opts.Help || len(tail) == 0 {
		_, _ = io.WriteString(stdout, rootHelp())
		return 0
	}

	if err := output.ValidateFormat(opts.Output); err != nil {
		return printError(err, stderr)
	}

	creds, sources, err := auth.Load(auth.FlagCredentials{
		User:      opts.AuthUser,
		Password:  opts.AuthPass,
		Token:     opts.AuthToken,
		Ticket:    opts.AuthTicket,
		CSRFToken: opts.AuthCSRF,
	}, opts.AuthFile)
	if err != nil {
		return printError(err, stderr)
	}
	if err := auth.ValidateScope(opts.AuthScope, creds); err != nil {
		return printError(err, stderr)
	}

	headers := map[string]string{"Accept": "application/json"}
	if creds.Token != "" {
		headers["Authorization"] = "PVEAPIToken=" + creds.Token
	}
	if creds.Ticket != "" {
		headers["Cookie"] = "PVEAuthCookie=" + creds.Ticket
	}
	if creds.CSRFToken != "" {
		headers["CSRFPreventionToken"] = creds.CSRFToken
	}

	var logger *log.Logger
	if opts.Debug {
		logger = log.New(stderr, "[debug] ", 0)
	}

	client, err := pveapi.New(pveapi.Config{
		BaseURL:     opts.APIBase,
		Timeout:     opts.Timeout,
		InsecureTLS: opts.InsecureTLS,
		Headers:     headers,
		Logger:      logger,
	})
	if err != nil {
		return printError(err, stderr)
	}

	runtime := commandRuntime{
		Opts:    opts,
		Creds:   creds,
		Sources: sources,
		Client:  client,
		Stdout:  stdout,
		Stderr:  stderr,
	}

	err = dispatchCommand(runtime, tail)
	if err != nil {
		return printError(err, stderr)
	}
	return 0
}

type commandRuntime struct {
	Opts    GlobalOptions
	Creds   auth.Credentials
	Sources auth.Sources
	Client  *pveapi.Client
	Stdout  io.Writer
	Stderr  io.Writer
}

func dispatchCommand(rt commandRuntime, args []string) error {
	command := args[0]
	commandArgs := []string{}
	if len(args) > 1 {
		commandArgs = args[1:]
	}

	var payload map[string]any
	var err error
	switch command {
	case "action":
		payload, err = runActionCommand(rt, commandArgs)
	case "workflow":
		payload, err = runWorkflowCommand(rt, commandArgs)
	case "console":
		payload, err = runConsoleCommand(rt, commandArgs)
	case "auth":
		payload, err = runAuthCommand(rt, commandArgs)
	case "help", "--help", "-h":
		_, _ = io.WriteString(rt.Stdout, rootHelp())
		return nil
	default:
		return apperr.New(apperr.CodeInvalidArgs, "unknown command: "+command)
	}
	if err != nil {
		return err
	}
	return output.Render(rt.Stdout, rt.Opts.Output, payload)
}

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

func runActionCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, actionHelp())
		return nil, nil
	}
	if len(args) == 0 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "action name is required")
	}
	return map[string]any{
		"command":     "action",
		"action":      args[0],
		"args":        args[1:],
		"scope":       rt.Opts.AuthScope,
		"dry_run":     rt.Opts.DryRun,
		"wait":        rt.Opts.Wait,
		"status":      "skeleton-ready",
		"next_action": "implement action registry in phase 1+",
	}, nil
}

func runWorkflowCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, workflowHelp())
		return nil, nil
	}
	if len(args) == 0 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "workflow name is required")
	}
	return map[string]any{
		"command":  "workflow",
		"workflow": args[0],
		"args":     args[1:],
		"status":   "skeleton-ready",
	}, nil
}

func runConsoleCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, consoleHelp())
		return nil, nil
	}
	subcommand := ""
	rest := []string{}
	if len(args) > 0 {
		subcommand = args[0]
		rest = args[1:]
	}
	if subcommand == "" {
		subcommand = "script"
	}
	return map[string]any{
		"command":    "console",
		"subcommand": subcommand,
		"args":       rest,
		"status":     "skeleton-ready",
	}, nil
}

func runAuthCommand(rt commandRuntime, args []string) (map[string]any, error) {
	if hasHelp(args) {
		_, _ = io.WriteString(rt.Stdout, authHelp())
		return nil, nil
	}
	subcommand := "inspect"
	if len(args) > 0 {
		subcommand = args[0]
	}
	switch subcommand {
	case "inspect":
		return map[string]any{
			"command":        "auth",
			"subcommand":     "inspect",
			"auth_scope":     rt.Opts.AuthScope,
			"credentials":    rt.Creds.SafeSummary(),
			"source_summary": rt.Sources.Summary(),
		}, nil
	case "example-auth-file":
		return map[string]any{
			"auth_file_example": auth.ExampleAuthFile(),
		}, nil
	default:
		return nil, apperr.New(apperr.CodeInvalidArgs, "unknown auth subcommand: "+subcommand)
	}
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func printError(err error, stderr io.Writer) int {
	if err == nil {
		return 0
	}
	message := redact.String(err.Error())
	if errors.Is(err, os.ErrNotExist) {
		message = "[config_error] file not found"
	}
	_, _ = io.WriteString(stderr, fmt.Sprintf("error: %s\n", message))
	return apperr.ExitCode(err)
}
