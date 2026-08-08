package category

import (
	"net/http"

	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"
	"cash-core/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service   Service
	responder common.Responder
}

func NewHandler(service Service, responder common.Responder) *Handler {
	return &Handler{service: service, responder: responder}
}

func (h *Handler) create(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	var request CreateRequest
	if err := utils.DecodeJSON(c, &request); err != nil {
		h.responder.Error(c, err)
		return
	}
	value, err := h.service.Create(c.Request.Context(), userID, request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusCreated, "category created", value)
}

func (h *Handler) list(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	page, err := utils.ParsePage(c)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	values, total, err := h.service.List(c.Request.Context(), userID, c.Query("type"), page)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "ok", common.PageData{Items: values, Total: total, Limit: page.Limit, Offset: page.Offset})
}

func (h *Handler) get(c *gin.Context) {
	userID, id, ok := h.parseID(c)
	if !ok {
		return
	}
	value, err := h.service.Get(c.Request.Context(), userID, id)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "ok", value)
}

func (h *Handler) delete(c *gin.Context) {
	userID, id, ok := h.parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, id); err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "category deleted", nil)
}

func (h *Handler) authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.AuthenticatedUserID(c)
	if !ok || userID == uuid.Nil {
		h.responder.Error(c, common.ErrUnauthenticated)
		return uuid.Nil, false
	}
	return userID, true
}

func (h *Handler) parseID(c *gin.Context) (userID, id uuid.UUID, ok bool) {
	userID, ok = h.authenticatedUserID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}
	id, err := utils.ParseUUID(c.Param("id"))
	if err != nil {
		h.responder.Error(c, err)
		return uuid.Nil, uuid.Nil, false
	}
	return userID, id, true
}
