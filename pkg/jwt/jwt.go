package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/thangnguyen/todo_api_v1/config"
)

// Claims là payload của access token (JWT, ký HS256).
type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Manager phát hành và xác thực access token.
type Manager struct {
	secret    []byte
	accessTTL time.Duration
}

func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{
		secret:    []byte(cfg.Secret),
		accessTTL: cfg.AccessTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

// Generate tạo access token cho user.
func (m *Manager) Generate(userID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse xác thực chữ ký + hạn dùng, trả về claims nếu hợp lệ.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
