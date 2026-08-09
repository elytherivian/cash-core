package account

import (
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const apiVersion = "/api/v1"

func RegisterAPI(engine *gin.Engine, db *gorm.DB, responder common.Responder, location *time.Location, verifier middleware.TokenVerifier) {
	handler := NewHandler(NewService(NewRepository(db)), responder, location)
	api := engine.Group(apiVersion)

	{
		routes := api.Group("/accounts")
		routes.Use(middleware.Authentication(verifier, responder))

		// /api/v1/accounts
		routes.POST("/create", handler.create)
		routes.GET("/list", handler.listAccounts)
	}
}
