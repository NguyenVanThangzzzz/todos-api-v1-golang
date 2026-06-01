package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"github.com/thangnguyen/todo_api_v1/pkg/hash"
	"github.com/thangnguyen/todo_api_v1/pkg/jwt"
)


type TokenResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type Usecase struct {
	users      UserRepository
	refresh    RefreshTokenRepository
	jwt        *jwt.Manager
	refreshTTL time.Duration
}

func NewUsecase(u UserRepository, rt RefreshTokenRepository, j *jwt.Manager) *Usecase {
	return &Usecase{
		users:      u,
		refresh:    rt,
		jwt:        j,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

func (a *Usecase) Register(req RegisterRequest) (*domain.User, error) {
	hashed, err := hash.Password(req.Password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		ID:       uuid.NewString(),
		Email:    req.Email,
		Password: hashed,
		Role:     domain.RoleUser,
	}
	if err := a.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (a *Usecase) Login(req LoginRequest) (*TokenResult, error) {
	u, err := a.users.GetByEmail(req.Email)
	if err != nil {
		return nil, domain.NewUnauthorized("invalid email or password")
	}
	if !hash.Compare(u.Password, req.Password) {
		return nil, domain.NewUnauthorized("invalid email or password")
	}
	return a.issueTokens(u)
}

func (a *Usecase) Refresh(req RefreshRequest) (*TokenResult, error) {
	h := hashToken(req.RefreshToken)
	stored, err := a.refresh.GetByHash(h)
	if err != nil {
		return nil, domain.NewUnauthorized("invalid or expired refresh token")
	}

	// Rotation: xóa token cũ trước, sau đó kiểm tra hạn dùng.
	if err := a.refresh.DeleteByHash(h); err != nil {
		return nil, err
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, domain.NewUnauthorized("invalid or expired refresh token")
	}

	u, err := a.users.GetByID(stored.UserID)
	if err != nil {
		return nil, domain.NewUnauthorized("invalid or expired refresh token")
	}
	return a.issueTokens(u)
}

func (a *Usecase) Logout(req LogoutRequest) error {
	return a.refresh.DeleteByHash(hashToken(req.RefreshToken))
}

func (a *Usecase) issueTokens(u *domain.User) (*TokenResult, error) {
	access, err := a.jwt.Generate(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, err
	}
	raw, h, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: h,
		ExpiresAt: time.Now().Add(a.refreshTTL),
	}
	if err := a.refresh.Create(rt); err != nil {
		return nil, err
	}
	return &TokenResult{
		User:         u,
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(a.jwt.AccessTTL().Seconds()),
	}, nil
}

func generateRefreshToken() (raw, hashHex string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
