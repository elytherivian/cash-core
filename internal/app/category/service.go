package category

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, request CreateRequest) (*Category, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	List(ctx context.Context, userID uuid.UUID, transactionType string, page common.Page) ([]Category, int64, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type service struct{ repository Repository }

func NewService(repository Repository) Service { return &service{repository: repository} }

func (s *service) Create(ctx context.Context, userID uuid.UUID, request CreateRequest) (*Category, error) {
	request.Normalize()
	if length := utf8.RuneCountInString(request.CategoryName); length < 1 || length > 80 {
		return nil, fmt.Errorf("%w: category_name length must be between 1 and 80", common.ErrInvalidInput)
	}
	if !request.Type.Valid() {
		return nil, fmt.Errorf("%w: type must be income or expense", common.ErrInvalidInput)
	}
	value := &Category{
		ID: uuid.New(), UserID: userID, CategoryName: request.CategoryName, Type: request.Type,
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (*Category, error) {
	return s.repository.FindByID(ctx, userID, id)
}

func (s *service) List(ctx context.Context, userID uuid.UUID, value string, page common.Page) ([]Category, int64, error) {
	var transactionType *TransactionType
	value = strings.ToLower(strings.TrimSpace(value))
	if value != "" {
		parsed := TransactionType(value)
		if !parsed.Valid() {
			return nil, 0, fmt.Errorf("%w: type must be income or expense", common.ErrInvalidInput)
		}
		transactionType = &parsed
	}
	return s.repository.ListByUser(ctx, userID, transactionType, page)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repository.Delete(ctx, userID, id, time.Now().UTC())
}
