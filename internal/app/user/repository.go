package user

import (
	"context"
	"time"

	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Delete(ctx context.Context, id uuid.UUID, deletedAt time.Time) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 数据库创建用户
func (r *repository) Create(ctx context.Context, user *User) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(user).Error)
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var value User
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_active = TRUE", id).
		First(&value).Error
	return &value, database.NormalizeError(err)
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID, deletedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND is_active = TRUE", id).
		Updates(map[string]any{"is_active": false, "deleted_at": deletedAt, "updated_at": deletedAt})
	if result.Error != nil {
		return database.NormalizeError(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.NormalizeError(gorm.ErrRecordNotFound)
	}
	return nil
}
