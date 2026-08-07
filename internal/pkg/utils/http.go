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

// maxRequestBodyBytes 限制单个 JSON 请求体最大为 1 MiB，防止请求占用过多内存。
const maxRequestBodyBytes = 1 << 20

// DecodeJSON 将 HTTP 请求体解码到 target。
// 请求体必须是不超过 1 MiB 的单个 JSON 对象，且不能包含 target 中未定义的字段；
// 解码失败时返回包装了 common.ErrInvalidInput 的错误。
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

// ParseUUID 将字符串解析为 UUID。
// 输入格式无效时返回 uuid.Nil 和包装了 common.ErrInvalidInput 的错误。
func ParseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid UUID", common.ErrInvalidInput)
	}
	return id, nil
}

// ParsePage 从查询参数中解析分页信息。
// limit 默认为 20，取值范围为 1 到 100；offset 默认为 0，且必须是非负整数。
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

// parseNonNegativeInt 将字符串解析为非负整数。
// 字符串不是整数或解析结果小于 0 时返回错误。
func parseNonNegativeInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid non-negative integer")
	}
	return parsed, nil
}
