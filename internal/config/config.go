package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      App      `yaml:"app"`
	HTTP     HTTP     `yaml:"http"`
	Database Database `yaml:"database"`
	Log      Log      `yaml:"log"`
}

type App struct {
	Environment string `yaml:"environment"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
}

type HTTP struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
	AllowedOrigins    []string      `yaml:"allowed_origins"`
}

func (c HTTP) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

type Database struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout"`
}

func (c Database) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.Name,
	}
	query := u.Query()
	query.Set("sslmode", c.SSLMode)
	query.Set("connect_timeout", strconv.Itoa(max(1, int(c.ConnectTimeout.Seconds()))))
	u.RawQuery = query.Encode()
	return u.String()
}

type Log struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	TimeZone string `yaml:"timezone"`
}

func Load() (Config, error) {
	cfg := defaults()
	path := env("CONFIG_FILE", "configs/config.yaml")
	if err := loadYAML(path, &cfg); err != nil {
		return Config{}, err
	}

	var loadErrors []error
	applyEnvironment(&cfg, &loadErrors)
	loadErrors = append(loadErrors, cfg.Validate())
	if err := errors.Join(loadErrors...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaults() Config {
	return Config{
		App: App{Environment: "local", Name: "cash", Version: "dev"},
		HTTP: HTTP{
			Host: "0.0.0.0", Port: 8080, ReadTimeout: 10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 15 * time.Second,
			IdleTimeout: 60 * time.Second, ShutdownTimeout: 10 * time.Second,
			AllowedOrigins: []string{"*"},
		},
		Database: Database{
			Host: "localhost", Port: 5432, Name: "cash", User: "cash", Password: "cash",
			SSLMode: "disable", MaxOpenConns: 20, MaxIdleConns: 5,
			ConnMaxLifetime: time.Hour, ConnMaxIdleTime: 30 * time.Minute, ConnectTimeout: 5 * time.Second,
		},
		Log: Log{Level: "info", Format: "json", TimeZone: "Asia/Shanghai"},
	}
}

func loadYAML(path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return fmt.Errorf("decode config file %s: %w", path, err)
	}
	return nil
}

func applyEnvironment(cfg *Config, loadErrors *[]error) {
	cfg.App.Environment = env("APP_ENV", cfg.App.Environment)
	cfg.App.Name = env("APP_NAME", cfg.App.Name)
	cfg.App.Version = env("APP_VERSION", cfg.App.Version)
	cfg.HTTP.Host = env("HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt(loadErrors, "HTTP_PORT", cfg.HTTP.Port)
	cfg.HTTP.ReadTimeout = envDuration(loadErrors, "HTTP_READ_TIMEOUT", cfg.HTTP.ReadTimeout)
	cfg.HTTP.ReadHeaderTimeout = envDuration(loadErrors, "HTTP_READ_HEADER_TIMEOUT", cfg.HTTP.ReadHeaderTimeout)
	cfg.HTTP.WriteTimeout = envDuration(loadErrors, "HTTP_WRITE_TIMEOUT", cfg.HTTP.WriteTimeout)
	cfg.HTTP.IdleTimeout = envDuration(loadErrors, "HTTP_IDLE_TIMEOUT", cfg.HTTP.IdleTimeout)
	cfg.HTTP.ShutdownTimeout = envDuration(loadErrors, "HTTP_SHUTDOWN_TIMEOUT", cfg.HTTP.ShutdownTimeout)
	cfg.HTTP.AllowedOrigins = envCSV("HTTP_ALLOWED_ORIGINS", cfg.HTTP.AllowedOrigins)
	cfg.Database.Host = env("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = envInt(loadErrors, "DB_PORT", cfg.Database.Port)
	cfg.Database.Name = env("DB_NAME", cfg.Database.Name)
	cfg.Database.User = env("DB_USER", cfg.Database.User)
	cfg.Database.Password = env("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.SSLMode = env("DB_SSLMODE", cfg.Database.SSLMode)
	cfg.Database.MaxOpenConns = envInt(loadErrors, "DB_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	cfg.Database.MaxIdleConns = envInt(loadErrors, "DB_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)
	cfg.Database.ConnMaxLifetime = envDuration(loadErrors, "DB_CONN_MAX_LIFETIME", cfg.Database.ConnMaxLifetime)
	cfg.Database.ConnMaxIdleTime = envDuration(loadErrors, "DB_CONN_MAX_IDLE_TIME", cfg.Database.ConnMaxIdleTime)
	cfg.Database.ConnectTimeout = envDuration(loadErrors, "DB_CONNECT_TIMEOUT", cfg.Database.ConnectTimeout)
	cfg.Log.Level = env("LOG_LEVEL", cfg.Log.Level)
	cfg.Log.Format = env("LOG_FORMAT", cfg.Log.Format)
	cfg.Log.TimeZone = env("LOG_TIMEZONE", cfg.Log.TimeZone)
}

func (c Config) Validate() error {
	var validationErrors []error
	if c.App.Name == "" {
		validationErrors = append(validationErrors, errors.New("app name must not be empty"))
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		validationErrors = append(validationErrors, errors.New("HTTP port must be between 1 and 65535"))
	}
	if c.Database.Host == "" || c.Database.Name == "" || c.Database.User == "" {
		validationErrors = append(validationErrors, errors.New("database host, name and user must not be empty"))
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		validationErrors = append(validationErrors, errors.New("database port must be between 1 and 65535"))
	}
	if c.Database.MaxIdleConns < 0 || c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		validationErrors = append(validationErrors, errors.New("database pool sizes are invalid"))
	}
	if c.Log.Format != "json" && c.Log.Format != "text" {
		validationErrors = append(validationErrors, errors.New("log format must be json or text"))
	}
	if _, err := time.LoadLocation(c.Log.TimeZone); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("log timezone must be a valid IANA timezone: %w", err))
	}
	return errors.Join(validationErrors...)
}

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
