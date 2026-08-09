package category

import (
	"context"
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
	listedUserID uuid.UUID
}

func (s *handlerService) CreateCategory(context.Context, uuid.UUID, CreateCategoryRequest) (*Category, error) {
	return nil, nil
}

func (s *handlerService) ListCategories(_ context.Context, userID uuid.UUID) ([]Category, error) {
	s.listedUserID = userID
	return []Category{}, nil
}

func TestListCategoriesUsesAuthenticatedUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := new(handlerService)
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		c.Next()
	})
	engine.GET("/api/v1/categories/list", handler.listCategories)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/categories/list", nil)
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if service.listedUserID != userID {
		t.Fatalf("service user ID = %s, want %s", service.listedUserID, userID)
	}
}
