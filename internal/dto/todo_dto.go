package dto

// CreateTodoRequest - DTO cho request tạo mới todo
// Tag `validate:"..."` hoạt động giống schema của Zod ở Node.js
//   - required: bắt buộc
//   - min=N / max=N: độ dài min/max của string
type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

// UpdateTodoRequest - DTO cho request cập nhật todo
// Dùng con trỏ (*) để phân biệt "không truyền" (nil) và "truyền giá trị rỗng/false"
// Đây là pattern PATCH partial update.
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Completed   *bool   `json:"completed,omitempty"`
}
