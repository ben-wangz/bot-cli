package capability

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
)

var allowedGlobalOptionKeys = map[string]bool{
	"bt-detach-seed-only":             true,
	"bt-enable-hook-after-hash-check": true,
	"bt-enable-lpd":                   true,
	"bt-external-ip":                  true,
	"bt-force-encryption":             true,
	"bt-hash-check-seed":              true,
	"bt-load-saved-metadata":          true,
	"bt-max-open-files":               true,
	"bt-max-peers":                    true,
	"bt-metadata-only":                true,
	"bt-min-crypto-level":             true,
	"bt-prioritize-piece":             true,
	"bt-remove-unselected-file":       true,
	"bt-request-peer-speed-limit":     true,
	"bt-require-crypto":               true,
	"bt-save-metadata":                true,
	"bt-seed-unverified":              true,
	"bt-stop-timeout":                 true,
	"bt-tracker":                      true,
	"bt-tracker-connect-timeout":      true,
	"bt-tracker-interval":             true,
	"bt-tracker-timeout":              true,
	"connect-timeout":                 true,
	"continue":                        true,
	"daemon":                          true,
	"deferred-input":                  true,
	"dir":                             true,
	"disable-ipv6":                    true,
	"disk-cache":                      true,
	"download-result":                 true,
	"dry-run":                         true,
	"enable-color":                    true,
	"enable-http-keep-alive":          true,
	"enable-http-pipelining":          true,
	"enable-mmap":                     true,
	"enable-peer-exchange":            true,
	"file-allocation":                 true,
	"follow-metalink":                 true,
	"follow-torrent":                  true,
	"force-save":                      true,
	"ftp-pasv":                        true,
	"hash-check-only":                 true,
	"http-accept-gzip":                true,
	"lowest-speed-limit":              true,
	"max-concurrent-downloads":        true,
	"max-connection-per-server":       true,
	"max-download-limit":              true,
	"max-file-not-found":              true,
	"max-mmap-limit":                  true,
	"max-overall-download-limit":      true,
	"max-overall-upload-limit":        true,
	"max-resume-failure-tries":        true,
	"max-tries":                       true,
	"metalink-enable-unique-protocol": true,
	"min-split-size":                  true,
	"no-file-allocation-limit":        true,
	"optimize-concurrent-downloads":   true,
	"parameterized-uri":               true,
	"pause":                           true,
	"pause-metadata":                  true,
	"piece-length":                    true,
	"proxy-method":                    true,
	"quiet":                           true,
	"realtime-chunk-checksum":         true,
	"remote-time":                     true,
	"remove-control-file":             true,
	"reuse-uri":                       true,
	"rpc-allow-origin-all":            true,
	"rpc-listen-all":                  true,
	"save-not-found":                  true,
	"seed-ratio":                      true,
	"seed-time":                       true,
	"split":                           true,
	"stream-piece-selector":           true,
	"timeout":                         true,
	"truncate-console-readout":        true,
	"uri-selector":                    true,
	"use-head":                        true,
	"user-agent":                      true,
}

func runGetGlobalOption(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	res, err := client.Call(ctx, "aria2.getGlobalOption", []any{})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), nil), nil
}

func runChangeGlobalOption(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	fromOption, err := OptionalKeyValueList(req.Args, "option")
	if err != nil {
		return nil, err
	}
	fromJSON, err := OptionalJSONObject(req.Args, "options")
	if err != nil {
		return nil, err
	}
	if len(fromOption) == 0 && len(fromJSON) == 0 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "missing global options: provide --option key=value or --options '{\"key\":\"value\"}'")
	}
	opts, err := buildGlobalOptionsMap(fromOption, fromJSON)
	if err != nil {
		return nil, err
	}
	if err := validateGlobalOptionKeys(opts); err != nil {
		return nil, err
	}
	res, err := client.Call(ctx, "aria2.changeGlobalOption", []any{opts})
	if err != nil {
		return nil, err
	}
	return envelope(req, json.RawMessage(res), map[string]any{"applied_keys": sortedKeys(opts)}), nil
}

func buildGlobalOptionsMap(entries []string, fromJSON map[string]any) (map[string]string, error) {
	result := map[string]string{}
	for k, v := range fromJSON {
		key := strings.TrimSpace(k)
		if key == "" {
			return nil, apperr.New(apperr.CodeInvalidArgs, "options contains empty key")
		}
		result[key] = strings.TrimSpace(stringValue(v))
	}
	for _, entry := range entries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, apperr.New(apperr.CodeInvalidArgs, "option must be key=value: "+entry)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, apperr.New(apperr.CodeInvalidArgs, "option key must not be empty")
		}
		result[key] = strings.TrimSpace(parts[1])
	}
	return result, nil
}

func validateGlobalOptionKeys(options map[string]string) error {
	unknown := make([]string, 0)
	for k := range options {
		if !allowedGlobalOptionKeys[k] {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return apperr.New(apperr.CodeInvalidArgs, "unsupported global option key(s): "+strings.Join(unknown, ", "))
}

func sortedKeys(options map[string]string) []string {
	keys := make([]string, 0, len(options))
	for k := range options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
