package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const (
	ScopeUser      = "user"
	ScopeRootToken = "root-token"
	ScopeRoot      = "root"
)

type Credentials struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	Token     string `json:"token"`
	Ticket    string `json:"ticket"`
	CSRFToken string `json:"csrf_token"`
}

type Sources struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	Token     string `json:"token"`
	Ticket    string `json:"ticket"`
	CSRFToken string `json:"csrf_token"`
}

type FlagCredentials struct {
	User      string
	Password  string
	Token     string
	Ticket    string
	CSRFToken string
}

func Load(flagCreds FlagCredentials, authFile string) (Credentials, Sources, error) {
	fileCreds, err := readAuthFile(authFile)
	if err != nil {
		return Credentials{}, Sources{}, err
	}

	envCreds := Credentials{
		User:      strings.TrimSpace(os.Getenv("PVE_USER")),
		Password:  strings.TrimSpace(os.Getenv("PVE_PASSWORD")),
		Token:     strings.TrimSpace(os.Getenv("PVE_TOKEN")),
		Ticket:    strings.TrimSpace(os.Getenv("PVE_TICKET")),
		CSRFToken: strings.TrimSpace(os.Getenv("PVE_CSRF_TOKEN")),
	}

	flagValues := Credentials{
		User:      strings.TrimSpace(flagCreds.User),
		Password:  strings.TrimSpace(flagCreds.Password),
		Token:     strings.TrimSpace(flagCreds.Token),
		Ticket:    strings.TrimSpace(flagCreds.Ticket),
		CSRFToken: strings.TrimSpace(flagCreds.CSRFToken),
	}

	merged, sources := mergeCredentials(fileCreds, envCreds, flagValues)
	if merged.User == "" && merged.Token == "" && merged.Ticket == "" {
		return Credentials{}, Sources{}, apperr.New(apperr.CodeAuth, "missing credentials: provide flags, env, or auth-file")
	}
	return merged, sources, nil
}

func ValidateScope(scope string, creds Credentials) error {
	if scope != ScopeUser && scope != ScopeRootToken && scope != ScopeRoot {
		return apperr.New(apperr.CodeInvalidArgs, "auth-scope must be one of user, root-token, root")
	}

	switch scope {
	case ScopeUser:
		if creds.User == "" {
			return apperr.New(apperr.CodeAuth, "scope user requires user identity")
		}
		if creds.Password == "" && creds.Token == "" && creds.Ticket == "" {
			return apperr.New(apperr.CodeAuth, "scope user requires password, token, or ticket")
		}
	case ScopeRootToken:
		if !strings.HasPrefix(strings.ToLower(creds.User), "root@") {
			return apperr.New(apperr.CodeAuth, "scope root-token requires root@ identity")
		}
		if creds.Token == "" {
			return apperr.New(apperr.CodeAuth, "scope root-token requires token")
		}
	case ScopeRoot:
		if !strings.HasPrefix(strings.ToLower(creds.User), "root@") {
			return apperr.New(apperr.CodeAuth, "scope root requires root@ identity")
		}
		if creds.Password == "" && creds.Ticket == "" {
			return apperr.New(apperr.CodeAuth, "scope root requires password or ticket")
		}
	}
	return nil
}

func readAuthFile(authFile string) (Credentials, error) {
	if strings.TrimSpace(authFile) == "" {
		return Credentials{}, nil
	}
	content, err := os.ReadFile(authFile)
	if err != nil {
		return Credentials{}, apperr.Wrap(apperr.CodeConfig, "failed to read auth-file", err)
	}
	var creds Credentials
	if err := json.Unmarshal(content, &creds); err != nil {
		return Credentials{}, apperr.Wrap(apperr.CodeConfig, "failed to parse auth-file JSON", err)
	}
	return creds, nil
}

func mergeCredentials(fileCreds, envCreds, flagCreds Credentials) (Credentials, Sources) {
	merged := Credentials{}
	sources := Sources{}

	merged.User, sources.User = pickValue(fileCreds.User, envCreds.User, flagCreds.User)
	merged.Password, sources.Password = pickValue(fileCreds.Password, envCreds.Password, flagCreds.Password)
	merged.Token, sources.Token = pickValue(fileCreds.Token, envCreds.Token, flagCreds.Token)
	merged.Ticket, sources.Ticket = pickValue(fileCreds.Ticket, envCreds.Ticket, flagCreds.Ticket)
	merged.CSRFToken, sources.CSRFToken = pickValue(fileCreds.CSRFToken, envCreds.CSRFToken, flagCreds.CSRFToken)

	return merged, sources
}

func pickValue(fromFile, fromEnv, fromFlag string) (string, string) {
	if fromFlag != "" {
		return fromFlag, "flag"
	}
	if fromEnv != "" {
		return fromEnv, "env"
	}
	if fromFile != "" {
		return fromFile, "auth-file"
	}
	return "", ""
}

func (c Credentials) SafeSummary() map[string]any {
	return map[string]any{
		"user":              c.User,
		"password_provided": c.Password != "",
		"token_provided":    c.Token != "",
		"ticket_provided":   c.Ticket != "",
		"csrf_provided":     c.CSRFToken != "",
	}
}

func (s Sources) Summary() map[string]string {
	result := map[string]string{}
	if s.User != "" {
		result["user"] = s.User
	}
	if s.Password != "" {
		result["password"] = s.Password
	}
	if s.Token != "" {
		result["token"] = s.Token
	}
	if s.Ticket != "" {
		result["ticket"] = s.Ticket
	}
	if s.CSRFToken != "" {
		result["csrf_token"] = s.CSRFToken
	}
	return result
}

func ExampleAuthFile() string {
	return fmt.Sprintf("{\n  \"user\": \"user@pve\",\n  \"password\": \"...\",\n  \"token\": \"...\"\n}")
}
