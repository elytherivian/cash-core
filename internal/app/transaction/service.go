package transaction

import (
	"context"
	"fmt"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, request CreateRequest) (*Transaction, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Transaction, error)
	List(ctx context.Context, userID uuid.UUID, filter Filter) ([]Transaction, int64, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type service struct{ repository Repository }

func NewService(repository Repository) Service { return &service{repository: repository} }

func (s *service) Create(ctx context.Context, userID uuid.UUID, request CreateRequest) (*Transaction, error) {
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
		return nil, fmt.Errorf("%w: occurred_at is required", common.ErrInvalidInput)
	}
	value := &Transaction{
		ID: uuid.New(), UserID: userID, Type: request.Type, Amount: request.Amount,
		AccountID: request.AccountID, CategoryID: request.CategoryID, OccurredAt: request.OccurredAt,
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (*Transaction, error) {
	return s.repository.FindByID(ctx, userID, id)
}

func (s *service) List(ctx context.Context, userID uuid.UUID, filter Filter) ([]Transaction, int64, error) {
	if filter.From != nil && filter.To != nil && !filter.From.Before(*filter.To) {
		return nil, 0, fmt.Errorf("%w: from must be before to", common.ErrInvalidInput)
	}
	return s.repository.ListByUser(ctx, userID, filter)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repository.Delete(ctx, userID, id, time.Now().UTC())
}
