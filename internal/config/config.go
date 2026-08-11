package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata"
)

type Config struct {
	App      App
	API      API
	Auth     Auth
	HTTP     HTTP
	Database Database
	Log      Log
}

type App struct {
	Environment string
	Name        string
	Version     string
}

// API 包含影响 HTTP 响应内容的配置。
type API struct {
	TimeZone string
}

// Location 返回 API 响应使用的时区。配置在启动前已经校验；这里保留 UTC
// 兜底，便于直接构造 Config 的测试和调用方安全使用。
func (c API) Location() *time.Location {
	location, err := time.LoadLocation(c.TimeZone)
	if err != nil {
		return time.UTC
	}
	return location
}

const defaultJWTSecret = "local-development-only-jwt-secret-change-me"

type Auth struct {
	Issuer          string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type HTTP struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	AllowedOrigins    []string
}

func (c HTTP) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

type Database struct {
	Path string
}

func (c Database) DSN() string {
	return "file:" + c.Path + "?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000"
}

// Directory returns the containing directory for a file-backed SQLite database.
// SQLite's in-memory databases do not need a directory.
func (c Database) Directory() string {
	if c.Path == ":memory:" {
		return ""
	}
	return filepath.Dir(c.Path)
}

type Log struct {
	Level    string
	Format   string
	TimeZone string
}

func Load() (Config, error) {
	// 获取默认 Config 对象
	cfg := defaults()
	// 读取 env 环境变量 并校验
	var loadErrors []error
	applyEnvironment(&cfg, &loadErrors)
	loadErrors = append(loadErrors, cfg.Validate())
	// errors.Join 使用换行符拼接 errors
	if err := errors.Join(loadErrors...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// defaults 返回默认 Config 对象
func defaults() Config {
	return Config{
		App: App{Environment: "local", Name: "cash", Version: "dev"},
		API: API{TimeZone: "Asia/Shanghai"},
		Auth: Auth{
			Issuer: "cash-core", JWTSecret: defaultJWTSecret,
			AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 30 * 24 * time.Hour,
		},
		HTTP: HTTP{
			Host: "0.0.0.0", Port: 8080, ReadTimeout: 10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 15 * time.Second,
			IdleTimeout: 60 * time.Second, ShutdownTimeout: 10 * time.Second,
			AllowedOrigins: []string{"*"},
		},
		Database: Database{
			Path: "data/cash.db",
		},
		Log: Log{Level: "info", Format: "json", TimeZone: "Asia/Shanghai"},
	}
}

func applyEnvironment(cfg *Config, loadErrors *[]error) {
	cfg.App.Environment = env("APP_ENV", cfg.App.Environment)
	cfg.App.Name = env("APP_NAME", cfg.App.Name)
	cfg.App.Version = env("APP_VERSION", cfg.App.Version)
	cfg.API.TimeZone = env("API_TIMEZONE", cfg.API.TimeZone)
	cfg.Auth.Issuer = env("JWT_ISSUER", cfg.Auth.Issuer)
	cfg.Auth.JWTSecret = env("JWT_SECRET", cfg.Auth.JWTSecret)
	cfg.Auth.AccessTokenTTL = envDuration(loadErrors, "JWT_ACCESS_TOKEN_TTL", cfg.Auth.AccessTokenTTL)
	cfg.Auth.RefreshTokenTTL = envDuration(loadErrors, "JWT_REFRESH_TOKEN_TTL", cfg.Auth.RefreshTokenTTL)
	cfg.HTTP.Host = env("HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt(loadErrors, "HTTP_PORT", cfg.HTTP.Port)
	cfg.HTTP.ReadTimeout = envDuration(loadErrors, "HTTP_READ_TIMEOUT", cfg.HTTP.ReadTimeout)
	cfg.HTTP.ReadHeaderTimeout = envDuration(loadErrors, "HTTP_READ_HEADER_TIMEOUT", cfg.HTTP.ReadHeaderTimeout)
	cfg.HTTP.WriteTimeout = envDuration(loadErrors, "HTTP_WRITE_TIMEOUT", cfg.HTTP.WriteTimeout)
	cfg.HTTP.IdleTimeout = envDuration(loadErrors, "HTTP_IDLE_TIMEOUT", cfg.HTTP.IdleTimeout)
	cfg.HTTP.ShutdownTimeout = envDuration(loadErrors, "HTTP_SHUTDOWN_TIMEOUT", cfg.HTTP.ShutdownTimeout)
	cfg.HTTP.AllowedOrigins = envCSV("HTTP_ALLOWED_ORIGINS", cfg.HTTP.AllowedOrigins)
	cfg.Database.Path = env("DB_PATH", cfg.Database.Path)
	cfg.Log.Level = env("LOG_LEVEL", cfg.Log.Level)
	cfg.Log.Format = env("LOG_FORMAT", cfg.Log.Format)
	cfg.Log.TimeZone = env("LOG_TIMEZONE", cfg.Log.TimeZone)
}

func (c Config) Validate() error {
	var validationErrors []error
	if c.App.Name == "" {
		validationErrors = append(validationErrors, errors.New("app name must not be empty"))
	}
	if _, err := time.LoadLocation(c.API.TimeZone); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("API timezone must be a valid IANA timezone: %w", err))
	}
	if strings.TrimSpace(c.Auth.Issuer) == "" {
		validationErrors = append(validationErrors, errors.New("JWT issuer must not be empty"))
	}
	if len(c.Auth.JWTSecret) < 32 {
		validationErrors = append(validationErrors, errors.New("JWT secret must be at least 32 bytes"))
	}
	if (c.App.Environment == "production" || c.App.Environment == "staging") && c.Auth.JWTSecret == defaultJWTSecret {
		validationErrors = append(validationErrors, errors.New("JWT secret must be changed outside local development"))
	}
	if c.Auth.AccessTokenTTL <= 0 {
		validationErrors = append(validationErrors, errors.New("JWT access token TTL must be positive"))
	}
	if c.Auth.RefreshTokenTTL <= c.Auth.AccessTokenTTL {
		validationErrors = append(validationErrors, errors.New("JWT refresh token TTL must be greater than access token TTL"))
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		validationErrors = append(validationErrors, errors.New("HTTP port must be between 1 and 65535"))
	}
	if strings.TrimSpace(c.Database.Path) == "" {
		validationErrors = append(validationErrors, errors.New("database path must not be empty"))
	}
	if c.Log.Format != "json" && c.Log.Format != "text" {
		validationErrors = append(validationErrors, errors.New("log format must be json or text"))
	}
	if _, err := time.LoadLocation(c.Log.TimeZone); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("log timezone must be a valid IANA timezone: %w", err))
	}
	return errors.Join(validationErrors...)
}

// env 从系统环境变量中查找 key value
// 没有则返回默认值 fallback
func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envInt(loadErrors *[]error, key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		*loadErrors = append(*loadErrors, fmt.Errorf("%s must be an integer: %w", key, err))
		return fallback
	}
	return value
}

func envDuration(loadErrors *[]error, key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		*loadErrors = append(*loadErrors, fmt.Errorf("%s must be a duration: %w", key, err))
		return fallback
	}
	return value
}

func envCSV(key string, fallback []string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	values := make([]string, 0)
	for _, part := range strings.Split(raw, ",") {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}
