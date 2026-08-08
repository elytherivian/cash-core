package category

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlerService struct {
	listedUserID uuid.UUID
}

func (s *handlerService) Create(context.Context, uuid.UUID, CreateRequest) (*Category, error) {
	return nil, nil
}

func (s *handlerService) Get(context.Context, uuid.UUID, uuid.UUID) (*Category, error) {
	return nil, nil
}

func (s *handlerService) List(_ context.Context, userID uuid.UUID, _ string, _ common.Page) ([]Category, int64, error) {
	s.listedUserID = userID
	return []Category{}, 0, nil
}

func (s *handlerService) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func TestListUsesAuthenticatedUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := new(handlerService)
	handler := NewHandler(service, common.NewResponder("test"))
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	})
	engine.GET("/api/v1/categories", handler.list)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil))

	if response.Code != http.StatusOK || service.listedUserID != userID {
		t.Fatalf("status = %d, service user ID = %s; want %s", response.Code, service.listedUserID, userID)
	}
}
