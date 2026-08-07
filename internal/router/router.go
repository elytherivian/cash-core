package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"cash-core/internal/app/account"
	"cash-core/internal/app/category"
	"cash-core/internal/app/transaction"
	"cash-core/internal/app/user"
	"cash-core/internal/common"
	"cash-core/internal/config"
	"cash-core/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Pinger interface {
	Ping(context.Context) error
}

func New(cfg config.Config, db *gorm.DB, pinger Pinger, log *slog.Logger) *gin.Engine {
	setMode(cfg.App.Environment)
	responder := common.NewResponder(cfg.App.Version)

	engine := gin.New()
	// 开启 405 Method Not Allowed 支持
	engine.HandleMethodNotAllowed = true
	// 设置可信代理 nil 指的是不信任任何代理
	_ = engine.SetTrustedProxies(nil)
	// 给每个请求生成或读取一个 Request ID
	engine.Use(middleware.RequestID())
	// 处理 panic 防止服务因为某个请求崩掉 返回错误信息
	engine.Use(middleware.Recovery(log, responder))
	// 设置请求超时 30s 超时后返回 504 Gateway Timeout
	engine.Use(middleware.Timeout(30 * time.Second))
	engine.Use(middleware.SecurityHeaders())
	// 设置跨域请求允许的来源 UI 只做桌面端和移动端 不用设置
	engine.Use(middleware.CORS(cfg.HTTP.AllowedOrigins))
	engine.Use(middleware.AccessLog(log))

	// 注册系统级别的健康检查路由 /health/live /health/ready
	registerSystemRoutes(engine, responder, pinger, log, cfg.App)

	// 注册 app 路由
	user.RegisterAPI(engine, db, responder)
	account.RegisterAPI(engine, db, responder)
	category.RegisterAPI(engine, db, responder)
	transaction.RegisterAPI(engine, db, responder)

	// 注册全局路由 404 Not Found 405 Method Not Allowed
	// 处理未注册的路由和方法 返回错误信息
	engine.NoRoute(func(c *gin.Context) { responder.Error(c, common.ErrNotFound) })
	engine.NoMethod(func(c *gin.Context) { responder.Error(c, common.ErrMethodNotAllowed) })
	return engine
}

// registerSystemRoutes 注册系统级别的路由
// 健康检查 /health/live /health/ready
func registerSystemRoutes(
	engine *gin.Engine,
	responder common.Responder,
	pinger Pinger,
	log *slog.Logger,
	app config.App,
) {
	engine.GET("/health/live", func(c *gin.Context) {
		responder.Success(c, http.StatusOK, "ok", gin.H{"name": app.Name, "version": app.Version})
	})
	engine.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := pinger.Ping(ctx); err != nil {
			log.ErrorContext(ctx, "readiness check failed", "error", err)
			responder.Error(c, common.ErrUnavailable)
			return
		}
		responder.Success(c, http.StatusOK, "ready", nil)
	})
}

// setMode 设置 gin 的运行模式
// 不同模式下 gin 的日志输出和调试信息不同
func setMode(environment string) {
	switch environment {
	case "production", "staging":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
