package category

import (
	"context"
	"fmt"
	"unicode/utf8"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Service interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, request CreateCategoryRequest) (*Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]Category, error)
}

type service struct{ repository Repository }

func NewService(repository Repository) Service { return &service{repository: repository} }

func (s *service) CreateCategory(ctx context.Context, userID uuid.UUID, request CreateCategoryRequest) (*Category, error) {
	request.Normalize()
	if length := utf8.RuneCountInString(request.CategoryName); length < 1 || length > 80 {
		return nil, fmt.Errorf("%w: category_name length must be between 1 and 80", common.ErrInvalidInput)
	}
	category := &Category{
		ID: uuid.New(), UserID: userID, CategoryName: request.CategoryName,
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *service) ListCategories(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	return s.repository.ListCategoriesByUserID(ctx, userID)
}
