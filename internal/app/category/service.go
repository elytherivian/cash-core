package category

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	if !request.CategoryType.Valid() {
		return nil, fmt.Errorf("%w: category_type must be income or expense", common.ErrInvalidInput)
	}
	category := &Category{
		ID: uuid.New(), UserID: userID, CategoryType: request.CategoryType,
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *service) Get(ctx context.Context, userID, id uuid.UUID) (*Category, error) {
	return s.repository.FindByID(ctx, userID, id)
}

func (s *service) List(ctx context.Context, userID uuid.UUID, categoryTypeValue string, page common.Page) ([]Category, int64, error) {
	var categoryType *CategoryType
	categoryTypeValue = strings.ToLower(strings.TrimSpace(categoryTypeValue))
	if categoryTypeValue != "" {
		parsed := CategoryType(categoryTypeValue)
		if !parsed.Valid() {
			return nil, 0, fmt.Errorf("%w: category_type must be income or expense", common.ErrInvalidInput)
		}
		categoryType = &parsed
	}
	return s.repository.ListByUser(ctx, userID, categoryType, page)
}

func (s *service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repository.Delete(ctx, userID, id, time.Now().UTC())
}
