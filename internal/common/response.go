package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Version string `json:"version"`
	Code    int    `json:"code"`
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
	c.JSON(status, Response{Version: r.version, Code: status, Message: message, Data: data})
}

func (r Responder) Error(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "internal server error"
	switch {
	case errors.Is(err, ErrInvalidInput):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, ErrUnauthenticated):
		status, message = http.StatusUnauthorized, err.Error()
	case errors.Is(err, ErrForbidden):
		status, message = http.StatusForbidden, err.Error()
	case errors.Is(err, ErrNotFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.Is(err, ErrConflict):
		status, message = http.StatusConflict, err.Error()
	case errors.Is(err, ErrUnavailable):
		status, message = http.StatusServiceUnavailable, err.Error()
	}
	c.AbortWithStatusJSON(status, Response{Version: r.version, Code: status, Message: message})
}
