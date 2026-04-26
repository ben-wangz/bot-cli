package taskwait

import (
	"context"
	"errors"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

type PollOptions struct {
	Timeout            time.Duration
	Interval           time.Duration
	TimeoutMessage     string
	InterruptedMessage string
	TimeoutError       func() error
	InterruptedError   func(cause error) error
}

func Poll(ctx context.Context, options PollOptions, probe func(ctx context.Context) (bool, error)) (int, error) {
	if probe == nil {
		return 0, apperr.New(apperr.CodeInternal, "poll probe is required")
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	interval := options.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	timeoutMessage := options.TimeoutMessage
	if timeoutMessage == "" {
		timeoutMessage = "poll timeout exceeded"
	}
	interruptedMessage := options.InterruptedMessage
	if interruptedMessage == "" {
		interruptedMessage = "poll interrupted"
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	polls := 0
	for {
		done, err := probe(ctxWithTimeout)
		if err != nil {
			return polls, err
		}
		polls++
		if done {
			return polls, nil
		}
		select {
		case <-ctxWithTimeout.Done():
			if errors.Is(ctxWithTimeout.Err(), context.DeadlineExceeded) {
				if options.TimeoutError != nil {
					return polls, options.TimeoutError()
				}
				return polls, apperr.New(apperr.CodeNetwork, timeoutMessage)
			}
			if options.InterruptedError != nil {
				return polls, options.InterruptedError(ctxWithTimeout.Err())
			}
			return polls, apperr.Wrap(apperr.CodeNetwork, interruptedMessage, ctxWithTimeout.Err())
		case <-time.After(interval):
		}
	}
}
