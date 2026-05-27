package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/dto"
	"github.com/thangnguyen/todo_api_v1/internal/repository"
	"github.com/thangnguyen/todo_api_v1/internal/usecase"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
	"github.com/thangnguyen/todo_api_v1/pkg/response"
	"github.com/thangnguyen/todo_api_v1/pkg/validator"
	"go.uber.org/zap"
)

// TodoHandler là lớp trên cùng - nhận HTTP request, decode JSON, validate,
// gọi usecase, rồi trả response. Giống controller bên NestJS/Express.
type TodoHandler struct {
	uc        *usecase.TodoUsecase
	validator *validator.Validator
	log       *logger.Logger
}

func NewTodoHandler(uc *usecase.TodoUsecase, v *validator.Validator, l *logger.Logger) *TodoHandler {
	return &TodoHandler{uc: uc, validator: v, log: l}
}

// Create - POST /todos
func (h *TodoHandler) Create(c *gin.Context) {
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	t, err := h.uc.Create(req)
	if err != nil {
		h.log.Error("create todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusCreated, "todo created", t)
}

// List - GET /todos
func (h *TodoHandler) List(c *gin.Context) {
	response.OK(c, http.StatusOK, "", h.uc.List())
}

// GetByID - GET /todos/:id
func (h *TodoHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	t, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, "todo not found")
			return
		}
		h.log.Error("get todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "", t)
}

// Update - PATCH /todos/:id
func (h *TodoHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	t, err := h.uc.Update(id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, "todo not found")
			return
		}
		h.log.Error("update todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "todo updated", t)
}

// Delete - DELETE /todos/:id
func (h *TodoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, "todo not found")
			return
		}
		h.log.Error("delete todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "todo deleted", nil)
}
