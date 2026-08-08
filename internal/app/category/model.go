package category

import (
	"strings"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type CategoryType string

const (
	Income  CategoryType = "income"
	Expense CategoryType = "expense"
)

func (t CategoryType) Valid() bool { return t == Income || t == Expense }

type Category struct {
	ID           uuid.UUID    `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID    `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	CategoryType CategoryType `gorm:"column:category_type;size:20;not null" json:"category_type"`
	common.Lifecycle
}

func (Category) TableName() string { return "categories" }

type CreateRequest struct {
	CategoryType CategoryType `json:"category_type"`
}

func (r *CreateRequest) Normalize() {
	r.CategoryType = CategoryType(strings.ToLower(strings.TrimSpace(string(r.CategoryType))))
}
