package todo

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"github.com/thangnguyen/todo_api_v1/internal/middleware"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
	"github.com/thangnguyen/todo_api_v1/pkg/response"
	"go.uber.org/zap"
)

func handleErr(c *gin.Context, err error) bool {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.Code, appErr.Message)
		return true
	}
	return false
}

type Handler struct {
	uc  *Usecase
	log *logger.Logger
}

func NewHandler(uc *Usecase, l *logger.Logger) *Handler {
	return &Handler{uc: uc, log: l}
}

// Create - POST /api/v1/todos
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := validateStruct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.uc.Create(middleware.UserID(c), req)
	if err != nil {
		h.log.Error("create todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusCreated, "todo created", t)
}

// List - GET /api/v1/todos
func (h *Handler) List(c *gin.Context) {
	todos, err := h.uc.List(middleware.UserID(c), middleware.IsAdmin(c))
	if err != nil {
		h.log.Error("list todos failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "", todos)
}

// GetByID - GET /api/v1/todos/:id
func (h *Handler) GetByID(c *gin.Context) {
	t, err := h.uc.GetByID(middleware.UserID(c), middleware.IsAdmin(c), c.Param("id"))
	if err != nil {
		if handleErr(c, err) {
			return
		}
		h.log.Error("get todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "", t)
}

// Update - PATCH /api/v1/todos/:id
func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := validateStruct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.uc.Update(middleware.UserID(c), middleware.IsAdmin(c), c.Param("id"), req)
	if err != nil {
		if handleErr(c, err) {
			return
		}
		h.log.Error("update todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "todo updated", t)
}

// Delete - DELETE /api/v1/todos/:id
func (h *Handler) Delete(c *gin.Context) {
	err := h.uc.Delete(middleware.UserID(c), middleware.IsAdmin(c), c.Param("id"))
	if err != nil {
		if handleErr(c, err) {
			return
		}
		h.log.Error("delete todo failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "todo deleted", nil)
}
