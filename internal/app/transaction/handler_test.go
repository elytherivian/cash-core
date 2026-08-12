package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlerService struct {
	updatedUserID      uuid.UUID
	updatedID          uuid.UUID
	updateRequest      UpdateTransactionRequest
	updatedTransaction *Transaction
	listedUserID       uuid.UUID
	listRequest        ListTransactionsRequest
}

func (s *handlerService) CreateTransaction(context.Context, uuid.UUID, CreateTransactionRequest) (*Transaction, error) {
	return nil, nil
}

func (s *handlerService) UpdateTransaction(
	_ context.Context,
	userID, transactionID uuid.UUID,
	request UpdateTransactionRequest,
) (*Transaction, error) {
	s.updatedUserID = userID
	s.updatedID = transactionID
	s.updateRequest = request
	s.updatedTransaction = &Transaction{ID: transactionID, UserID: userID, Type: Expense}
	return s.updatedTransaction, nil
}

func (s *handlerService) ListTransactions(_ context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error) {
	s.listedUserID = userID
	s.listRequest = request
	return []Transaction{}, nil
}

func TestTransactionQueryRoutesUseAuthenticatedUserAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID, accountID, categoryID := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name           string
		path           string
		wantAccountID  bool
		wantCategoryID bool
	}{
		{
			name: "account", path: "/api/v1/transactions/getByAccountID?account_id=" + accountID.String(),
			wantAccountID: true,
		},
		{
			name: "category", path: "/api/v1/transactions/getByCategoryID?category_id=" + categoryID.String(),
			wantCategoryID: true,
		},
		{
			name: "account and category", path: "/api/v1/transactions/getByAccountIDAndCategoryID?account_id=" + accountID.String() + "&category_id=" + categoryID.String(),
			wantAccountID: true, wantCategoryID: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := new(handlerService)
			handler := NewHandler(service, common.NewResponder("test"), time.UTC)
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Set(middleware.UserIDKey, userID)
				c.Next()
			})
			switch test.name {
			case "account":
				engine.GET("/api/v1/transactions/getByAccountID", handler.getByAccount)
			case "category":
				engine.GET("/api/v1/transactions/getByCategoryID", handler.getByCategory)
			case "account and category":
				engine.GET("/api/v1/transactions/getByAccountIDAndCategoryID", handler.getByAccountAndCategory)
			}

			response := httptest.NewRecorder()
			engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if service.listedUserID != userID || (test.wantAccountID && (service.listRequest.AccountID == nil || *service.listRequest.AccountID != accountID)) ||
				(test.wantCategoryID && (service.listRequest.CategoryID == nil || *service.listRequest.CategoryID != categoryID)) {
				t.Fatalf("user ID = %s, request = %+v", service.listedUserID, service.listRequest)
			}
		})
	}
}

func TestUpdateTransactionUsesBodyIDAndAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID, transactionID := uuid.New(), uuid.New()
	service := new(handlerService)
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	})
	engine.PATCH("/api/v1/transactions/update", handler.updateTransaction)

	body, err := json.Marshal(map[string]any{"id": transactionID.String(), "type": "expense"})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/transactions/update", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if service.updatedUserID != userID || service.updatedID != transactionID ||
		service.updateRequest.Type == nil || *service.updateRequest.Type != Expense {
		t.Fatalf("user ID = %s, transaction ID = %s, request = %+v", service.updatedUserID, service.updatedID, service.updateRequest)
	}
}
