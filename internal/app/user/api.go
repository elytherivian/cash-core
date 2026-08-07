package user

import (
	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const apiVersion = "/api/v1"

func RegisterAPI(engine *gin.Engine, db *gorm.DB, responder common.Responder) {
	handler := NewHandler(NewService(NewRepository(db)), responder)
	routes := engine.Group(apiVersion + "/users")
	routes.POST("", handler.create)
	routes.GET("/:user_id", handler.get)
	routes.DELETE("/:user_id", handler.delete)
}
