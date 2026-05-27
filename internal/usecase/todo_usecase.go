package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"github.com/thangnguyen/todo_api_v1/internal/dto"
	"github.com/thangnguyen/todo_api_v1/internal/repository"
)

// TodoUsecase chứa toàn bộ business logic.
// Handler (HTTP) gọi xuống đây; usecase gọi xuống repository.
// Đây là nơi đặt rule: ai có quyền sửa, chuyển trạng thái ra sao, ...
type TodoUsecase struct {
	repo repository.TodoRepository
}

func NewTodoUsecase(r repository.TodoRepository) *TodoUsecase {
	return &TodoUsecase{repo: r}
}

func (u *TodoUsecase) Create(req dto.CreateTodoRequest) (*domain.Todo, error) {
	now := time.Now()
	t := &domain.Todo{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := u.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (u *TodoUsecase) GetByID(id string) (*domain.Todo, error) {
	return u.repo.GetByID(id)
}

func (u *TodoUsecase) List() []*domain.Todo {
	return u.repo.List()
}

func (u *TodoUsecase) Update(id string, req dto.UpdateTodoRequest) (*domain.Todo, error) {
	t, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Partial update: chỉ gán field nào client thật sự truyền lên (con trỏ != nil).
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Completed != nil {
		t.Completed = *req.Completed
	}
	t.UpdatedAt = time.Now()

	if err := u.repo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (u *TodoUsecase) Delete(id string) error {
	return u.repo.Delete(id)
}
