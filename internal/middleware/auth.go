package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"github.com/thangnguyen/todo_api_v1/pkg/jwt"
)

const (
	ctxUserID = "userID"
	ctxRole   = "role"
	ctxEmail  = "email"
)

// Auth xác thực access token (header "Authorization: Bearer <token>").
// Hợp lệ thì set userID/role/email vào context; không thì 401.
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"success": false, "error": "missing or malformed Authorization header"})
			return
		}
		claims, err := jwtManager.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"success": false, "error": "invalid or expired token"})
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Set(ctxEmail, claims.Email)
		c.Next()
	}
}

// RequireAdmin chặn route chỉ dành cho admin. Phải đặt SAU Auth.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(ctxRole) != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden,
				gin.H{"success": false, "error": "admin only"})
			return
		}
		c.Next()
	}
}

// UserID trả về ID của user đang đăng nhập (set bởi Auth middleware).
func UserID(c *gin.Context) string { return c.GetString(ctxUserID) }

// IsAdmin kiểm tra user hiện tại có phải admin không.
func IsAdmin(c *gin.Context) bool { return c.GetString(ctxRole) == domain.RoleAdmin }
