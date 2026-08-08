package transaction

import (
	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const apiVersion = "/api/v1"

func RegisterAPI(engine *gin.Engine, db *gorm.DB, responder common.Responder, verifier middleware.TokenVerifier) {
	handler := NewHandler(NewService(NewRepository(db)), responder)
	api := engine.Group(apiVersion)

	{
		routes := api.Group("/transactions")
		routes.Use(middleware.Authentication(verifier, responder))

		// /api/v1/transactions
		routes.POST("", handler.create)
		routes.GET("", handler.list)
		routes.GET("/:id", handler.get)
		routes.DELETE("/:id", handler.delete)
	}
}
