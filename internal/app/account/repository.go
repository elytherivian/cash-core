package account

import (
	"context"

	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, account *Account) error
	ListActiveAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, account *Account) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(account).Error)
}

func (r *repository) ListActiveAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	accounts := make([]Account, 0)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = TRUE", userID).
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, database.NormalizeError(err)
}
