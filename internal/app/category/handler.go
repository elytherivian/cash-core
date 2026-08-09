package category

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

func (h *Handler) createCategory(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	var request CreateCategoryRequest
	if err := utils.DecodeJSON(c, &request); err != nil {
		h.responder.Error(c, err)
		return
	}
	category, err := h.service.CreateCategory(c.Request.Context(), userID, request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusCreated, "category created", category.Response(h.location))
}

func (h *Handler) listCategories(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	categories, err := h.service.ListCategories(c.Request.Context(), userID)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	responses := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, category.Response(h.location))
	}
	h.responder.Success(c, http.StatusOK, "categories listed", responses)
}

func (h *Handler) authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.AuthenticatedUserID(c)
	if !ok || userID == uuid.Nil {
		h.responder.Error(c, common.ErrUnauthenticated)
		return uuid.Nil, false
	}
	return userID, true
}
