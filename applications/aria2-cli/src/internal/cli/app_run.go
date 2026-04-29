package cli

import (
	"fmt"
	"io"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/output"
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
	client, err := aria2rpc.New(opts.RPCEndpoint, opts.RPCSecret, opts.Timeout)
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
	if err == nil {
		return 0
	}
	_, _ = io.WriteString(stderr, fmt.Sprintf("error: %s\n", err.Error()))
	return apperr.ExitCode(err)
}
