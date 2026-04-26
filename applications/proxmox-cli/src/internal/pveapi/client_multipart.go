package pveapi

import (
	"bytes"
	"io"
	"mime/multipart"
	"sort"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

func buildMultipartPrefixSuffix(fields map[string]string, fileField string, filename string) (string, []byte, []byte, error) {
	const placeholder = "__BOT_CLI_MULTIPART_FILE_PLACEHOLDER__"
	buffer := bytes.Buffer{}
	writer := multipart.NewWriter(&buffer)
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if strings.TrimSpace(key) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writer.WriteField(key, fields[key]); err != nil {
			return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to build multipart field", err)
		}
	}
	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to create multipart file part", err)
	}
	if _, err := io.WriteString(part, placeholder); err != nil {
		return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to build multipart placeholder", err)
	}
	if err := writer.Close(); err != nil {
		return "", nil, nil, apperr.Wrap(apperr.CodeNetwork, "failed to finalize multipart payload metadata", err)
	}
	raw := buffer.Bytes()
	marker := []byte(placeholder)
	index := bytes.Index(raw, marker)
	if index < 0 {
		return "", nil, nil, apperr.New(apperr.CodeInternal, "multipart placeholder marker not found")
	}
	prefix := append([]byte(nil), raw[:index]...)
	suffix := append([]byte(nil), raw[index+len(marker):]...)
	return writer.FormDataContentType(), prefix, suffix, nil
}
