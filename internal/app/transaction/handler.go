package transaction

import (
	"fmt"
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

func (h *Handler) createTransaction(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	var request CreateTransactionRequest
	if err := utils.DecodeJSON(c, &request); err != nil {
		h.responder.Error(c, err)
		return
	}
	transaction, err := h.service.CreateTransaction(c.Request.Context(), userID, request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusCreated, "transaction created", transaction.Response(h.location))
}

func (h *Handler) updateTransaction(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	var request UpdateTransactionRequest
	if err := utils.DecodeJSON(c, &request); err != nil {
		h.responder.Error(c, err)
		return
	}
	transactionID, err := utils.ParseUUID(request.ID)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	updatedTransaction, err := h.service.UpdateTransaction(c.Request.Context(), userID, transactionID, request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	h.responder.Success(c, http.StatusOK, "transaction updated", updatedTransaction.Response(h.location))
}

func (h *Handler) getByAccount(c *gin.Context) {
	h.listTransactions(c, true, false)
}

func (h *Handler) getByCategory(c *gin.Context) {
	h.listTransactions(c, false, true)
}

func (h *Handler) getByAccountAndCategory(c *gin.Context) {
	h.listTransactions(c, true, true)
}

func (h *Handler) getByDayTimestamp(c *gin.Context) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	request, err := parseTimeRangeTransactionsRequest(c)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	transactions, err := h.service.ListTransactionsByTimeRange(c.Request.Context(), userID, request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	responses := make([]TransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, transaction.Response(h.location))
	}
	h.responder.Success(c, http.StatusOK, "transactions listed", responses)
}

func (h *Handler) listTransactions(c *gin.Context, requireAccountID, requireCategoryID bool) {
	userID, ok := h.authenticatedUserID(c)
	if !ok {
		return
	}
	request, err := parseListTransactionsRequest(c, requireAccountID, requireCategoryID)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	transactions, err := h.service.ListTransactions(c.Request.Context(), userID, request)
	if err != nil {
		h.responder.Error(c, err)
		return
	}
	responses := make([]TransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, transaction.Response(h.location))
	}
	h.responder.Success(c, http.StatusOK, "transactions listed", responses)
}

func parseListTransactionsRequest(c *gin.Context, requireAccountID, requireCategoryID bool) (ListTransactionsRequest, error) {
	request := ListTransactionsRequest{}
	if requireAccountID {
		accountID, err := parseRequiredQueryUUID(c, "account_id")
		if err != nil {
			return ListTransactionsRequest{}, err
		}
		request.AccountID = &accountID
	}
	if requireCategoryID {
		categoryID, err := parseRequiredQueryUUID(c, "category_id")
		if err != nil {
			return ListTransactionsRequest{}, err
		}
		request.CategoryID = &categoryID
	}
	return request, nil
}

func parseRequiredQueryUUID(c *gin.Context, queryName string) (uuid.UUID, error) {
	value := c.Query(queryName)
	if value == "" {
		return uuid.Nil, fmt.Errorf("%w: %s is required", common.ErrInvalidInput, queryName)
	}
	return utils.ParseUUID(value)
}

func parseTimeRangeTransactionsRequest(c *gin.Context) (ListTransactionsByTimeRangeRequest, error) {
	startTimestamp, err := parseRequiredUTCTimestamp(c, "start_timestamp")
	if err != nil {
		return ListTransactionsByTimeRangeRequest{}, err
	}
	endTimestamp, err := parseRequiredUTCTimestamp(c, "end_timestamp")
	if err != nil {
		return ListTransactionsByTimeRangeRequest{}, err
	}
	if endTimestamp.Before(startTimestamp) {
		return ListTransactionsByTimeRangeRequest{}, fmt.Errorf("%w: end_timestamp must not be before start_timestamp", common.ErrInvalidInput)
	}
	return ListTransactionsByTimeRangeRequest{
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}, nil
}

func parseRequiredUTCTimestamp(c *gin.Context, queryName string) (time.Time, error) {
	value := c.Query(queryName)
	if value == "" {
		return time.Time{}, fmt.Errorf("%w: %s is required", common.ErrInvalidInput, queryName)
	}
	timestamp, err := time.Parse(time.RFC3339, value)
	if err != nil || timestamp.IsZero() {
		return time.Time{}, fmt.Errorf("%w: %s must be an RFC3339 UTC timestamp", common.ErrInvalidInput, queryName)
	}
	return timestamp.UTC(), nil
}

func (h *Handler) authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.AuthenticatedUserID(c)
	if !ok || userID == uuid.Nil {
		h.responder.Error(c, common.ErrUnauthenticated)
		return uuid.Nil, false
	}
	return userID, true
}
