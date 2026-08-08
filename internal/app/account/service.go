package account

import (
	"context"
	"fmt"
	"unicode/utf8"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*Account, error)
	ListAccounts(ctx context.Context, userID uuid.UUID) ([]Account, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*Account, error) {
	req.Normalize()
	if !req.AccountType.IsValid() {
		return nil, fmt.Errorf("%w: account_type must be one of WeChat, AliPay, BOC", common.ErrInvalidInput)
	}
	if length := utf8.RuneCountInString(req.AccountName); length < 1 || length > 100 {
		return nil, fmt.Errorf("%w: account_name length must be between 1 and 100", common.ErrInvalidInput)
	}
	account := &Account{
		ID: uuid.New(), UserID: userID, AccountType: req.AccountType, AccountName: req.AccountName,
		InitialBalance: req.InitialBalance, Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *service) ListAccounts(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	return s.repository.ListActiveAccountsByUserID(ctx, userID)
}
