package account

import (
	"strings"

	"cash-core/internal/common"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID             uuid.UUID       `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID       `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	AccountType    AccountType     `gorm:"column:account_type;size:100;not null" json:"account_type"`
	AccountName    string          `gorm:"column:account_name;size:100;not null" json:"account_name"`
	InitialBalance decimal.Decimal `gorm:"column:initial_balance;type:numeric(19,4);not null" json:"initial_balance"`
	common.Lifecycle
}

func (Account) TableName() string { return "accounts" }

type CreateAccountRequest struct {
	AccountType    AccountType     `json:"account_type"`
	AccountName    string          `json:"account_name"`
	InitialBalance decimal.Decimal `json:"initial_balance"`
}

func (r *CreateAccountRequest) Normalize() {
	r.AccountType = AccountType(strings.TrimSpace(string(r.AccountType)))
	r.AccountName = strings.TrimSpace(r.AccountName)
}
