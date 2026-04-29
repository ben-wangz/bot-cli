package cli

import (
	"strconv"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
)

func parseGlobalOptions(args []string) (GlobalOptions, []string, error) {
	opts := defaultGlobalOptions()
	tail := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		token := args[i]
		if !strings.HasPrefix(token, "--") {
			tail = append(tail, args[i:]...)
			break
		}
		key := strings.TrimPrefix(token, "--")
		if key == "help" {
			opts.Help = true
			continue
		}
		if i+1 >= len(args) {
			return opts, nil, apperr.New(apperr.CodeInvalidArgs, "missing value for --"+key)
		}
		value := args[i+1]
		i++
		switch key {
		case "rpc-endpoint":
			opts.RPCEndpoint = value
		case "rpc-secret":
			opts.RPCSecret = value
		case "timeout":
			seconds, err := strconv.Atoi(value)
			if err != nil || seconds <= 0 {
				return opts, nil, apperr.New(apperr.CodeInvalidArgs, "timeout must be a positive integer")
			}
			opts.Timeout = time.Duration(seconds) * time.Second
		case "output":
			opts.Output = strings.ToLower(strings.TrimSpace(value))
		default:
			return opts, nil, apperr.New(apperr.CodeInvalidArgs, "unknown global option --"+key)
		}
	}
	return opts, tail, nil
}
