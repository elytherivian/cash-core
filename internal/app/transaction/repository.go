package transaction

import (
	"context"

	"cash-core/internal/common"
	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	UpdateTransaction(ctx context.Context, userID, transactionID uuid.UUID, request UpdateTransactionRequest) (*Transaction, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(transaction).Error)
}

func (r *repository) UpdateTransaction(
	ctx context.Context,
	userID, transactionID uuid.UUID,
	request UpdateTransactionRequest,
) (*Transaction, error) {
	updates := make(map[string]any, 5)
	if request.Type != nil {
		updates["type"] = *request.Type
	}
	if request.Amount != nil {
		updates["amount"] = *request.Amount
	}
	if request.AccountID != nil {
		updates["account_id"] = *request.AccountID
	}
	if request.CategoryID != nil {
		updates["category_id"] = *request.CategoryID
	}
	if request.OccurredAt != nil {
		updates["occurred_at"] = *request.OccurredAt
	}

	var updatedTransaction Transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Transaction{}).
			Where("id = ? AND user_id = ? AND is_active = TRUE", transactionID, userID).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return common.ErrNotFound
		}
		return tx.Where("id = ? AND user_id = ? AND is_active = TRUE", transactionID, userID).
			First(&updatedTransaction).Error
	})
	if err != nil {
		return nil, database.NormalizeError(err)
	}
	return &updatedTransaction, nil
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
