package domain

import "net/http"

// AppError mang theo HTTP status code và message cụ thể.
// Unwrap() giữ lại sentinel gốc nên errors.Is(appErr, domain.ErrNotFound) vẫn đúng.
//
// Handler pattern:
//
//	var appErr *domain.AppError
//	if errors.As(err, &appErr) {
//	    response.Fail(c, appErr.Code, appErr.Message)
//	    return
//	}
//	// unexpected error → log + 500
type AppError struct {
	Code    int    // HTTP status code
	Message string // thông điệp trả về client
	Err     error  // sentinel để errors.Is hoạt động
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NewNotFound(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Err: ErrNotFound}
}

func NewForbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg, Err: ErrForbidden}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg, Err: ErrUnauthorized}
}

func NewConflict(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: ErrConflict}
}
