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
	h.responder.Success(c, http.StatusCreated, "user created", value.Response())
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
