package transaction

import (
	"context"
	"fmt"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Service interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, request CreateTransactionRequest) (*Transaction, error)
	UpdateTransaction(ctx context.Context, userID, transactionID uuid.UUID, request UpdateTransactionRequest) (*Transaction, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error)
}

type service struct{ repository Repository }

func NewService(repository Repository) Service { return &service{repository: repository} }

func (s *service) CreateTransaction(ctx context.Context, userID uuid.UUID, request CreateTransactionRequest) (*Transaction, error) {
	request.Normalize()
	if !request.Type.Valid() {
		return nil, fmt.Errorf("%w: type must be income or expense", common.ErrInvalidInput)
	}
	if !request.Amount.IsPositive() {
		return nil, fmt.Errorf("%w: amount must be positive", common.ErrInvalidInput)
	}
	if request.AccountID == uuid.Nil || request.CategoryID == uuid.Nil {
		return nil, fmt.Errorf("%w: account_id and category_id are required", common.ErrInvalidInput)
	}
	if request.OccurredAt.IsZero() {
		request.OccurredAt = time.Now().UTC()
	}
	transaction := &Transaction{
		ID: uuid.New(), UserID: userID, Type: request.Type, Amount: request.Amount,
		AccountID: request.AccountID, CategoryID: request.CategoryID, OccurredAt: request.OccurredAt,
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.CreateTransaction(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *service) UpdateTransaction(
	ctx context.Context,
	userID, transactionID uuid.UUID,
	request UpdateTransactionRequest,
) (*Transaction, error) {
	request.Normalize()
	if userID == uuid.Nil || transactionID == uuid.Nil {
		return nil, fmt.Errorf("%w: transaction id is required", common.ErrInvalidInput)
	}
	if !request.HasChanges() {
		return nil, fmt.Errorf("%w: at least one field is required", common.ErrInvalidInput)
	}
	if request.Type != nil && !request.Type.Valid() {
		return nil, fmt.Errorf("%w: type must be income or expense", common.ErrInvalidInput)
	}
	if request.Amount != nil && !request.Amount.IsPositive() {
		return nil, fmt.Errorf("%w: amount must be positive", common.ErrInvalidInput)
	}
	if request.AccountID != nil && *request.AccountID == uuid.Nil {
		return nil, fmt.Errorf("%w: account_id is required", common.ErrInvalidInput)
	}
	if request.CategoryID != nil && *request.CategoryID == uuid.Nil {
		return nil, fmt.Errorf("%w: category_id is required", common.ErrInvalidInput)
	}
	if request.OccurredAt != nil && request.OccurredAt.IsZero() {
		return nil, fmt.Errorf("%w: occurred_at must not be zero", common.ErrInvalidInput)
	}
	return s.repository.UpdateTransaction(ctx, userID, transactionID, request)
}

func (s *service) ListTransactions(ctx context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error) {
	if request.AccountID == nil && request.CategoryID == nil {
		return nil, fmt.Errorf("%w: account_id or category_id is required", common.ErrInvalidInput)
	}
	return s.repository.ListTransactions(ctx, userID, request)
}
