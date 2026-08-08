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
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByUsernameIncludingDeleted(ctx context.Context, username string) (*User, error)
	Delete(ctx context.Context, id uuid.UUID, deletedAt time.Time) error
	Restore(ctx context.Context, id uuid.UUID, restoredAt time.Time) error
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

// FindByUsername 只返回尚未软删除的有效用户。
func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var value User
	err := r.db.WithContext(ctx).
		Where("username = ? AND is_active = TRUE AND deleted_at IS NULL", username).
		First(&value).Error
	return &value, database.NormalizeError(err)
}

// FindByUsernameIncludingDeleted 同时查询有效和已软删除用户，供恢复流程判断状态。
func (r *repository) FindByUsernameIncludingDeleted(ctx context.Context, username string) (*User, error) {
	var value User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&value).Error
	return &value, database.NormalizeError(err)
}

// Delete 通过更新生命周期字段完成软删除，不物理移除用户记录。
func (r *repository) Delete(ctx context.Context, id uuid.UUID, deletedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND is_active = TRUE AND deleted_at IS NULL", id).
		Updates(map[string]any{"is_active": false, "deleted_at": deletedAt, "updated_at": deletedAt})
	if result.Error != nil {
		return database.NormalizeError(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.NormalizeError(gorm.ErrRecordNotFound)
	}
	return nil
}

// Restore 只恢复当前处于软删除状态的用户。
func (r *repository) Restore(ctx context.Context, id uuid.UUID, restoredAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND is_active = FALSE AND deleted_at IS NOT NULL", id).
		Updates(map[string]any{"is_active": true, "deleted_at": nil, "updated_at": restoredAt})
	if result.Error != nil {
		return database.NormalizeError(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.NormalizeError(gorm.ErrRecordNotFound)
	}
	return nil
}
