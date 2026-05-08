package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/ttsapi"
)

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
	client, err := ttsapi.New(opts.APIBaseURL, opts.APIKey, opts.Timeout)
	if err != nil {
		if !commandNeedsClient(tail) {
			client = nil
		} else {
			return printError(err, stderr)
		}
	}
	runtime := commandRuntime{Opts: opts, Client: client, Stdout: stdout}
	if err := dispatchCommand(runtime, tail); err != nil {
		return printError(err, stderr)
	}
	return 0
}

func printError(err error, stderr io.Writer) int {
	payload := map[string]any{
		"ok":      false,
		"request": map[string]any{},
		"result":  map[string]any{},
		"error": map[string]any{
			"code":      "internal_error",
			"message":   err.Error(),
			"retryable": false,
		},
		"diagnostics": map[string]any{},
	}
	if typed, ok := err.(*apperr.Error); ok {
		payload["error"] = map[string]any{
			"code":      string(typed.Code),
			"message":   typed.Error(),
			"retryable": typed.Code == apperr.CodeNetwork,
		}
	}
	encoded, encodeErr := json.MarshalIndent(payload, "", "  ")
	if encodeErr != nil {
		_, _ = io.WriteString(stderr, fmt.Sprintf("error: %s\n", err.Error()))
		return apperr.ExitCode(err)
	}
	_, _ = io.WriteString(stderr, string(encoded)+"\n")
	return apperr.ExitCode(err)
}

func commandNeedsClient(tail []string) bool {
	if len(tail) == 0 {
		return false
	}
	if tail[0] != "capability" {
		return false
	}
	if len(tail) == 1 {
		return false
	}
	sub := strings.TrimSpace(tail[1])
	if sub == "" || sub == "--help" || sub == "-h" || sub == "list" || sub == "describe" || sub == "suggest_voices" {
		return false
	}
	for _, token := range tail[2:] {
		if token == "--help" || token == "-h" {
			return false
		}
	}
	return true
}
