package apperr

import "fmt"

type Code string

const (
	CodeInvalidArgs Code = "invalid_args"
	CodeConfig      Code = "config_error"
	CodeNetwork     Code = "network_error"
	CodeRPC         Code = "rpc_error"
	CodeInternal    Code = "internal_error"
)

type Error struct {
	Code    Code
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func New(code Code, message string) error {
	return &Error{Code: code, Message: message}
}

func Wrap(code Code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if typed, ok := err.(*Error); ok {
		switch typed.Code {
		case CodeInvalidArgs:
			return 2
		case CodeConfig:
			return 3
		case CodeNetwork:
			return 4
		case CodeRPC:
			return 5
		default:
			return 1
		}
	}
	return 1
}
