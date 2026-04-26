package capability

import (
	"context"
	"net/url"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/auth"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func runCreatePVEUserWithRoot(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	if err := ensureRootBootstrapScope(req.Scope); err != nil {
		return nil, err
	}
	userID, err := requiredUserID(req.Args)
	if err != nil {
		return nil, err
	}
	ifExists := strings.TrimSpace(strings.ToLower(req.Args["if-exists"]))
	if ifExists == "" {
		ifExists = "fail"
	}
	if !isOneOf(ifExists, "fail", "reuse") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "if-exists must be one of fail|reuse")
	}
	exists, existing, err := getUserByID(ctx, client, userID)
	if err != nil {
		return nil, err
	}
	request := map[string]any{"userid": userID, "if_exists": ifExists}
	if exists {
		if ifExists == "fail" {
			return nil, apperr.New(apperr.CodeInvalidArgs, "user already exists: "+userID)
		}
		result := map[string]any{"userid": userID, "created": false, "reused": true, "user": existing}
		diagnostics := map[string]any{"wait_skipped": "user already exists; reused existing user"}
		return buildResult(req, request, result, diagnostics), nil
	}
	form := url.Values{}
	form.Set("userid", userID)
	for _, key := range []string{"password", "comment", "email", "firstname", "lastname", "enable", "expire", "keys"} {
		if v := strings.TrimSpace(req.Args[key]); v != "" {
			form.Set(key, v)
			if key != "password" {
				request[key] = v
			}
		}
	}
	data, err := client.PostFormData(ctx, "/access/users", form)
	if err != nil {
		return nil, err
	}
	createdUser, _, inspectErr := getUserByID(ctx, client, userID)
	result := map[string]any{"userid": userID, "created": true, "reused": false, "api_response": data}
	if inspectErr == nil && createdUser {
		if user, ok := unwrapResultField(data).(map[string]any); ok {
			result["user"] = user
		}
	}
	return buildResult(req, request, result, map[string]any{}), nil
}

func runCreatePoolWithRoot(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	if err := ensureRootBootstrapScope(req.Scope); err != nil {
		return nil, err
	}
	poolID, err := RequiredString(req.Args, "poolid")
	if err != nil {
		return nil, err
	}
	ifExists := strings.TrimSpace(strings.ToLower(req.Args["if-exists"]))
	if ifExists == "" {
		ifExists = "fail"
	}
	if !isOneOf(ifExists, "fail", "reuse") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "if-exists must be one of fail|reuse")
	}
	exists, existing, err := getPoolByID(ctx, client, poolID)
	if err != nil {
		return nil, err
	}
	request := map[string]any{"poolid": poolID, "if_exists": ifExists}
	if comment := strings.TrimSpace(req.Args["comment"]); comment != "" {
		request["comment"] = comment
	}
	if exists {
		if ifExists == "fail" {
			return nil, apperr.New(apperr.CodeInvalidArgs, "pool already exists: "+poolID)
		}
		result := map[string]any{"poolid": poolID, "created": false, "reused": true, "pool": existing}
		diagnostics := map[string]any{"wait_skipped": "pool already exists; reused existing pool"}
		return buildResult(req, request, result, diagnostics), nil
	}
	form := url.Values{}
	form.Set("poolid", poolID)
	if comment := strings.TrimSpace(req.Args["comment"]); comment != "" {
		form.Set("comment", comment)
	}
	data, err := client.PostFormData(ctx, "/pools", form)
	if err != nil {
		return nil, err
	}
	poolExists, pool, inspectErr := getPoolByID(ctx, client, poolID)
	result := map[string]any{"poolid": poolID, "created": true, "reused": false, "api_response": data}
	if inspectErr == nil && poolExists {
		result["pool"] = pool
	}
	return buildResult(req, request, result, map[string]any{}), nil
}

func runGetUserACLBinding(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	if err := ensureRootBootstrapScope(req.Scope); err != nil {
		return nil, err
	}
	userID, err := requiredUserID(req.Args)
	if err != nil {
		return nil, err
	}
	pathFilter := strings.TrimSpace(req.Args["path"])
	roleFilter := strings.TrimSpace(req.Args["role"])
	rows, err := fetchACLRows(ctx, client, pathFilter)
	if err != nil {
		return nil, err
	}
	bindings := filterUserACLRows(rows, userID, pathFilter, roleFilter)
	request := map[string]any{"userid": userID}
	if pathFilter != "" {
		request["path"] = pathFilter
	}
	if roleFilter != "" {
		request["role"] = roleFilter
	}
	result := map[string]any{"userid": userID, "bindings": bindings, "count": len(bindings)}
	diagnostics := map[string]any{"acl_rows_total": len(rows), "acl_rows_matched": len(bindings)}
	return buildResult(req, request, result, diagnostics), nil
}

func runGrantUserACL(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	if err := ensureRootBootstrapScope(req.Scope); err != nil {
		return nil, err
	}
	userID, pathValue, roleValue, propagate, err := parseACLMutationArgs(req.Args)
	if err != nil {
		return nil, err
	}
	rows, err := fetchACLRows(ctx, client, pathValue)
	if err != nil {
		return nil, err
	}
	matched := filterUserACLRows(rows, userID, pathValue, roleValue)
	request := map[string]any{"userid": userID, "path": pathValue, "role": roleValue, "propagate": propagate}
	if len(matched) > 0 {
		result := map[string]any{"userid": userID, "path": pathValue, "role": roleValue, "granted": true, "changed": false}
		diagnostics := map[string]any{"wait_skipped": "acl binding already exists"}
		return buildResult(req, request, result, diagnostics), nil
	}
	form := url.Values{}
	form.Set("path", pathValue)
	form.Set("users", userID)
	form.Set("roles", roleValue)
	form.Set("propagate", boolTo01(propagate))
	data, err := client.PutFormData(ctx, "/access/acl", form)
	if err != nil {
		return nil, err
	}
	result := map[string]any{"userid": userID, "path": pathValue, "role": roleValue, "granted": true, "changed": true, "api_response": data}
	return buildResult(req, request, result, map[string]any{}), nil
}

func runRevokeUserACL(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	if err := ensureRootBootstrapScope(req.Scope); err != nil {
		return nil, err
	}
	userID, pathValue, roleValue, _, err := parseACLMutationArgs(req.Args)
	if err != nil {
		return nil, err
	}
	rows, err := fetchACLRows(ctx, client, pathValue)
	if err != nil {
		return nil, err
	}
	matched := filterUserACLRows(rows, userID, pathValue, roleValue)
	request := map[string]any{"userid": userID, "path": pathValue, "role": roleValue}
	if len(matched) == 0 {
		result := map[string]any{"userid": userID, "path": pathValue, "role": roleValue, "revoked": true, "changed": false}
		diagnostics := map[string]any{"wait_skipped": "acl binding does not exist"}
		return buildResult(req, request, result, diagnostics), nil
	}
	form := url.Values{}
	form.Set("path", pathValue)
	form.Set("users", userID)
	form.Set("roles", roleValue)
	data, deleteErr := client.DeleteFormData(ctx, "/access/acl", form)
	if deleteErr != nil {
		fallbackForm := url.Values{}
		fallbackForm.Set("path", pathValue)
		fallbackForm.Set("users", userID)
		fallbackForm.Set("roles", roleValue)
		fallbackForm.Set("delete", "1")
		data, err = client.PutFormData(ctx, "/access/acl", fallbackForm)
		if err != nil {
			return nil, deleteErr
		}
	}
	result := map[string]any{"userid": userID, "path": pathValue, "role": roleValue, "revoked": true, "changed": true, "api_response": data}
	return buildResult(req, request, result, map[string]any{}), nil
}

func ensureRootBootstrapScope(scope string) error {
	if scope == auth.ScopeRoot || scope == auth.ScopeRootToken {
		return nil
	}
	return apperr.New(apperr.CodeAuth, "action requires --auth-scope root or root-token")
}
