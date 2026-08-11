package integration

import (
	"sync"
	"testing"

	"cash-core/internal/app/account"
	"cash-core/internal/app/category"
	transactionapp "cash-core/internal/app/transaction"
	"cash-core/internal/app/user"

	"gorm.io/gorm/schema"
)

func TestGORMModelsMatchSchemaTables(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		table     string
		required  []string
		forbidden []string
	}{
		{name: "user", value: &user.User{}, table: "users", required: []string{"id", "username", "password_hash"}},
		{name: "account", value: &account.Account{}, table: "accounts", required: []string{"id", "user_id", "account_type", "account_name", "initial_balance"}},
		{name: "category", value: &category.Category{}, table: "categories", required: []string{"id", "user_id", "category_name"}, forbidden: []string{"category_type", "type"}},
		{name: "transaction", value: &transactionapp.Transaction{}, table: "transactions", required: []string{"id", "user_id", "type", "amount", "account_id", "category_id", "occurred_at"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := schema.Parse(test.value, &sync.Map{}, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parse GORM schema: %v", err)
			}
			if parsed.Table != test.table {
				t.Fatalf("table=%q, want %q", parsed.Table, test.table)
			}
			required := append(test.required, "created_at", "updated_at", "is_active", "deleted_at")
			for _, field := range required {
				if parsed.LookUpField(field) == nil {
					t.Errorf("GORM schema is missing %q", field)
				}
			}
			for _, field := range test.forbidden {
				if parsed.LookUpField(field) != nil {
					t.Errorf("GORM schema still contains %q", field)
				}
			}
		})
	}
}
