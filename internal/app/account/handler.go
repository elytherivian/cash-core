package account

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
	var req CreateAccountRequest
	if err := utils.DecodeJSON(c, &req); err != nil {
		h.responder.Error(c, err)
		return
	}
	createdAccount, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusCreated, "account created", createdAccount)
}

func (h *Handler) listAccounts(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	accounts, err := h.service.ListAccounts(c.Request.Context(), userID)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "accounts listed", accounts)
}

func (h *Handler) authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.AuthenticatedUserID(c)
	if !ok || userID == uuid.Nil {
		h.responder.Error(c, common.ErrUnauthenticated)
		return uuid.Nil, false
	}
	return userID, true
}
