package common

import "errors"

// 通用错误类别 映射为 HTTP 状态码
var (
	ErrNotFound         = errors.New("resource not found")
	ErrConflict         = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthenticated  = errors.New("authentication required")
	ErrForbidden        = errors.New("permission denied")
	ErrUnavailable      = errors.New("service unavailable")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
