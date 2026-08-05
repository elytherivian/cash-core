package integration

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cash-core/internal/config"
	"cash-core/internal/router"
)

type healthyDatabase struct{}

func (healthyDatabase) Ping(context.Context) error { return nil }

func TestLivenessEndpoint(t *testing.T) {
	cfg := config.Config{
		App:  config.App{Environment: "test", Name: "cash", Version: "test"},
		HTTP: config.HTTP{AllowedOrigins: []string{"*"}},
	}
	engine := router.New(cfg, nil, healthyDatabase{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"version":"test"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestNotFoundUsesCommonResponse(t *testing.T) {
	cfg := config.Config{
		App:  config.App{Environment: "test", Name: "cash", Version: "test"},
		HTTP: config.HTTP{AllowedOrigins: []string{"*"}},
	}
	engine := router.New(cfg, nil, healthyDatabase{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":404`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
