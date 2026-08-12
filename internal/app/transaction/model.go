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

type TransactionResponse struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	Type       TransactionType `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	AccountID  uuid.UUID       `json:"account_id"`
	CategoryID uuid.UUID       `json:"category_id"`
	OccurredAt time.Time       `json:"occurred_at"`
	common.LifecycleResponse
}

func (t Transaction) Response(location *time.Location) TransactionResponse {
	if location == nil {
		location = time.UTC
	}
	return TransactionResponse{
		ID: t.ID, UserID: t.UserID, Type: t.Type, Amount: t.Amount, AccountID: t.AccountID,
		CategoryID: t.CategoryID, OccurredAt: t.OccurredAt.In(location), LifecycleResponse: t.Lifecycle.Response(location),
	}
}

type CreateTransactionRequest struct {
	Type       TransactionType `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	AccountID  uuid.UUID       `json:"account_id"`
	CategoryID uuid.UUID       `json:"category_id"`
	OccurredAt time.Time       `json:"occurred_at"`
}

func (r *CreateTransactionRequest) Normalize() {
	r.Type = TransactionType(strings.ToLower(strings.TrimSpace(string(r.Type))))
	r.OccurredAt = r.OccurredAt.UTC()
}

// UpdateTransactionRequest uses pointers to distinguish an omitted field from
// a field whose zero value was explicitly submitted. PATCH only changes fields
// present in the request body.
type UpdateTransactionRequest struct {
	ID         string           `json:"id"`
	Type       *TransactionType `json:"type"`
	Amount     *decimal.Decimal `json:"amount"`
	AccountID  *uuid.UUID       `json:"account_id"`
	CategoryID *uuid.UUID       `json:"category_id"`
	OccurredAt *time.Time       `json:"occurred_at"`
}

func (r *UpdateTransactionRequest) Normalize() {
	if r.Type != nil {
		normalizedType := TransactionType(strings.ToLower(strings.TrimSpace(string(*r.Type))))
		r.Type = &normalizedType
	}
	if r.OccurredAt != nil {
		occurredAt := r.OccurredAt.UTC()
		r.OccurredAt = &occurredAt
	}
}

func (r UpdateTransactionRequest) HasChanges() bool {
	return r.Type != nil || r.Amount != nil || r.AccountID != nil || r.CategoryID != nil || r.OccurredAt != nil
}

type ListTransactionsRequest struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
}
