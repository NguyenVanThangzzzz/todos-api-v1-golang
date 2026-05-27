package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
	"go.uber.org/zap"
)

// RequestLogger là middleware ghi log mỗi request - giống morgan ở Node.
// Trong Gin, middleware là một func(*gin.Context). Gọi c.Next() để cho request
// chạy tiếp xuống handler; code sau c.Next() sẽ chạy khi handler đã trả về.
func RequestLogger(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // chạy handler

		l.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// Recovery bắt panic trong handler, log lại và trả 500 thay vì crash server.
func Recovery(l *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		l.Error("panic recovered", zap.Any("error", err), zap.String("path", c.Request.URL.Path))
		c.AbortWithStatusJSON(500, gin.H{"success": false, "error": "internal server error"})
	})
}
