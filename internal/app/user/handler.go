package user

import (
	"net/http"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	responder common.Responder
	location  *time.Location
}

func NewHandler(service Service, responder common.Responder, location *time.Location) *Handler {
	if location == nil {
		location = time.UTC
	}
	return &Handler{service: service, responder: responder, location: location}
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

	h.responder.Success(
		c, http.StatusCreated,
		"user registered",
		user.Response(h.location),
	)
}

func (h *Handler) delete(c *gin.Context) {
	var req DeleteUserRequest
	if err := utils.DecodeJSON(c, &req); err != nil {
		h.responder.Error(c, err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), req); err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusOK, "user deleted", nil)
}

func (h *Handler) restore(c *gin.Context) {
	var req RestoreUserRequest
	if err := utils.DecodeJSON(c, &req); err != nil {
		h.responder.Error(c, err)
		return
	}

	if err := h.service.Restore(c.Request.Context(), req); err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusOK, "user restored", nil)
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := utils.DecodeJSON(c, &req); err != nil {
		h.responder.Error(c, err)
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "login successful", tokens)
}

func (h *Handler) refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := utils.DecodeJSON(c, &req); err != nil {
		h.responder.Error(c, err)
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), req)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "token refreshed", tokens)
}
