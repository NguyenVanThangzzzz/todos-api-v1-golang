package domain

import "errors"

// Sentinel errors dùng chung cho tất cả module.
// Repository và usecase trả về các lỗi này (hoặc AppError bọc chúng);
// handler dùng errors.Is / errors.As để map sang HTTP status.
var (
	ErrNotFound     = errors.New("record not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
)
