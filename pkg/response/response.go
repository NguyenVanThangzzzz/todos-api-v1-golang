package response

import "github.com/gin-gonic/gin"

// Envelope - format response chuẩn cho toàn bộ API
type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func OK(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Envelope{Success: true, Message: message, Data: data})
}

func Fail(c *gin.Context, status int, err string) {
	c.JSON(status, Envelope{Success: false, Error: err})
}
