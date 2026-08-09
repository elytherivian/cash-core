package user

import (
	"strings"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

// User 数据库模型
type User struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"column:username;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	common.Lifecycle
}

func (User) TableName() string { return "users" }

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *RegisterUserRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	common.LifecycleResponse
}

func (u User) Response(location *time.Location) UserResponse {
	return UserResponse{
		ID: u.ID, Username: u.Username, LifecycleResponse: u.Lifecycle.Response(location),
	}
}

type DeleteUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *DeleteUserRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
}

type RestoreUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *RestoreUserRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *LoginRequest) Normalize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *RefreshTokenRequest) Normalize() {
	r.RefreshToken = strings.TrimSpace(r.RefreshToken)
}
