package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"cash-core/internal/config"
)

func New(cfgLog config.Log) *slog.Logger {
	location, err := time.LoadLocation(cfgLog.TimeZone)
	if err != nil {
		// Config.Validate rejects invalid locations before logger initialization.
		// This should never happen, but if it does, fall back to UTC.
		location = time.UTC
	}

	level := new(slog.LevelVar)
	// debug < info < warn < error
	// 低于最低级别的日志不会输出 开发环境 debug 生产环境 info
	switch strings.ToLower(cfgLog.Level) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	// slog.HandlerOptions 允许自定义日志处理程序的行为
	// Level 设置日志级别 ReplaceAttr 允许自定义日志属性的处理方式 如时间戳的时区
	options := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attribute slog.Attr) slog.Attr {
			if attribute.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, attribute.Value.Time().In(location))
			}
			return attribute
		},
	}
	// 本地开发环境使用 text 开发者使用 生产环境使用 json 日志系统查询使用
	if cfgLog.Format == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}
