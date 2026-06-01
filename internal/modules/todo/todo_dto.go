package todo

type CreateRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

// UpdateRequest dùng con trỏ để phân biệt "không truyền" (nil) và "truyền giá trị rỗng/false".
type UpdateRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Completed   *bool   `json:"completed,omitempty"`
}
