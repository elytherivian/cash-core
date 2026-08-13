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

func TestRepositoryListTransactionsByTimeRangeReturnsOnlyOwnedActiveMatches(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "transaction.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&Transaction{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	userID, otherUserID := uuid.New(), uuid.New()
	start := time.Date(2026, time.August, 9, 0, 0, 0, 0, time.UTC)
	matched := Transaction{ID: uuid.New(), UserID: userID, Type: Expense, Amount: decimal.NewFromInt(1), AccountID: uuid.New(), CategoryID: uuid.New(), OccurredAt: start, Lifecycle: common.Lifecycle{IsActive: true}}
	transactions := []Transaction{
		matched,
		{ID: uuid.New(), UserID: userID, Type: Expense, Amount: decimal.NewFromInt(1), AccountID: uuid.New(), CategoryID: uuid.New(), OccurredAt: start.Add(24 * time.Hour), Lifecycle: common.Lifecycle{IsActive: true}},
		{ID: uuid.New(), UserID: otherUserID, Type: Expense, Amount: decimal.NewFromInt(1), AccountID: uuid.New(), CategoryID: uuid.New(), OccurredAt: start.Add(time.Hour), Lifecycle: common.Lifecycle{IsActive: true}},
		{ID: uuid.New(), UserID: userID, Type: Expense, Amount: decimal.NewFromInt(1), AccountID: uuid.New(), CategoryID: uuid.New(), OccurredAt: start.Add(time.Hour), Lifecycle: common.Lifecycle{IsActive: false}},
	}
	if err := db.Create(&transactions).Error; err != nil {
		t.Fatalf("create transactions: %v", err)
	}
	if err := db.Model(&Transaction{}).Where("id = ?", transactions[3].ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("deactivate transaction: %v", err)
	}

	listed, err := NewRepository(db).ListTransactionsByTimeRange(context.Background(), userID, ListTransactionsByTimeRangeRequest{
		StartTimestamp: start,
		EndTimestamp:   start.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("ListTransactionsByTimeRange(): %v", err)
	}
	if len(listed) != 2 || listed[0].ID != matched.ID || !listed[1].OccurredAt.Equal(start.Add(24*time.Hour)) {
		t.Fatalf("listed transactions = %+v", listed)
	}
}
