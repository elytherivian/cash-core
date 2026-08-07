package user

import (
	"strings"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"column:username;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	common.Lifecycle
}

func (User) TableName() string { return "users" }

type CreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Normalize 去掉开头结尾的空白字符 空格 换行 tab 等
func (r *CreateRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Normalize 去掉开头结尾的空白字符 空格 换行 tab 等
func (r *RegisterUserRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
}

type Response struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IsActive  bool       `json:"is_active"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (u *User) Response() Response {
	return Response{
		ID: u.ID, Username: u.Username, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		IsActive: u.IsActive, DeletedAt: u.DeletedAt,
	}
}
