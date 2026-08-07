package user

import (
	"net/http"

	"cash-core/internal/common"
	"cash-core/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	responder common.Responder
}

func NewHandler(service Service, responder common.Responder) *Handler {
	return &Handler{service: service, responder: responder}
}

func (h *Handler) create(c *gin.Context) {
	var request CreateRequest
	if err := utils.DecodeJSON(c, &request); err != nil {
		h.responder.Error(c, err)
		return
	}
	value, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	// 写入返回
	h.responder.Success(c, http.StatusCreated, "user created", value.Response())
}

// register 注册用户
func (h *Handler) register(c *gin.Context) {
	// handler 解析请求，调用 service 返回响应
	var req RegisterUserRequest
	if err := utils.DecodeJSON(c, &req); err != nil {
		// 这里的 error 已经被包装为 common.ErrInvalidInput
		h.responder.Error(c, err)
		return
	}
	// 调用 service 层的 Register 方法
	user, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusCreated, "user registered", user.Response())
}

func (h *Handler) get(c *gin.Context) {
	id, err := utils.ParseUUID(c.Param("user_id"))
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	value, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "ok", value.Response())
}

func (h *Handler) delete(c *gin.Context) {
	id, err := utils.ParseUUID(c.Param("user_id"))
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "user deleted", nil)
}
