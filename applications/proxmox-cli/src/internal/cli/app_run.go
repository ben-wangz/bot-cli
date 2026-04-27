package cli

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/auth"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/output"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/redact"
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
	if tryPrintCommandHelp(tail, stdout) {
		return 0
	}

	if err := output.ValidateFormat(opts.Output); err != nil {
		return printError(err, stderr)
	}

	if canDispatchWithoutAuth(tail) {
		runtime := commandRuntime{
			Opts:   opts,
			Stdout: stdout,
			Stderr: stderr,
		}
		err = dispatchCommand(runtime, tail)
		if err != nil {
			return printError(err, stderr)
		}
		return 0
	}

	creds, sources, err := auth.Load(auth.FlagCredentials{
		User:      opts.AuthUser,
		Password:  opts.AuthPass,
		Token:     opts.AuthToken,
		Ticket:    opts.AuthTicket,
		CSRFToken: opts.AuthCSRF,
	}, opts.AuthFile)
	if err != nil {
		return printError(err, stderr)
	}
	if err := auth.ValidateScope(opts.AuthScope, creds); err != nil {
		return printError(err, stderr)
	}
	if creds.Token == "" && creds.Ticket == "" && creds.User != "" && creds.Password != "" {
		ticket, csrf, loginErr := loginWithPassword(opts.APIBase, opts.InsecureTLS, opts.Timeout, creds.User, creds.Password)
		if loginErr != nil {
			return printError(loginErr, stderr)
		}
		creds.Ticket = ticket
		creds.CSRFToken = csrf
	}

	headers := map[string]string{"Accept": "application/json"}
	if creds.Token != "" {
		headers["Authorization"] = "PVEAPIToken=" + creds.Token
	}
	if creds.Ticket != "" {
		headers["Cookie"] = "PVEAuthCookie=" + creds.Ticket
	}
	if creds.CSRFToken != "" {
		headers["CSRFPreventionToken"] = creds.CSRFToken
	}

	var logger *log.Logger
	if opts.Debug {
		logger = log.New(stderr, "[debug] ", 0)
	}

	client, err := pveapi.New(pveapi.Config{
		BaseURL:     opts.APIBase,
		Timeout:     opts.Timeout,
		InsecureTLS: opts.InsecureTLS,
		Headers:     headers,
		Logger:      logger,
	})
	if err != nil {
		return printError(err, stderr)
	}

	runtime := commandRuntime{
		Opts:    opts,
		Creds:   creds,
		Sources: sources,
		Client:  client,
		Stdout:  stdout,
		Stderr:  stderr,
	}

	err = dispatchCommand(runtime, tail)
	if err != nil {
		return printError(err, stderr)
	}
	return 0
}

func loginWithPassword(apiBase string, insecureTLS bool, timeout time.Duration, user string, password string) (string, string, error) {
	base := strings.TrimRight(strings.TrimSpace(apiBase), "/")
	if base == "" {
		return "", "", apperr.New(apperr.CodeConfig, "api-base is required")
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureTLS}, //nolint:gosec
		},
	}
	form := url.Values{}
	form.Set("username", user)
	form.Set("password", password)
	req, err := http.NewRequest(http.MethodPost, base+"/access/ticket", strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", apperr.Wrap(apperr.CodeNetwork, "failed to build auth request", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", apperr.Wrap(apperr.CodeNetwork, "failed to request access ticket", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return "", "", apperr.New(apperr.CodeAuth, fmt.Sprintf("failed to login with password, status=%d", resp.StatusCode))
	}
	var envelope struct {
		Data struct {
			Ticket string `json:"ticket"`
			CSRF   string `json:"CSRFPreventionToken"`
		} `json:"data"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&envelope); decodeErr != nil {
		return "", "", apperr.Wrap(apperr.CodeNetwork, "failed to decode access ticket response", decodeErr)
	}
	if envelope.Data.Ticket == "" {
		return "", "", apperr.New(apperr.CodeAuth, "access ticket is empty")
	}
	return envelope.Data.Ticket, envelope.Data.CSRF, nil
}

func printError(err error, stderr io.Writer) int {
	if err == nil {
		return 0
	}
	message := redact.String(err.Error())
	if errors.Is(err, os.ErrNotExist) {
		message = "[config_error] file not found"
	}
	_, _ = io.WriteString(stderr, fmt.Sprintf("error: %s\n", message))
	return apperr.ExitCode(err)
}

func canDispatchWithoutAuth(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] != "capability" {
		return false
	}
	if len(args) > 1 && strings.TrimSpace(args[1]) == "describe" {
		return true
	}
	return hasHelp(args[1:])
}
