package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Version string `json:"version"`
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Responder struct {
	version string
}

func NewResponder(version string) Responder {
	return Responder{version: version}
}

func (r Responder) Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Response{Version: r.version, Code: CodeSuccess, Message: message, Data: data})
}

func (r Responder) Error(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, CodeInternalServerError, "internal server error"
	switch {
	case errors.Is(err, ErrInvalidInput):
		status, code, message = http.StatusBadRequest, CodeInvalidInput, err.Error()
	case errors.Is(err, ErrUnauthenticated):
		status, code, message = http.StatusUnauthorized, CodeUnauthenticated, err.Error()
	case errors.Is(err, ErrForbidden):
		status, code, message = http.StatusForbidden, CodeForbidden, err.Error()
	case errors.Is(err, ErrNotFound):
		status, code, message = http.StatusNotFound, CodeResourceNotFound, err.Error()
	case errors.Is(err, ErrMethodNotAllowed):
		status, code, message = http.StatusMethodNotAllowed, CodeMethodNotAllowed, err.Error()
	case errors.Is(err, ErrConflict):
		status, code, message = http.StatusConflict, CodeResourceConflict, err.Error()
	case errors.Is(err, ErrUnavailable):
		status, code, message = http.StatusServiceUnavailable, CodeServiceUnavailable, err.Error()
	}
	if businessCode, ok := BusinessCode(err); ok {
		code = businessCode
	}
	c.AbortWithStatusJSON(status, Response{Version: r.version, Code: code, Message: message})
}
