package category

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type serviceRepositoryStub struct {
	createdCategory *Category
}

func (r *serviceRepositoryStub) CreateCategory(_ context.Context, category *Category) error {
	r.createdCategory = category
	return nil
}

func (r *serviceRepositoryStub) ListCategoriesByUserID(context.Context, uuid.UUID) ([]Category, error) {
	return nil, nil
}

func TestCreateCategoryStoresNormalizedCategoryName(t *testing.T) {
	repository := new(serviceRepositoryStub)
	createdCategory, err := NewService(repository).CreateCategory(context.Background(), uuid.New(), CreateCategoryRequest{
		CategoryName: " 日用品 ",
	})
	if err != nil {
		t.Fatalf("CreateCategory(): %v", err)
	}
	if createdCategory != repository.createdCategory || createdCategory.CategoryName != "日用品" {
		t.Fatalf("created category = %+v", createdCategory)
	}
}
