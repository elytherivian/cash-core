package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"cash-core/internal/app/account"
	"cash-core/internal/app/category"
	transactionapp "cash-core/internal/app/transaction"
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
	engine.Use(middleware.Timeout(30 * time.Second))
	engine.Use(middleware.SecurityHeaders())
	engine.Use(middleware.CORS(cfg.HTTP.AllowedOrigins))
	engine.Use(middleware.AccessLog(log))

	registerSystemRoutes(engine, responder, pinger, log, cfg.App)
	api := engine.Group("/api/v1")

	user.NewHandler(user.NewService(user.NewRepository(db)), responder).RegisterRoutes(api)
	account.NewHandler(account.NewService(account.NewRepository(db)), responder).RegisterRoutes(api)
	category.NewHandler(category.NewService(category.NewRepository(db)), responder).RegisterRoutes(api)
	transactionapp.NewHandler(
		transactionapp.NewService(transactionapp.NewRepository(db)),
		responder,
	).RegisterRoutes(api)

	engine.NoRoute(func(c *gin.Context) { responder.Error(c, common.ErrNotFound) })
	engine.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, common.Response{
			Version: cfg.App.Version, Code: http.StatusMethodNotAllowed, Message: "method not allowed",
		})
	})
	return engine
}

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
	engine.GET("/api/v1", func(c *gin.Context) {
		responder.Success(c, http.StatusOK, "ok", gin.H{"name": app.Name, "version": app.Version})
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
