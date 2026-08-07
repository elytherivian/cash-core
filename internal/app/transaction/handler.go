package transaction

import (
	"fmt"
	"net/http"
	"time"

	"cash-core/internal/common"
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
	userID, err := utils.ParseUUID(c.Param("user_id"))
	if err != nil {
		h.responder.Error(c, err)
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
	h.responder.Success(c, http.StatusCreated, "transaction created", value)
}

func (h *Handler) list(c *gin.Context) {
	userID, err := utils.ParseUUID(c.Param("user_id"))
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	filter, err := parseFilter(c)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	values, total, err := h.service.List(c.Request.Context(), userID, filter)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "ok", common.PageData{
		Items: values, Total: total, Limit: filter.Page.Limit, Offset: filter.Page.Offset,
	})
}

func (h *Handler) get(c *gin.Context) {
	userID, id, ok := h.parseIDs(c)
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
	userID, id, ok := h.parseIDs(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, id); err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "transaction deleted", nil)
}

func (h *Handler) parseIDs(c *gin.Context) (userID, id uuid.UUID, ok bool) {
	userID, err := utils.ParseUUID(c.Param("user_id"))
	if err != nil {
		h.responder.Error(c, err)
		return uuid.Nil, uuid.Nil, false
	}
	id, err = utils.ParseUUID(c.Param("id"))
	if err != nil {
		h.responder.Error(c, err)
		return uuid.Nil, uuid.Nil, false
	}
	return userID, id, true
}

func parseFilter(c *gin.Context) (Filter, error) {
	page, err := utils.ParsePage(c)
	if err != nil {
		return Filter{}, err
	}
	from, err := parseOptionalTime(c.Query("from"))
	if err != nil {
		return Filter{}, err
	}
	to, err := parseOptionalTime(c.Query("to"))
	if err != nil {
		return Filter{}, err
	}
	return Filter{From: from, To: to, Page: page}, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("%w: from and to must use RFC3339", common.ErrInvalidInput)
	}
	parsed = parsed.UTC()
	return &parsed, nil
}
