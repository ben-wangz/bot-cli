package workflow

import (
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const bootstrapUserPoolACLWorkflowName = "bootstrap-bot-user-pool-acl"

type BootstrapUserPoolACLDeps struct {
	AuthScope string
	RunStep   func(name string, stepArgs map[string]string, wait bool) (map[string]any, error)
}

func RunBootstrapUserPoolACL(args map[string]string, deps BootstrapUserPoolACLDeps) (map[string]any, error) {
	if deps.RunStep == nil {
		return nil, apperr.New(apperr.CodeInternal, "workflow dependency missing: run step")
	}
	if err := EnsureAllowedArgs(args, "userid", "poolid", "password", "pool-comment", "user-comment", "if-exists", "sdn-acl-path", "sdn-role"); err != nil {
		return nil, err
	}
	if deps.AuthScope != "root" && deps.AuthScope != "root-token" {
		return nil, apperr.New(apperr.CodeAuth, "workflow bootstrap-bot-user-pool-acl requires --auth-scope root or root-token")
	}
	userID, err := RequiredString(args, "userid")
	if err != nil {
		return nil, err
	}
	if !strings.Contains(userID, "@") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "userid must include realm suffix, for example user@pve")
	}
	poolID, err := RequiredString(args, "poolid")
	if err != nil {
		return nil, err
	}
	ifExists := strings.TrimSpace(strings.ToLower(args["if-exists"]))
	if ifExists == "" {
		ifExists = "reuse"
	}
	if ifExists != "fail" && ifExists != "reuse" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "if-exists must be one of fail|reuse")
	}
	sdnACLPath := strings.TrimSpace(args["sdn-acl-path"])
	if sdnACLPath == "" {
		sdnACLPath = "/sdn/zones/localnetwork"
	}
	if !strings.HasPrefix(sdnACLPath, "/") {
		return nil, apperr.New(apperr.CodeInvalidArgs, "sdn-acl-path must start with /")
	}
	sdnRole := strings.TrimSpace(args["sdn-role"])
	if sdnRole == "" {
		sdnRole = "PVEAdmin"
	}
	password := strings.TrimSpace(args["password"])
	passwordGenerated := false
	if password == "" {
		generated, genErr := GeneratePassword(20)
		if genErr != nil {
			return nil, genErr
		}
		password = generated
		passwordGenerated = true
	}

	steps := make([]map[string]any, 0, 8)

	createUserArgs := map[string]string{"userid": userID, "password": password, "if-exists": ifExists}
	if userComment := strings.TrimSpace(args["user-comment"]); userComment != "" {
		createUserArgs["comment"] = userComment
	}
	createUserPayload, err := deps.RunStep("create_pve_user_with_root", createUserArgs, false)
	if err != nil {
		return nil, err
	}
	steps = append(steps, StepSummary("create_pve_user_with_root", createUserPayload))

	createPoolArgs := map[string]string{"poolid": poolID, "if-exists": ifExists}
	if poolComment := strings.TrimSpace(args["pool-comment"]); poolComment != "" {
		createPoolArgs["comment"] = poolComment
	}
	createPoolPayload, err := deps.RunStep("create_pool_with_root", createPoolArgs, false)
	if err != nil {
		return nil, err
	}
	steps = append(steps, StepSummary("create_pool_with_root", createPoolPayload))

	grants := []map[string]any{
		{"path": "/pool/" + poolID, "role": "PVEAdmin", "propagate": true},
		{"path": "/", "role": "PVEAuditor", "propagate": true},
		{"path": "/storage", "role": "PVEDatastoreAdmin", "propagate": true},
		{"path": sdnACLPath, "role": sdnRole, "propagate": true},
	}
	for _, grant := range grants {
		path := asStringValue(grant["path"])
		role := asStringValue(grant["role"])
		grantPayload, grantErr := deps.RunStep("grant_user_acl", map[string]string{
			"userid":    userID,
			"path":      path,
			"role":      role,
			"propagate": "1",
		}, false)
		if grantErr != nil {
			return nil, grantErr
		}
		steps = append(steps, StepSummary("grant_user_acl", grantPayload))
	}

	bindingPayload, err := deps.RunStep("get_user_acl_binding", map[string]string{"userid": userID}, false)
	if err != nil {
		return nil, err
	}
	steps = append(steps, StepSummary("get_user_acl_binding", bindingPayload))
	bindingResult, _ := bindingPayload["result"].(map[string]any)
	bindingRows := 0
	if count, ok := bindingResult["count"].(int); ok {
		bindingRows = count
	} else if n, ok := AnyToInt(bindingResult["count"]); ok {
		bindingRows = n
	}

	result := map[string]any{
		"userid":             userID,
		"poolid":             poolID,
		"password":           password,
		"password_generated": passwordGenerated,
		"grants":             grants,
		"steps":              steps,
	}
	request := map[string]any{"userid": userID, "poolid": poolID, "if_exists": ifExists}
	if strings.TrimSpace(args["user-comment"]) != "" {
		request["user_comment"] = strings.TrimSpace(args["user-comment"])
	}
	if strings.TrimSpace(args["pool-comment"]) != "" {
		request["pool_comment"] = strings.TrimSpace(args["pool-comment"])
	}
	request["sdn_acl_path"] = sdnACLPath
	request["sdn_role"] = sdnRole
	diagnostics := map[string]any{"step_count": len(steps), "acl_bindings_total": bindingRows}

	return map[string]any{
		"workflow":    bootstrapUserPoolACLWorkflowName,
		"ok":          true,
		"scope":       deps.AuthScope,
		"request":     request,
		"result":      result,
		"diagnostics": diagnostics,
	}, nil
}
