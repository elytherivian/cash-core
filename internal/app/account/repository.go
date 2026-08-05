package account

import (
	"context"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, value *Account) error
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Account, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page common.Page) ([]Account, int64, error)
	Delete(ctx context.Context, userID, id uuid.UUID, deletedAt time.Time) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, value *Account) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(value).Error)
}

func (r *repository) FindByID(ctx context.Context, userID, id uuid.UUID) (*Account, error) {
	var value Account
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ? AND is_active = TRUE", userID, id).
		First(&value).Error
	return &value, database.NormalizeError(err)
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID, page common.Page) ([]Account, int64, error) {
	query := r.db.WithContext(ctx).Model(&Account{}).
		Where("user_id = ? AND is_active = TRUE", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, database.NormalizeError(err)
	}
	var values []Account
	err := query.Order("created_at DESC").Limit(page.Limit).Offset(page.Offset).Find(&values).Error
	return values, total, database.NormalizeError(err)
}

func (r *repository) Delete(ctx context.Context, userID, id uuid.UUID, deletedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&Account{}).
		Where("user_id = ? AND id = ? AND is_active = TRUE", userID, id).
		Updates(map[string]any{"is_active": false, "deleted_at": deletedAt, "updated_at": deletedAt})
	if result.Error != nil {
		return database.NormalizeError(result.Error)
	}
	if result.RowsAffected == 0 {
		return database.NormalizeError(gorm.ErrRecordNotFound)
	}
	return nil
}
