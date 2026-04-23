package action

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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
	var lastStatus map[string]any
	for {
		select {
		case <-ctxWithTimeout.Done():
			logTail := fetchTaskLogTail(ctx, client, node, upid, 20)
			statusSummary := summarizeTaskStatus(lastStatus)
			if logTail == "" {
				return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("wait_task timeout exceeded (node=%s upid=%s%s)", node, upid, statusSummary))
			}
			return nil, apperr.New(apperr.CodeNetwork, fmt.Sprintf("wait_task timeout exceeded (node=%s upid=%s%s) task_log_tail=%s", node, upid, statusSummary, logTail))
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
		lastStatus = statusMap
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

func summarizeTaskStatus(status map[string]any) string {
	if status == nil {
		return ""
	}
	s := strings.TrimSpace(asString(status["status"]))
	e := strings.TrimSpace(asString(status["exitstatus"]))
	parts := []string{}
	if s != "" {
		parts = append(parts, " status="+s)
	}
	if e != "" {
		parts = append(parts, " exitstatus="+e)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "")
}

func fetchTaskLogTail(ctx context.Context, client *pveapi.Client, node string, upid string, maxLines int) string {
	if maxLines <= 0 {
		maxLines = 20
	}
	path := fmt.Sprintf("/nodes/%s/tasks/%s/log", url.PathEscape(node), url.PathEscape(upid))
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return ""
	}
	entries, ok := data.([]any)
	if !ok || len(entries) == 0 {
		return ""
	}
	start := len(entries) - maxLines
	if start < 0 {
		start = 0
	}
	lines := make([]string, 0, len(entries)-start)
	for i := start; i < len(entries); i++ {
		obj, ok := entries[i].(map[string]any)
		if !ok {
			continue
		}
		text := strings.TrimSpace(asString(obj["t"]))
		if text == "" {
			continue
		}
		n := strings.TrimSpace(asString(obj["n"]))
		if n != "" {
			if _, convErr := strconv.Atoi(n); convErr == nil {
				lines = append(lines, n+":"+text)
				continue
			}
		}
		lines = append(lines, text)
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, " | ")
}
