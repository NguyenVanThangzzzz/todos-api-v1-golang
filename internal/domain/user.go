package domain

import "time"

// Vai trò của user. Register luôn tạo RoleUser; admin được nâng quyền thủ công.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User là entity tài khoản.
// Quan hệ 1-N với Todo: 1 User sở hữu nhiều Todo (khóa ngoại Todo.UserID).
type User struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"` // đã hash bằng bcrypt, không bao giờ serialize
	Role      string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// has-many: khai báo quan hệ + ràng buộc khóa ngoại ON DELETE CASCADE.
	Todos []Todo `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"todos,omitempty"`
}

func (User) TableName() string { return "users" }

func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
