package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxRequestBodyBytes = 1 << 20

func DecodeJSON(c *gin.Context, target any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: decode request body: %v", common.ErrInvalidInput, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: request body must contain exactly one JSON object", common.ErrInvalidInput)
	}
	return nil
}

func ParseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid UUID", common.ErrInvalidInput)
	}
	return id, nil
}

func ParsePage(c *gin.Context) (common.Page, error) {
	limit, err := parseNonNegativeInt(c.DefaultQuery("limit", "20"))
	if err != nil || limit == 0 || limit > 100 {
		return common.Page{}, fmt.Errorf("%w: limit must be between 1 and 100", common.ErrInvalidInput)
	}
	offset, err := parseNonNegativeInt(c.DefaultQuery("offset", "0"))
	if err != nil {
		return common.Page{}, fmt.Errorf("%w: offset must be a non-negative integer", common.ErrInvalidInput)
	}
	return common.Page{Limit: limit, Offset: offset}, nil
}

func parseNonNegativeInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid non-negative integer")
	}
	return parsed, nil
}
