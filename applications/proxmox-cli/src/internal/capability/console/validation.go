package consolecap

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/pveapi"
)

func RunValidateK1SerialReadable(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	sessionArgs := map[string]string{}
	for key, value := range req.Args {
		sessionArgs[key] = value
	}
	if strings.TrimSpace(sessionArgs["timeout-seconds"]) == "" {
		sessionArgs["timeout-seconds"] = "20"
	}
	sessionData, err := RunSerialWSSessionControl(ctx, client, Request{Name: "serial_ws_session_control", Args: sessionArgs, Scope: req.Scope})
	if err != nil {
		return nil, err
	}
	sessionResult, _ := sessionData["result"].(map[string]any)
	rawTranscript := asString(sessionResult["transcript"])
	cleanTranscript := normalizeSerialText(rawTranscript)
	if strings.TrimSpace(cleanTranscript) == "" {
		return nil, apperr.New(apperr.CodeNetwork, "serial output is empty; cannot validate readability")
	}
	bannerOnly := isOnlySerialStartupBanner(cleanTranscript)
	request := map[string]any{"node": node, "vmid": vmid, "timeout_seconds": sessionArgs["timeout-seconds"], "expect": strings.TrimSpace(sessionArgs["expect"])}
	result := map[string]any{
		"readable":         true,
		"banner_only":      bannerOnly,
		"transcript_clean": cleanTranscript,
		"transcript_tail":  tailText(cleanTranscript, 240),
		"bytes":            len(rawTranscript),
		"websocket":        asString(sessionResult["websocket"]),
	}
	return buildResult(req, request, result, map[string]any{"criterion": "k1_serial_readable", "banner_only": bannerOnly}), nil
}

func RunValidateSerialOutputCriterion2(ctx context.Context, client *pveapi.Client, req Request) (map[string]any, error) {
	node, err := RequiredNode(req.Args)
	if err != nil {
		return nil, err
	}
	vmid, err := RequiredOperationVMID(req.Args)
	if err != nil {
		return nil, err
	}
	captureArgs := map[string]string{}
	for key, value := range req.Args {
		captureArgs[key] = value
	}
	if strings.TrimSpace(captureArgs["log-path"]) == "" {
		captureArgs["log-path"] = filepath.Join("build", fmt.Sprintf("serial-criterion2-%d.log", vmid))
	}
	if strings.TrimSpace(captureArgs["append"]) == "" {
		captureArgs["append"] = "1"
	}
	if strings.TrimSpace(captureArgs["timeout-seconds"]) == "" {
		captureArgs["timeout-seconds"] = "120"
	}
	captureData, err := RunSerialWSCaptureToFile(ctx, client, Request{Name: "serial_ws_capture_to_file", Args: captureArgs, Scope: req.Scope})
	if err != nil {
		return nil, err
	}
	captureResult, _ := captureData["result"].(map[string]any)
	cleanTranscript := normalizeSerialText(asString(captureResult["transcript_clean"]))
	if strings.TrimSpace(cleanTranscript) == "" {
		return nil, apperr.New(apperr.CodeNetwork, "serial output is empty; criterion2 failed")
	}
	if isOnlySerialStartupBanner(cleanTranscript) {
		return nil, apperr.New(apperr.CodeNetwork, "serial output only contains termproxy startup banner; verify kernel cmdline enables serial console")
	}
	request := map[string]any{
		"node":            node,
		"vmid":            vmid,
		"log_path":        captureArgs["log-path"],
		"append":          captureArgs["append"] == "1",
		"timeout_seconds": captureArgs["timeout-seconds"],
		"expect":          strings.TrimSpace(captureArgs["expect"]),
	}
	result := map[string]any{
		"criterion2_passed": true,
		"log_path":          captureResult["log_path"],
		"bytes_written":     captureResult["bytes_written"],
		"matched":           captureResult["matched"],
		"transcript_clean":  cleanTranscript,
		"transcript_tail":   tailText(cleanTranscript, 240),
		"websocket":         captureResult["websocket"],
	}
	return buildResult(req, request, result, map[string]any{"criterion": "serial_output_criterion2"}), nil
}
