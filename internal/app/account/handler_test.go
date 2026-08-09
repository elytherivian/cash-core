package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlerService struct {
	createdUserID uuid.UUID
	listedUserID  uuid.UUID
}

func (s *handlerService) Create(_ context.Context, userID uuid.UUID, _ CreateAccountRequest) (*Account, error) {
	s.createdUserID = userID
	return &Account{ID: uuid.New(), UserID: userID}, nil
}

func (s *handlerService) ListAccounts(_ context.Context, userID uuid.UUID) ([]Account, error) {
	s.listedUserID = userID
	return []Account{}, nil
}

func TestCreateUsesAuthenticatedUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := new(handlerService)
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	})
	engine.POST("/api/v1/accounts/create", handler.create)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/accounts/create",
		strings.NewReader(`{"account_type":"WeChat","account_name":"wechat1"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || service.createdUserID != userID {
		t.Fatalf("status = %d, service user ID = %s; want %s", response.Code, service.createdUserID, userID)
	}
}

func TestHandlerListAccountsUsesAuthenticatedUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := new(handlerService)
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	})
	engine.GET("/api/v1/accounts/list", handler.listAccounts)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/accounts/list", nil))

	if response.Code != http.StatusOK || service.listedUserID != userID {
		t.Fatalf("status = %d, service user ID = %s; want %s", response.Code, service.listedUserID, userID)
	}
}
