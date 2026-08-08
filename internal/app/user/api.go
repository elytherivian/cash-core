package user

import (
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const apiVersion = "/api/v1"

func RegisterAPI(
	engine *gin.Engine,
	db *gorm.DB,
	responder common.Responder,
	location *time.Location,
	tokens *auth.Manager,
) {
	handler := NewHandler(NewService(NewRepository(db), tokens), responder, location)

	api := engine.Group(apiVersion)

	{
		routes := api.Group("/users")

		// /api/v1/users
		routes.POST("/register", handler.register)
		routes.DELETE("/delete", handler.delete)
		routes.POST("/restore", handler.restore)
	}

	{
		routes := api.Group("/auth")

		// /api/v1/auth
		routes.POST("/login", handler.login)
		routes.POST("/refresh", handler.refresh)
	}
}
