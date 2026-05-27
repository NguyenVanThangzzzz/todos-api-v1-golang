package router

import (
	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/handler"
	"github.com/thangnguyen/todo_api_v1/internal/middleware"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
)

// New khởi tạo Gin engine và đăng ký routes.
// Tách hàm này khỏi main.go giúp test dễ hơn (có thể spin lên 1 engine trong test).
func New(todoH *handler.TodoHandler, log *logger.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		todos := v1.Group("/todos")
		{
			todos.POST("", todoH.Create)
			todos.GET("", todoH.List)
			todos.GET(":id", todoH.GetByID)
			todos.PATCH(":id", todoH.Update)
			todos.DELETE(":id", todoH.Delete)
		}
	}
	return r
}
