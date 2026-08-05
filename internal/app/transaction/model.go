package transaction

import (
	"strings"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

func (t TransactionType) Valid() bool { return t == Income || t == Expense }

type Transaction struct {
	ID         uuid.UUID       `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID       `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Type       TransactionType `gorm:"column:type;size:20;not null" json:"type"`
	Amount     decimal.Decimal `gorm:"column:amount;type:numeric(19,4);not null" json:"amount"`
	AccountID  uuid.UUID       `gorm:"column:account_id;type:uuid;not null" json:"account_id"`
	CategoryID uuid.UUID       `gorm:"column:category_id;type:uuid;not null" json:"category_id"`
	OccurredAt time.Time       `gorm:"column:occurred_at;not null" json:"occurred_at"`
	common.Lifecycle
}

func (Transaction) TableName() string { return "transactions" }

type CreateRequest struct {
	Type       TransactionType `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	AccountID  uuid.UUID       `json:"account_id"`
	CategoryID uuid.UUID       `json:"category_id"`
	OccurredAt time.Time       `json:"occurred_at"`
}

func (r *CreateRequest) Normalize() {
	r.Type = TransactionType(strings.ToLower(strings.TrimSpace(string(r.Type))))
	r.OccurredAt = r.OccurredAt.UTC()
}

type Filter struct {
	From *time.Time
	To   *time.Time
	Page common.Page
}
