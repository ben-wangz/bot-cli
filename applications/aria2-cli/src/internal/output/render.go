package output

import (
	"encoding/json"
	"io"

	"github.com/ben-wangz/bot-cli/applications/aria2-cli/src/internal/apperr"
)

const FormatJSON = "json"

func ValidateFormat(format string) error {
	if format == FormatJSON {
		return nil
	}
	return apperr.New(apperr.CodeInvalidArgs, "output must be json")
}

func Render(w io.Writer, format string, payload any) error {
	if err := ValidateFormat(format); err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}
