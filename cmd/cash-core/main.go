package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cash-core/internal/config"
	"cash-core/internal/pkg/database"
	"cash-core/internal/pkg/logger"
	"cash-core/internal/router"
)

func main() {
	if err := run(); err != nil {
		// message key value
		// slog 存在包级别的默认 logger 使用 slog.SetDefault 设置 在 run() 中设置
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Config 对象 app http database log 加载环境变量并校验
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	log := logger.New(cfg.Log)
	// 设置默认 logger 时区转换 日志最低等级 日志输出格式
	slog.SetDefault(log)
	slog.Info("application started", "name", cfg.App.Name, "version", cfg.App.Version, "environment", cfg.App.Environment)

	// os/signal 包在监听操作系统发来的信号 os.Interrupt SIGTERM
	// 创建 ctx 上下文 当收到信号时 ctx.Done() 会被关闭 不再阻塞
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// 如果接收到关闭信号 最后执行清理操作
	// 如果 run() 函数返回 err defer stop() 会被执行 取消 ctx
	defer stop()

	// 为什么还是把 log 注入而不是 slog
	// 因为设置 slog 默认 logger 是为了全局兜底 把 log 传给后端代码 是为了依赖注入和模块化
	connection, err := database.Open(ctx, cfg.Database, log, cfg.Log.Level)
	if err != nil {
		return err
	}
	// 断开数据库连接
	defer func() {
		if err := connection.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}()
	if err := database.InitializeSchema(connection.GORM); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.HTTP.Address(),
		Handler:           router.New(cfg, connection.GORM, connection, log),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}
	serveError := make(chan error, 1)
	// 启动 HTTP 服务
	go func() {
		slog.Info("HTTP server started", "address", server.Addr, "environment", cfg.App.Environment)
		serveError <- server.ListenAndServe()
	}()

	// select 监听多个 channel
	select {
	case err := <-serveError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		// 运行到这里说明接收到关闭信号 os.Interrupt SIGTERM
		// 此时 ctx 已经取消
		slog.Info("shutdown signal received")
	}

	// 接收到关闭信号后 创建一个新的 context 用于优雅关闭 HTTP 服务
	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	if err := <-serveError; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	slog.Info("application stopped")
	return nil
}
