package todo

import (
	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/middleware"
	"github.com/thangnguyen/todo_api_v1/pkg/jwt"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, jwtManager *jwt.Manager) {
	todos := r.Group("/todos")
	todos.Use(middleware.Auth(jwtManager))
	{
		todos.POST("", h.Create)
		todos.GET("", h.List)
		todos.GET(":id", h.GetByID)
		todos.PATCH(":id", h.Update)
		todos.DELETE(":id", h.Delete)
	}
}
