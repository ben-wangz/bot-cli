package capability

import (
	"context"
	"sort"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/aria2rpc"
)

type Request struct {
	Name string
	Args map[string]string
}

type Handler func(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error)

type registryEntry struct {
	handler  Handler
	readOnly bool
}

var operationRegistry = map[string]registryEntry{
	"add_uri":               {handler: runAddURI},
	"ensure_daemon_started": {handler: runEnsureDaemonStarted},
	"add_magnet":            {handler: runAddMagnet},
	"add_torrent":           {handler: runAddTorrent},
	"add_metalink":          {handler: runAddMetalink},
	"pause":                 {handler: runPause},
	"pause_all":             {handler: runPauseAll},
	"resume":                {handler: runResume},
	"resume_all":            {handler: runResumeAll},
	"remove":                {handler: runRemove},
	"remove_all":            {handler: runRemoveAll},
	"purge_download_result": {handler: runPurgeDownloadResult},
	"tell_status":           {handler: runTellStatus, readOnly: true},
	"list_active":           {handler: runListActive, readOnly: true},
	"list_waiting":          {handler: runListWaiting, readOnly: true},
	"list_stopped":          {handler: runListStopped, readOnly: true},
	"get_global_stat":       {handler: runGetGlobalStat, readOnly: true},
	"rpc_call":              {handler: runRPCCall},
}

func Dispatch(ctx context.Context, client *aria2rpc.Client, req Request) (map[string]any, error) {
	entry, ok := operationRegistry[req.Name]
	if !ok {
		return nil, apperr.New(apperr.CodeInvalidArgs, "operation not implemented yet: "+req.Name)
	}
	return entry.handler(ctx, client, req)
}

func Names() []string {
	names := make([]string, 0, len(operationRegistry))
	for name := range operationRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func IsReadOnly(name string) bool {
	entry, ok := operationRegistry[name]
	if !ok {
		return false
	}
	return entry.readOnly
}

func Describe(name string) (map[string]any, bool) {
	entry, ok := operationRegistry[name]
	if !ok {
		return nil, false
	}
	return map[string]any{
		"name":      name,
		"read_only": entry.readOnly,
	}, true
}
