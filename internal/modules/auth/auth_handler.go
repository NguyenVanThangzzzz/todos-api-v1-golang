package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
	"github.com/thangnguyen/todo_api_v1/pkg/response"
	"go.uber.org/zap"
)

// handleErr trả về true nếu err là AppError đã biết (4xx); caller chỉ cần log + 500 khi trả về false.
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

// Register - POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := validateStruct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	u, err := h.uc.Register(req)
	if err != nil {
		if handleErr(c, err) {
			return
		}
		h.log.Error("register failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusCreated, "registered", toUserResponse(u))
}

// Login - POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := validateStruct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.uc.Login(req)
	if err != nil {
		if handleErr(c, err) {
			return
		}
		h.log.Error("login failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "logged in", toTokenResponse(res))
}

// Refresh - POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := validateStruct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.uc.Refresh(req)
	if err != nil {
		if handleErr(c, err) {
			return
		}
		h.log.Error("refresh failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "token refreshed", toTokenResponse(res))
}

// Logout - POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if err := validateStruct(req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.uc.Logout(req); err != nil {
		h.log.Error("logout failed", zap.Error(err))
		response.Fail(c, http.StatusInternalServerError, "internal server error")
		return
	}
	response.OK(c, http.StatusOK, "logged out", nil)
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{ID: u.ID, Email: u.Email, Role: u.Role}
}

func toTokenResponse(r *TokenResult) TokenResponse {
	return TokenResponse{
		User:         toUserResponse(r.User),
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    r.ExpiresIn,
	}
}
