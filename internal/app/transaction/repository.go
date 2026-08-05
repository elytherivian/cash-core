package transaction

import (
	"context"
	"time"

	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, value *Transaction) error
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Transaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID, filter Filter) ([]Transaction, int64, error)
	Delete(ctx context.Context, userID, id uuid.UUID, deletedAt time.Time) error
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, value *Transaction) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(value).Error)
}

func (r *repository) FindByID(ctx context.Context, userID, id uuid.UUID) (*Transaction, error) {
	var value Transaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ? AND is_active = TRUE", userID, id).
		First(&value).Error
	return &value, database.NormalizeError(err)
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID, filter Filter) ([]Transaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&Transaction{}).
		Where("user_id = ? AND is_active = TRUE", userID)
	if filter.From != nil {
		query = query.Where("occurred_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("occurred_at < ?", *filter.To)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, database.NormalizeError(err)
	}
	var values []Transaction
	err := query.Order("occurred_at DESC, id DESC").
		Limit(filter.Page.Limit).Offset(filter.Page.Offset).Find(&values).Error
	return values, total, database.NormalizeError(err)
}

func (r *repository) Delete(ctx context.Context, userID, id uuid.UUID, deletedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&Transaction{}).
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
