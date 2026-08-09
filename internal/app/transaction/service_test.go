package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type serviceRepositoryStub struct {
	createdTransaction *Transaction
	listRequest        ListTransactionsRequest
	listedUserID       uuid.UUID
}

func (r *serviceRepositoryStub) CreateTransaction(_ context.Context, transaction *Transaction) error {
	r.createdTransaction = transaction
	return nil
}

func (r *serviceRepositoryStub) ListTransactions(_ context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error) {
	r.listedUserID = userID
	r.listRequest = request
	return []Transaction{}, nil
}

func TestCreateTransactionStoresAccountAndCategory(t *testing.T) {
	repository := new(serviceRepositoryStub)
	accountID := uuid.New()
	categoryID := uuid.New()
	createdTransaction, err := NewService(repository).CreateTransaction(context.Background(), uuid.New(), CreateTransactionRequest{
		Type:       " EXPENSE ",
		Amount:     decimal.NewFromInt(20),
		AccountID:  accountID,
		CategoryID: categoryID,
		OccurredAt: time.Date(2026, time.August, 9, 20, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60)),
	})
	if err != nil {
		t.Fatalf("CreateTransaction(): %v", err)
	}
	if createdTransaction != repository.createdTransaction {
		t.Fatal("created transaction was not sent to the repository")
	}
	if createdTransaction.Type != Expense || createdTransaction.Amount.Cmp(decimal.NewFromInt(20)) != 0 ||
		createdTransaction.AccountID != accountID || createdTransaction.CategoryID != categoryID ||
		createdTransaction.OccurredAt.Location() != time.UTC {
		t.Fatalf("created transaction = %+v", createdTransaction)
	}
}

func TestCreateTransactionUsesCurrentTimeWhenOccurredAtIsOmitted(t *testing.T) {
	repository := new(serviceRepositoryStub)
	before := time.Now().UTC()
	createdTransaction, err := NewService(repository).CreateTransaction(context.Background(), uuid.New(), CreateTransactionRequest{
		Type:       Expense,
		Amount:     decimal.NewFromInt(20),
		AccountID:  uuid.New(),
		CategoryID: uuid.New(),
	})
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("CreateTransaction(): %v", err)
	}
	if createdTransaction.OccurredAt.Before(before) || createdTransaction.OccurredAt.After(after) {
		t.Fatalf("OccurredAt = %s, want a time between %s and %s", createdTransaction.OccurredAt, before, after)
	}
}

func TestListTransactionsRequiresAtLeastOneFilter(t *testing.T) {
	if _, err := NewService(new(serviceRepositoryStub)).ListTransactions(context.Background(), uuid.New(), ListTransactionsRequest{}); err == nil {
		t.Fatal("ListTransactions() error = nil, want invalid input")
	}
}
