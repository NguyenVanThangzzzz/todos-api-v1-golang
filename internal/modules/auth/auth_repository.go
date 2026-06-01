package auth

import (
	"errors"

	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"gorm.io/gorm"
)

// --- User ---

type UserRepository interface {
	Create(u *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
}

type postgresUserRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) Create(u *domain.User) error {
	if err := r.db.Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.NewConflict("email already registered")
		}
		return err
	}
	return nil
}

func (r *postgresUserRepo) GetByEmail(email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepo) GetByID(id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// --- RefreshToken ---

type RefreshTokenRepository interface {
	Create(rt *domain.RefreshToken) error
	GetByHash(hash string) (*domain.RefreshToken, error)
	DeleteByHash(hash string) error
}

type postgresRefreshRepo struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &postgresRefreshRepo{db: db}
}

func (r *postgresRefreshRepo) Create(rt *domain.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *postgresRefreshRepo) GetByHash(hash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	if err := r.db.First(&rt, "token_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func (r *postgresRefreshRepo) DeleteByHash(hash string) error {
	return r.db.Delete(&domain.RefreshToken{}, "token_hash = ?", hash).Error
}
