package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.App.Name != "cash" || cfg.Database.Name != "cash" || cfg.HTTP.Port != 8080 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadEnvironmentOverridesDefaults(t *testing.T) {
	t.Setenv("APP_NAME", "from-environment")
	t.Setenv("HTTP_PORT", "9100")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "20s")
	t.Setenv("LOG_TIMEZONE", "Asia/Shanghai")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.App.Name != "from-environment" || cfg.HTTP.Port != 9100 || cfg.HTTP.ShutdownTimeout != 20*time.Second || cfg.Log.TimeZone != "Asia/Shanghai" {
		t.Fatalf("unexpected loaded config: %+v", cfg)
	}
}

func TestLoadRejectsInvalidLogTimeZone(t *testing.T) {
	t.Setenv("LOG_TIMEZONE", "UTC+8")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "log timezone must be a valid IANA timezone") {
		t.Fatalf("expected invalid LOG_TIMEZONE error, got %v", err)
	}
}

func TestLoadRejectsMalformedEnvironmentValue(t *testing.T) {
	t.Setenv("HTTP_PORT", "not-a-port")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "HTTP_PORT must be an integer") {
		t.Fatalf("expected malformed HTTP_PORT error, got %v", err)
	}
}
