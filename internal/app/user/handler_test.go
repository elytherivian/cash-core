package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

type deleteService struct {
	deleteRequest  DeleteUserRequest
	deleteErr      error
	deleteCalled   bool
	restoreRequest RestoreUserRequest
	restoreErr     error
	restoreCalled  bool
	loginRequest   LoginRequest
	loginTokens    auth.TokenPair
	loginErr       error
	refreshRequest RefreshTokenRequest
	refreshTokens  auth.TokenPair
	refreshErr     error
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

func (s *deleteService) Login(_ context.Context, request LoginRequest) (auth.TokenPair, error) {
	s.loginRequest = request
	return s.loginTokens, s.loginErr
}

func (s *deleteService) Refresh(_ context.Context, request RefreshTokenRequest) (auth.TokenPair, error) {
	s.refreshRequest = request
	return s.refreshTokens, s.refreshErr
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

func TestLoginHandlerReturnsTokenPair(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deleteService{loginTokens: auth.TokenPair{
		AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", ExpiresIn: 900,
	}}
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.POST("/api/v1/auth/login", handler.login)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"username":"user","password":"password123"}`),
	)

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"access_token":"access"`) {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.loginRequest.Username != "user" || service.loginRequest.Password != "password123" {
		t.Fatalf("login request = %+v", service.loginRequest)
	}
}

func TestRefreshHandlerReturnsNewTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deleteService{refreshTokens: auth.TokenPair{
		AccessToken: "new-access", RefreshToken: "new-refresh", TokenType: "Bearer", ExpiresIn: 900,
	}}
	handler := NewHandler(service, common.NewResponder("test"), time.UTC)
	engine := gin.New()
	engine.POST("/api/v1/auth/refresh", handler.refresh)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh",
		strings.NewReader(`{"refresh_token":"refresh"}`),
	)

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"access_token":"new-access"`) {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.refreshRequest.RefreshToken != "refresh" {
		t.Fatalf("refresh request = %+v", service.refreshRequest)
	}
}
