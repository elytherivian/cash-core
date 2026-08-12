package transaction

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryUpdateTransactionChangesRequestedFieldsForOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "transaction.db")), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&Transaction{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	userID, transactionID := uuid.New(), uuid.New()
	original := Transaction{
		ID: transactionID, UserID: userID, Type: Expense, Amount: decimal.NewFromInt(20),
		AccountID: uuid.New(), CategoryID: uuid.New(), OccurredAt: time.Now().UTC(),
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := db.Create(&original).Error; err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	newAmount := decimal.NewFromInt(25)
	updated, err := NewRepository(db).UpdateTransaction(context.Background(), userID, transactionID, UpdateTransactionRequest{
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("UpdateTransaction(): %v", err)
	}
	if updated.Amount.Cmp(newAmount) != 0 || updated.Type != original.Type ||
		updated.AccountID != original.AccountID || updated.CategoryID != original.CategoryID ||
		!updated.OccurredAt.Equal(original.OccurredAt) {
		t.Fatalf("updated transaction = %+v", updated)
	}

	otherAmount := decimal.NewFromInt(30)
	_, err = NewRepository(db).UpdateTransaction(context.Background(), uuid.New(), transactionID, UpdateTransactionRequest{
		Amount: &otherAmount,
	})
	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("other user's UpdateTransaction() error = %v, want ErrNotFound", err)
	}
}
