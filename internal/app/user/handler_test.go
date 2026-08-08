package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
)

type deleteService struct {
	deleteRequest  DeleteUserRequest
	deleteErr      error
	deleteCalled   bool
	restoreRequest RestoreUserRequest
	restoreErr     error
	restoreCalled  bool
}

func (s *deleteService) Register(context.Context, RegisterUserRequest) (*User, error) {
	return nil, nil
}

func (s *deleteService) Delete(_ context.Context, request DeleteUserRequest) error {
	s.deleteCalled = true
	s.deleteRequest = request
	return s.deleteErr
}

func (s *deleteService) Restore(_ context.Context, request RestoreUserRequest) error {
	s.restoreCalled = true
	s.restoreRequest = request
	return s.restoreErr
}

func TestDeleteHandlerCallsService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deleteService{}
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.DELETE("/api/v1/users/delete", handler.delete)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/users/delete",
		strings.NewReader(`{"username":"user","password":"password123"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s; want %d", response.Code, response.Body.String(), http.StatusOK)
	}
	if !service.deleteCalled || service.deleteRequest.Username != "user" || service.deleteRequest.Password != "password123" {
		t.Fatalf("service request = %+v, called = %t", service.deleteRequest, service.deleteCalled)
	}
}

func TestDeleteHandlerReturnsServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deleteService{deleteErr: common.ErrUnauthenticated}
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.DELETE("/api/v1/users/delete", handler.delete)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/users/delete",
		strings.NewReader(`{"username":"user","password":"incorrect"}`),
	)

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s; want %d", response.Code, response.Body.String(), http.StatusUnauthorized)
	}
}

func TestRestoreHandlerCallsService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deleteService{}
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.POST("/api/v1/users/restore", handler.restore)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/users/restore",
		strings.NewReader(`{"username":"user","password":"password123"}`),
	)

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s; want %d", response.Code, response.Body.String(), http.StatusOK)
	}
	if !service.restoreCalled || service.restoreRequest.Username != "user" || service.restoreRequest.Password != "password123" {
		t.Fatalf("service request = %+v, called = %t", service.restoreRequest, service.restoreCalled)
	}
}

func TestRestoreHandlerReturnsConflictForActiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deleteService{restoreErr: common.ErrConflict}
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.POST("/api/v1/users/restore", handler.restore)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/users/restore",
		strings.NewReader(`{"username":"active-user","password":"password123"}`),
	)

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s; want %d", response.Code, response.Body.String(), http.StatusConflict)
	}
}
