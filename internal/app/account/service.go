package account

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, request CreateRequest) (*Account, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Account, error)
	List(ctx context.Context, userID uuid.UUID, page common.Page) ([]Account, int64, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, request CreateRequest) (*Account, error) {
	request.Normalize()
	if length := utf8.RuneCountInString(request.AccountType); length < 1 || length > 100 {
		return nil, fmt.Errorf("%w: account_type length must be between 1 and 100", common.ErrInvalidInput)
	}
	value := &Account{
		ID: uuid.New(), UserID: userID, AccountType: request.AccountType,
		InitialBalance: request.InitialBalance, Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (*Account, error) {
	return s.repository.FindByID(ctx, userID, id)
}

func (s *service) List(ctx context.Context, userID uuid.UUID, page common.Page) ([]Account, int64, error) {
	return s.repository.ListByUser(ctx, userID, page)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repository.Delete(ctx, userID, id, time.Now().UTC())
}
