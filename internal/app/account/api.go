package account

import (
	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const apiVersion = "/api/v1"

func RegisterAPI(engine *gin.Engine, db *gorm.DB, responder common.Responder) {
	handler := NewHandler(NewService(NewRepository(db)), responder)
	routes := engine.Group(apiVersion + "/users/:user_id/accounts")
	routes.POST("", handler.create)
	routes.GET("", handler.list)
	routes.GET("/:id", handler.get)
	routes.DELETE("/:id", handler.delete)
}
