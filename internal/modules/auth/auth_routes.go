package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	a := r.Group("/auth")
	{
		a.POST("/register", h.Register)
		a.POST("/login", h.Login)
		a.POST("/refresh", h.Refresh)
		a.POST("/logout", h.Logout)
	}
}
