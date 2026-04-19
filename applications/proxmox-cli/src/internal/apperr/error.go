package apperr

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeInvalidArgs Code = "invalid_args"
	CodeConfig      Code = "config_error"
	CodeAuth        Code = "auth_error"
	CodeNetwork     Code = "network_error"
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

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var appErr *Error
	if !errors.As(err, &appErr) {
		return 1
	}
	switch appErr.Code {
	case CodeInvalidArgs:
		return 2
	case CodeConfig:
		return 3
	case CodeAuth:
		return 4
	case CodeNetwork:
		return 5
	default:
		return 1
	}
}
