package transaction

import (
	"context"

	"cash-core/internal/pkg/database"

	"gorm.io/gorm"
)

type Repository interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(transaction).Error)
}
