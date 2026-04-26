package capability

import (
	"context"
	"net/url"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func requiredUserID(args map[string]string) (string, error) {
	userID, err := RequiredString(args, "userid")
	if err != nil {
		return "", err
	}
	if !strings.Contains(userID, "@") {
		return "", apperr.New(apperr.CodeInvalidArgs, "userid must include realm suffix, for example user@pve")
	}
	return userID, nil
}

func getUserByID(ctx context.Context, client *pveapi.Client, userID string) (bool, map[string]any, error) {
	path := "/access/users/" + url.PathEscape(userID)
	data, err := client.GetData(ctx, path, url.Values{})
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "status 404") || strings.Contains(lower, "not found") || strings.Contains(lower, "does not exist") || strings.Contains(lower, "no such user") {
			return false, nil, nil
		}
		return false, nil, err
	}
	payload := firstObject(unwrapResultField(data))
	if payload == nil {
		if m, ok := unwrapResultField(data).(map[string]any); ok {
			payload = m
		}
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return true, payload, nil
}

func getPoolByID(ctx context.Context, client *pveapi.Client, poolID string) (bool, map[string]any, error) {
	data, err := client.GetData(ctx, "/pools", url.Values{})
	if err != nil {
		return false, nil, err
	}
	raw := unwrapResultField(data)
	list, ok := raw.([]any)
	if !ok {
		return false, nil, nil
	}
	for _, row := range list {
		m, isMap := row.(map[string]any)
		if !isMap {
			continue
		}
		if strings.TrimSpace(asString(m["poolid"])) == poolID {
			return true, m, nil
		}
	}
	return false, nil, nil
}

func fetchACLRows(ctx context.Context, client *pveapi.Client, pathFilter string) ([]map[string]any, error) {
	_ = pathFilter
	data, err := client.GetData(ctx, "/access/acl", url.Values{})
	if err != nil {
		return nil, err
	}
	raw := unwrapResultField(data)
	list, ok := raw.([]any)
	if !ok {
		if m, isMap := raw.(map[string]any); isMap {
			return []map[string]any{m}, nil
		}
		return []map[string]any{}, nil
	}
	rows := make([]map[string]any, 0, len(list))
	for _, row := range list {
		if m, ok := row.(map[string]any); ok {
			rows = append(rows, m)
		}
	}
	return rows, nil
}

func filterUserACLRows(rows []map[string]any, userID string, pathFilter string, roleFilter string) []map[string]any {
	result := make([]map[string]any, 0)
	for _, row := range rows {
		ugid := strings.TrimSpace(asString(row["ugid"]))
		if ugid == "" {
			ugid = strings.TrimSpace(asString(row["userid"]))
		}
		if ugid != userID {
			continue
		}
		if strings.TrimSpace(pathFilter) != "" && strings.TrimSpace(asString(row["path"])) != pathFilter {
			continue
		}
		if strings.TrimSpace(roleFilter) != "" && strings.TrimSpace(asString(row["roleid"])) != roleFilter {
			continue
		}
		result = append(result, row)
	}
	return result
}

func parseACLMutationArgs(args map[string]string) (string, string, string, bool, error) {
	userID, err := requiredUserID(args)
	if err != nil {
		return "", "", "", false, err
	}
	pathValue, err := RequiredString(args, "path")
	if err != nil {
		return "", "", "", false, err
	}
	roleValue := strings.TrimSpace(args["role"])
	if roleValue == "" {
		roleValue = strings.TrimSpace(args["roles"])
	}
	if roleValue == "" {
		return "", "", "", false, apperr.New(apperr.CodeInvalidArgs, "missing required capability arg --role")
	}
	propagate, err := parseOptionalBoolArg(args, "propagate")
	if err != nil {
		return "", "", "", false, err
	}
	if strings.TrimSpace(args["propagate"]) == "" {
		propagate = true
	}
	return userID, pathValue, roleValue, propagate, nil
}

func boolTo01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
