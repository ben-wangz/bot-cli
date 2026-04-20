package action

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

type WaitOptions struct {
	Timeout  time.Duration
	Interval time.Duration
}

func WaitTask(ctx context.Context, client *pveapi.Client, node string, upid string, options WaitOptions) (map[string]any, error) {
	if strings.TrimSpace(node) == "" || strings.TrimSpace(upid) == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "wait_task requires node and upid")
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	interval := options.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	path := fmt.Sprintf("/nodes/%s/tasks/%s/status", url.PathEscape(node), url.PathEscape(upid))
	for {
		select {
		case <-ctxWithTimeout.Done():
			return nil, apperr.New(apperr.CodeNetwork, "wait_task timeout exceeded")
		default:
		}

		data, err := client.GetData(ctxWithTimeout, path, url.Values{})
		if err != nil {
			return nil, err
		}
		statusMap, ok := data.(map[string]any)
		if !ok {
			return nil, apperr.New(apperr.CodeNetwork, "task status response is not an object")
		}
		status := strings.ToLower(strings.TrimSpace(asString(statusMap["status"])))
		if status == "stopped" {
			exitStatus := strings.TrimSpace(asString(statusMap["exitstatus"]))
			if strings.EqualFold(exitStatus, "OK") {
				return statusMap, nil
			}
			if exitStatus == "" {
				return nil, apperr.New(apperr.CodeNetwork, "task stopped without exitstatus")
			}
			return nil, apperr.New(apperr.CodeNetwork, "task failed with exitstatus: "+exitStatus)
		}
		time.Sleep(interval)
	}
}
