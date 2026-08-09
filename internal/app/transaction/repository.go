package transaction

import (
	"context"

	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	ListTransactions(ctx context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(transaction).Error)
}

func (r *repository) ListTransactions(
	ctx context.Context,
	userID uuid.UUID,
	request ListTransactionsRequest,
) ([]Transaction, error) {
	query := r.db.WithContext(ctx).Model(&Transaction{}).
		Where("user_id = ? AND is_active = TRUE", userID)
	if request.AccountID != nil {
		query = query.Where("account_id = ?", *request.AccountID)
	}
	if request.CategoryID != nil {
		query = query.Where("category_id = ?", *request.CategoryID)
	}

	transactions := make([]Transaction, 0)
	err := query.Order("occurred_at ASC, id ASC").Find(&transactions).Error
	return transactions, database.NormalizeError(err)
}
