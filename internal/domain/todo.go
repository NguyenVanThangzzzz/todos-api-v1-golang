package domain

import "time"

// Todo là entity (model) - giống như interface/class trong Node.js
// Tag `json:"..."` quy định tên field khi serialize ra JSON
// Tag `gorm:"..."` quy định cách GORM map field xuống cột trong PostgreSQL
type Todo struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string    `gorm:"type:uuid;not null;index" json:"user_id"` // khóa ngoại -> users.id (chủ sở hữu todo)
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:varchar(1000)" json:"description"`
	Completed   bool      `gorm:"not null;default:false" json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName cố định tên bảng là "todos" trong PostgreSQL.
func (Todo) TableName() string { return "todos" }
