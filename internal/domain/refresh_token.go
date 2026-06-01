package domain

import "time"

// RefreshToken lưu refresh token dạng opaque (đã hash SHA-256) trong DB.
// Nhờ lưu DB nên logout có thể thu hồi token (xóa row), refresh có thể xoay token.
type RefreshToken struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"` // sha256 hex của token gốc
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	// belongs-to User; xóa user thì xóa luôn refresh token của họ.
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
