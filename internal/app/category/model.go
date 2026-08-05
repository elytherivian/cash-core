package category

import (
	"strings"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

func (t TransactionType) Valid() bool { return t == Income || t == Expense }

type Category struct {
	ID           uuid.UUID       `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID       `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	CategoryName string          `gorm:"column:category_name;size:80;not null" json:"category_name"`
	Type         TransactionType `gorm:"column:type;size:20;not null" json:"type"`
	common.Lifecycle
}

func (Category) TableName() string { return "categories" }

type CreateRequest struct {
	CategoryName string          `json:"category_name"`
	Type         TransactionType `json:"type"`
}

func (r *CreateRequest) Normalize() {
	r.CategoryName = strings.TrimSpace(r.CategoryName)
	r.Type = TransactionType(strings.ToLower(strings.TrimSpace(string(r.Type))))
}
