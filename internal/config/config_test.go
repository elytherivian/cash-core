package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaultsWhenConfigFileIsMissing(t *testing.T) {
	t.Setenv("CONFIG_FILE", t.TempDir()+"/missing.yaml")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.App.Name != "cash" || cfg.Database.Name != "cash" || cfg.HTTP.Port != 8080 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadReadsYAMLAndEnvironmentTakesPriority(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	content := []byte("app:\n  name: from-yaml\nhttp:\n  port: 9000\n  shutdown_timeout: 20s\nlog:\n  timezone: America/New_York\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_FILE", path)
	t.Setenv("HTTP_PORT", "9100")
	t.Setenv("LOG_TIMEZONE", "Asia/Shanghai")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if cfg.App.Name != "from-yaml" || cfg.HTTP.Port != 9100 || cfg.HTTP.ShutdownTimeout != 20*time.Second || cfg.Log.TimeZone != "Asia/Shanghai" {
		t.Fatalf("unexpected loaded config: %+v", cfg)
	}
}

func TestLoadRejectsInvalidLogTimeZone(t *testing.T) {
	t.Setenv("CONFIG_FILE", t.TempDir()+"/missing.yaml")
	t.Setenv("LOG_TIMEZONE", "UTC+8")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "log timezone must be a valid IANA timezone") {
		t.Fatalf("expected invalid LOG_TIMEZONE error, got %v", err)
	}
}

func TestLoadRejectsMalformedEnvironmentValue(t *testing.T) {
	t.Setenv("CONFIG_FILE", t.TempDir()+"/missing.yaml")
	t.Setenv("HTTP_PORT", "not-a-port")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "HTTP_PORT must be an integer") {
		t.Fatalf("expected malformed HTTP_PORT error, got %v", err)
	}
}
