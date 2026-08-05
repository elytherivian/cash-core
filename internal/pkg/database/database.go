package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Connection struct {
	GORM *gorm.DB
	SQL  *sql.DB
}

func Open(ctx context.Context, cfg config.Database, log *slog.Logger, logLevel string) (*Connection, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DSN(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger:         newGORMLogger(log, logLevel),
		TranslateError: true,
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingContext, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingContext); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return &Connection{GORM: db, SQL: sqlDB}, nil
}

func (c *Connection) Ping(ctx context.Context) error {
	return c.SQL.PingContext(ctx)
}

func (c *Connection) Close() error {
	return c.SQL.Close()
}

func NormalizeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return common.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return common.ErrConflict
	case errors.Is(err, gorm.ErrForeignKeyViolated), errors.Is(err, gorm.ErrCheckConstraintViolated):
		return fmt.Errorf("%w: database constraint violation", common.ErrInvalidInput)
	default:
		return err
	}
}

type gormLogger struct {
	logger *slog.Logger
	level  logger.LogLevel
}

func newGORMLogger(log *slog.Logger, configuredLevel string) logger.Interface {
	level := logger.Warn
	switch strings.ToLower(configuredLevel) {
	case "debug":
		level = logger.Info
	case "error":
		level = logger.Error
	}
	return &gormLogger{logger: log.With("component", "gorm"), level: level}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *gormLogger) Info(ctx context.Context, message string, args ...any) {
	if l.level >= logger.Info {
		l.logger.DebugContext(ctx, fmt.Sprintf(message, args...))
	}
}

func (l *gormLogger) Warn(ctx context.Context, message string, args ...any) {
	if l.level >= logger.Warn {
		l.logger.WarnContext(ctx, fmt.Sprintf(message, args...))
	}
}

func (l *gormLogger) Error(ctx context.Context, message string, args ...any) {
	if l.level >= logger.Error {
		l.logger.ErrorContext(ctx, fmt.Sprintf(message, args...))
	}
}

func (l *gormLogger) Trace(ctx context.Context, started time.Time, trace func() (string, int64), err error) {
	elapsed := time.Since(started)
	statement, rows := trace()
	attributes := []any{"duration", elapsed, "rows", rows, "sql", statement}
	switch {
	case err != nil && l.level >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.logger.ErrorContext(ctx, "database query failed", append(attributes, "error", err)...)
	case elapsed > 200*time.Millisecond && l.level >= logger.Warn:
		l.logger.WarnContext(ctx, "slow database query", attributes...)
	case l.level >= logger.Info:
		l.logger.DebugContext(ctx, "database query", attributes...)
	}
}
