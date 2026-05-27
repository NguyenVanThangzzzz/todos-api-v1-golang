package repository

import (
	"errors"
	"sync"

	"github.com/thangnguyen/todo_api_v1/internal/domain"
)

// ErrNotFound khi không tìm thấy todo theo id - lớp usecase sẽ map sang HTTP 404
var ErrNotFound = errors.New("todo not found")

// TodoRepository là interface mô tả các thao tác lưu trữ.
// Lớp usecase chỉ phụ thuộc interface này (Dependency Inversion).
// Sau này muốn đổi sang Postgres/Mongo chỉ cần viết struct mới implement interface.
type TodoRepository interface {
	Create(t *domain.Todo) error
	GetByID(id string) (*domain.Todo, error)
	List() []*domain.Todo
	Update(t *domain.Todo) error
	Delete(id string) error
}

// inMemoryRepo - implementation lưu trong map.
// sync.RWMutex bảo vệ map khỏi race condition khi nhiều goroutine cùng đọc/ghi
// (Go xử lý request đồng thời bằng goroutine, không single-threaded như Node).
type inMemoryRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.Todo
}

func NewInMemoryTodoRepository() TodoRepository {
	return &inMemoryRepo{store: make(map[string]*domain.Todo)}
}

func (r *inMemoryRepo) Create(t *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[t.ID] = t
	return nil
}

func (r *inMemoryRepo) GetByID(id string) (*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.store[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *inMemoryRepo) List() []*domain.Todo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Todo, 0, len(r.store))
	for _, t := range r.store {
		out = append(out, t)
	}
	return out
}

func (r *inMemoryRepo) Update(t *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[t.ID]; !ok {
		return ErrNotFound
	}
	r.store[t.ID] = t
	return nil
}

func (r *inMemoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return ErrNotFound
	}
	delete(r.store, id)
	return nil
}
