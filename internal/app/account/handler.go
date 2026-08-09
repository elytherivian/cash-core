package account

import (
	"net/http"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"
	"cash-core/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	h.responder.Success(c, http.StatusCreated, "account created", createdAccount.Response(h.location))
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
	responses := make([]AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		responses = append(responses, account.Response(h.location))
	}
	h.responder.Success(c, http.StatusOK, "accounts listed", responses)
}

func (h *Handler) authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.AuthenticatedUserID(c)
	if !ok || userID == uuid.Nil {
		h.responder.Error(c, common.ErrUnauthenticated)
		return uuid.Nil, false
	}
	return userID, true
}
