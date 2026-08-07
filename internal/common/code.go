package common

import (
	"errors"
	"fmt"
)

type Code int

// 业务状态码 code
const (
	CodeSuccess Code = 0

	CodeInvalidInput              Code = 40000
	CodeRegisterUserAlreadyExists Code = 40001
	CodeUnauthenticated           Code = 40100
	CodeForbidden                 Code = 40300
	CodeResourceNotFound          Code = 40400
	CodeMethodNotAllowed          Code = 40500
	CodeResourceConflict          Code = 40900
	CodeInternalServerError       Code = 50000
	CodeServiceUnavailable        Code = 50300
)

// BusinessError adds an application code and a client-facing message while
// preserving the underlying error for errors.Is/errors.As checks.
type BusinessError struct {
	code    Code
	message string
	cause   error
}

func NewBusinessError(code Code, message string, cause error) error {
	return &BusinessError{code: code, message: message, cause: cause}
}

// Error 返回通用错误类别和具体业务信息。
func (e *BusinessError) Error() string {
	switch {
	case e.cause == nil:
		return e.message
	case e.message == "":
		return e.cause.Error()
	default:
		return fmt.Sprintf("%s: %s", e.cause, e.message)
	}
}

func (e *BusinessError) Unwrap() error { return e.cause }

func BusinessCode(err error) (Code, bool) {
	var businessError *BusinessError
	if !errors.As(err, &businessError) {
		return 0, false
	}
	return businessError.code, true
}
