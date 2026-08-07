package common

import "errors"

var (
	ErrNotFound         = errors.New("resource not found")
	ErrConflict         = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthenticated  = errors.New("authentication required")
	ErrForbidden        = errors.New("permission denied")
	ErrUnavailable      = errors.New("service unavailable")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
