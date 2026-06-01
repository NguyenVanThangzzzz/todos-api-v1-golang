package todo

import (
	"errors"

	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"gorm.io/gorm"
)

type Repository interface {
	Create(t *domain.Todo) error
	GetByID(id string) (*domain.Todo, error)
	List() ([]*domain.Todo, error)
	ListByUser(userID string) ([]*domain.Todo, error)
	Update(t *domain.Todo) error
	Delete(id string) error
}

type postgresRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(t *domain.Todo) error {
	return r.db.Create(t).Error
}

func (r *postgresRepo) GetByID(id string) (*domain.Todo, error) {
	var t domain.Todo
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresRepo) List() ([]*domain.Todo, error) {
	var todos []*domain.Todo
	if err := r.db.Order("created_at DESC").Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *postgresRepo) ListByUser(userID string) ([]*domain.Todo, error) {
	var todos []*domain.Todo
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *postgresRepo) Update(t *domain.Todo) error {
	res := r.db.Save(t)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *postgresRepo) Delete(id string) error {
	res := r.db.Delete(&domain.Todo{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
