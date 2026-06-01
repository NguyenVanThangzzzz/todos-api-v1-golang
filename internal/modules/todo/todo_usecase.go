package todo

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
)

type Usecase struct {
	repo Repository
}

func NewUsecase(r Repository) *Usecase {
	return &Usecase{repo: r}
}

func (u *Usecase) Create(actorID string, req CreateRequest) (*domain.Todo, error) {
	now := time.Now()
	t := &domain.Todo{
		ID:          uuid.NewString(),
		UserID:      actorID,
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

func (u *Usecase) List(actorID string, isAdmin bool) ([]*domain.Todo, error) {
	if isAdmin {
		return u.repo.List()
	}
	return u.repo.ListByUser(actorID)
}

func (u *Usecase) GetByID(actorID string, isAdmin bool, id string) (*domain.Todo, error) {
	t, err := u.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewNotFound("todo not found")
		}
		return nil, err
	}
	if !isAdmin && t.UserID != actorID {
		return nil, domain.NewForbidden("you do not own this todo")
	}
	return t, nil
}

func (u *Usecase) Update(actorID string, isAdmin bool, id string, req UpdateRequest) (*domain.Todo, error) {
	t, err := u.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewNotFound("todo not found")
		}
		return nil, err
	}
	if !isAdmin && t.UserID != actorID {
		return nil, domain.NewForbidden("you do not own this todo")
	}
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
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewNotFound("todo not found")
		}
		return nil, err
	}
	return t, nil
}

func (u *Usecase) Delete(actorID string, isAdmin bool, id string) error {
	t, err := u.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewNotFound("todo not found")
		}
		return err
	}
	if !isAdmin && t.UserID != actorID {
		return domain.NewForbidden("you do not own this todo")
	}
	return u.repo.Delete(id)
}
