package user

import (
	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const apiVersion = "/api/v1"

func RegisterAPI(engine *gin.Engine, db *gorm.DB, responder common.Responder) {
	handler := NewHandler(NewService(NewRepository(db)), responder)

	api := engine.Group(apiVersion)

	{
		routes := api.Group("/users")

		// /api/v1/users
		routes.POST("/register", handler.register)
		routes.GET("/:user_id", handler.get)
		routes.DELETE("/:user_id", handler.delete)
	}
}
