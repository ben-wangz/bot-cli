package storagecap

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func containsString(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

func getStorageConfig(ctx context.Context, client *pveapi.Client, node string, storage string) (map[string]any, error) {
	path := fmt.Sprintf("/nodes/%s/storage/%s", url.PathEscape(node), url.PathEscape(storage))
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		return nil, err
	}
	config := firstObject(unwrapResultField(data))
	if config != nil && strings.TrimSpace(asString(config["content"])) != "" {
		return config, nil
	}
	statusPath := fmt.Sprintf("/nodes/%s/storage/%s/status", url.PathEscape(node), url.PathEscape(storage))
	statusData, statusErr := client.GetData(ctx, statusPath, url.Values{})
	if statusErr == nil {
		config = firstObject(unwrapResultField(statusData))
		if config != nil && strings.TrimSpace(asString(config["content"])) != "" {
			return config, nil
		}
	}
	listPath := fmt.Sprintf("/nodes/%s/storage", url.PathEscape(node))
	listData, listErr := client.GetData(ctx, listPath, url.Values{})
	if listErr == nil {
		if list, ok := listData.([]any); ok {
			for _, entry := range list {
				row, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if strings.TrimSpace(asString(row["storage"])) == storage {
					return row, nil
				}
			}
		}
	}
	if config == nil {
		return nil, apperr.New(apperr.CodeNetwork, "storage response is not an object")
	}
	return config, nil
}

func storageVolumeExists(ctx context.Context, client *pveapi.Client, node string, storage string, expectedVolID string) (bool, error) {
	path := fmt.Sprintf("/nodes/%s/storage/%s/content", url.PathEscape(node), url.PathEscape(storage))
	query := url.Values{}
	query.Set("content", "iso")
	data, err := client.GetData(ctx, path, query)
	if err != nil {
		return false, err
	}
	list, ok := data.([]any)
	if !ok {
		return false, nil
	}
	for _, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(asString(entry["volid"])) == expectedVolID {
			return true, nil
		}
	}
	return false, nil
}

func firstObject(data any) map[string]any {
	if m, ok := data.(map[string]any); ok {
		return m
	}
	list, ok := data.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, _ := list[0].(map[string]any)
	return m
}
