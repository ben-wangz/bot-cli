package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/imageapi"
	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/output"
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
	client, err := imageapi.New(opts.APIBaseURL, opts.APIKey, opts.Timeout)
	if err != nil {
		return printError(err, stderr)
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
