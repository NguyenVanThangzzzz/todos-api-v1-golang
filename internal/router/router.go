package router

import (
	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/middleware"
	"github.com/thangnguyen/todo_api_v1/internal/modules/auth"
	"github.com/thangnguyen/todo_api_v1/internal/modules/todo"
	"github.com/thangnguyen/todo_api_v1/pkg/jwt"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
)

func New(todoH *todo.Handler, authH *auth.Handler, jwtManager *jwt.Manager, log *logger.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	auth.RegisterRoutes(v1, authH)
	todo.RegisterRoutes(v1, todoH, jwtManager)

	return r
}
