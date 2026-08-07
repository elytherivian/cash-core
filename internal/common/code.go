package common

import "errors"

// Code is an application-level response code. It is independent of HTTP status codes.
type Code int

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

func (e *BusinessError) Error() string { return e.message }

func (e *BusinessError) Unwrap() error { return e.cause }

func BusinessCode(err error) (Code, bool) {
	var businessError *BusinessError
	if !errors.As(err, &businessError) {
		return 0, false
	}
	return businessError.code, true
}
