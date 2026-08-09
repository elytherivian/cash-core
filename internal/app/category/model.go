package category

import (
	"strings"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type Category struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	CategoryName string    `gorm:"column:category_name;size:80;not null" json:"category_name"`
	common.Lifecycle
}

func (Category) TableName() string { return "categories" }

type CategoryResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	CategoryName string    `json:"category_name"`
	common.LifecycleResponse
}

func (c Category) Response(location *time.Location) CategoryResponse {
	return CategoryResponse{
		ID: c.ID, UserID: c.UserID, CategoryName: c.CategoryName, LifecycleResponse: c.Lifecycle.Response(location),
	}
}

type CreateCategoryRequest struct {
	CategoryName string `json:"category_name"`
}

func (r *CreateCategoryRequest) Normalize() {
	r.CategoryName = strings.TrimSpace(r.CategoryName)
}
