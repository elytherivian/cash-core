package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type tokenVerifier struct {
	userID uuid.UUID
	err    error
}

func (v tokenVerifier) Verify(context.Context, string) (uuid.UUID, error) {
	return v.userID, v.err
}

func TestAuthenticationStoresAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	engine := gin.New()
	engine.Use(Authentication(tokenVerifier{userID: userID}, common.NewResponder("test")))
	engine.GET("/accounts", func(c *gin.Context) {
		got, ok := AuthenticatedUserID(c)
		if !ok || got != userID {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAuthenticationRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Authentication(tokenVerifier{}, common.NewResponder("test")))
	engine.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAuthenticationRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Authentication(tokenVerifier{err: common.ErrUnauthenticated}, common.NewResponder("test")))
	engine.GET("/accounts", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
