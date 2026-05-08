package apperr

import "fmt"

type Code string

const (
	CodeInvalidArgs Code = "invalid_args"
	CodeConfig      Code = "config_error"
	CodeNetwork     Code = "network_error"
	CodeRPC         Code = "rpc_error"
	CodeParse       Code = "parse_error"
	CodeInternal    Code = "internal_error"
)

type Error struct {
	Code    Code
	Message string
	Cause   error
	Meta    map[string]any
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
}

func New(code Code, message string) error {
	return &Error{Code: code, Message: message}
}

func NewWithMeta(code Code, message string, meta map[string]any) error {
	return &Error{Code: code, Message: message, Meta: meta}
}

func Wrap(code Code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	typed, ok := err.(*Error)
	if !ok {
		return 1
	}
	switch typed.Code {
	case CodeInvalidArgs:
		return 2
	case CodeConfig:
		return 3
	case CodeNetwork:
		return 4
	case CodeRPC:
		return 5
	case CodeParse:
		return 6
	default:
		return 1
	}
}
