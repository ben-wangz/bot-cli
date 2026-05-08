package capability

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/tts-cli/src/internal/apperr"
)

func runFileToDataURI(req Request) (map[string]any, error) {
	path, err := RequiredString(req.Args, "file_path")
	if err != nil {
		return nil, err
	}
	includeDataURI := OptionalBool(req.Args, "include_data_uri", false)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInvalidArgs, "failed to read input file", err)
	}
	if len(b) == 0 {
		return nil, apperr.New(apperr.CodeInvalidArgs, "input file is empty")
	}
	mimeType := detectAudioMIMEFromContent(b)
	if mimeType == "" {
		return nil, apperr.New(apperr.CodeInvalidArgs, "unsupported or unknown audio format")
	}
	encoded := base64.StdEncoding.EncodeToString(b)
	dataURI := "data:" + mimeType + ";base64," + encoded

	warnings := make([]string, 0)
	if len(encoded) > 10*1024*1024 {
		warnings = append(warnings, "base64 payload exceeds 10MB; upstream may reject this voice clone sample")
	}

	result := map[string]any{
		"file_path":         path,
		"mime_type":         mimeType,
		"base64_size_bytes": len(encoded),
		"data_uri_omitted":  !includeDataURI,
	}
	if includeDataURI {
		result["data_uri"] = dataURI
	} else {
		result["data_uri_preview"] = previewDataURI(dataURI)
	}

	return map[string]any{
		"ok": true,
		"request": map[string]any{
			"capability": "file_to_data_uri",
			"args":       req.Args,
		},
		"result": result,
		"diagnostics": map[string]any{
			"warnings": warnings,
		},
	}, nil
}

func previewDataURI(dataURI string) string {
	v := strings.TrimSpace(dataURI)
	if len(v) <= 120 {
		return v
	}
	return v[:120] + "..."
}

func detectAudioMIMEFromContent(b []byte) string {
	probeLen := len(b)
	if probeLen > 512 {
		probeLen = 512
	}
	detected := strings.ToLower(strings.TrimSpace(http.DetectContentType(b[:probeLen])))
	switch detected {
	case "audio/mpeg", "audio/wav", "audio/x-wav", "audio/ogg", "audio/flac", "audio/mp4":
		if detected == "audio/x-wav" {
			return "audio/wav"
		}
		return detected
	}

	if len(b) >= 3 && string(b[:3]) == "ID3" {
		return "audio/mpeg"
	}
	if len(b) >= 2 && b[0] == 0xFF && (b[1]&0xE0) == 0xE0 {
		return "audio/mpeg"
	}
	if len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WAVE" {
		return "audio/wav"
	}
	if len(b) >= 4 && string(b[:4]) == "OggS" {
		return "audio/ogg"
	}
	if len(b) >= 4 && string(b[:4]) == "fLaC" {
		return "audio/flac"
	}
	if len(b) >= 12 && string(b[4:8]) == "ftyp" {
		brand := string(b[8:12])
		switch brand {
		case "M4A ", "M4B ", "isom", "iso2", "mp41", "mp42":
			return "audio/mp4"
		default:
			if looksLikePrintable(brand) {
				return "audio/mp4"
			}
		}
	}

	return ""
}

func looksLikePrintable(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
	}
	return true
}
