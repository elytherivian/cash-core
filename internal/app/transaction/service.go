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
