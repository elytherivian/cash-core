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
	updatedTransaction *Transaction
	updateRequest      UpdateTransactionRequest
	updatedUserID      uuid.UUID
	updatedID          uuid.UUID
	listRequest        ListTransactionsRequest
	listedUserID       uuid.UUID
}

func (r *serviceRepositoryStub) CreateTransaction(_ context.Context, transaction *Transaction) error {
	r.createdTransaction = transaction
	return nil
}

func (r *serviceRepositoryStub) UpdateTransaction(
	_ context.Context,
	userID, transactionID uuid.UUID,
	request UpdateTransactionRequest,
) (*Transaction, error) {
	r.updatedUserID = userID
	r.updatedID = transactionID
	r.updateRequest = request
	r.updatedTransaction = &Transaction{ID: transactionID, UserID: userID}
	return r.updatedTransaction, nil
}

func (r *serviceRepositoryStub) ListTransactions(_ context.Context, userID uuid.UUID, request ListTransactionsRequest) ([]Transaction, error) {
	r.listedUserID = userID
	r.listRequest = request
	return []Transaction{}, nil
}

func (r *serviceRepositoryStub) ListTransactionsByTimeRange(_ context.Context, userID uuid.UUID, request ListTransactionsByTimeRangeRequest) ([]Transaction, error) {
	r.listedUserID = userID
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

func TestListTransactionsByTimeRangeRejectsInvalidRange(t *testing.T) {
	service := NewService(new(serviceRepositoryStub))
	start := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)
	end := start.Add(-time.Second)
	if _, err := service.ListTransactionsByTimeRange(context.Background(), uuid.New(), ListTransactionsByTimeRangeRequest{
		StartTimestamp: start,
		EndTimestamp:   end,
	}); err == nil {
		t.Fatal("ListTransactionsByTimeRange() error = nil, want invalid input")
	}
}

func TestUpdateTransactionNormalizesAndPassesOnlyRequestedFields(t *testing.T) {
	repository := new(serviceRepositoryStub)
	userID, transactionID := uuid.New(), uuid.New()
	transactionType := TransactionType(" EXPENSE ")
	amount := decimal.NewFromInt(25)
	occurredAt := time.Date(2026, time.August, 10, 20, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))

	updatedTransaction, err := NewService(repository).UpdateTransaction(context.Background(), userID, transactionID, UpdateTransactionRequest{
		Type: &transactionType, Amount: &amount, OccurredAt: &occurredAt,
	})
	if err != nil {
		t.Fatalf("UpdateTransaction(): %v", err)
	}
	if updatedTransaction != repository.updatedTransaction || repository.updatedUserID != userID || repository.updatedID != transactionID {
		t.Fatalf("updated transaction = %+v, user ID = %s, transaction ID = %s", updatedTransaction, repository.updatedUserID, repository.updatedID)
	}
	if repository.updateRequest.Type == nil || *repository.updateRequest.Type != Expense ||
		repository.updateRequest.Amount == nil || repository.updateRequest.Amount.Cmp(decimal.NewFromInt(25)) != 0 ||
		repository.updateRequest.OccurredAt == nil || !repository.updateRequest.OccurredAt.Equal(occurredAt.UTC()) ||
		repository.updateRequest.OccurredAt.Location() != time.UTC ||
		repository.updateRequest.AccountID != nil || repository.updateRequest.CategoryID != nil {
		t.Fatalf("update request = %+v", repository.updateRequest)
	}
}

func TestUpdateTransactionRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateTransactionRequest
	}{
		{name: "empty request", request: UpdateTransactionRequest{}},
		{name: "invalid type", request: UpdateTransactionRequest{Type: transactionTypePointer("transfer")}},
		{name: "non-positive amount", request: UpdateTransactionRequest{Amount: decimalPointer(decimal.Zero)}},
		{name: "empty account ID", request: UpdateTransactionRequest{AccountID: uuidPointer(uuid.Nil)}},
		{name: "empty category ID", request: UpdateTransactionRequest{CategoryID: uuidPointer(uuid.Nil)}},
		{name: "zero occurred at", request: UpdateTransactionRequest{OccurredAt: timePointer(time.Time{})}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewService(new(serviceRepositoryStub)).UpdateTransaction(
				context.Background(), uuid.New(), uuid.New(), test.request,
			); err == nil {
				t.Fatal("UpdateTransaction() error = nil, want invalid input")
			}
		})
	}
}

func transactionTypePointer(value TransactionType) *TransactionType { return &value }
func decimalPointer(value decimal.Decimal) *decimal.Decimal         { return &value }
func uuidPointer(value uuid.UUID) *uuid.UUID                        { return &value }
func timePointer(value time.Time) *time.Time                        { return &value }
