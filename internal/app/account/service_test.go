package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type createRepository struct {
	created *Account
}

func (r *createRepository) Create(_ context.Context, value *Account) error {
	r.created = value
	return nil
}

func (r *createRepository) FindByID(context.Context, uuid.UUID, uuid.UUID) (*Account, error) {
	return nil, common.ErrNotFound
}

func (r *createRepository) ListByUser(context.Context, uuid.UUID, common.Page) ([]Account, int64, error) {
	return nil, 0, nil
}

func (r *createRepository) Delete(context.Context, uuid.UUID, uuid.UUID, time.Time) error {
	return nil
}

func TestCreateSupportsMultipleNamedAccountsOfSameType(t *testing.T) {
	userID := uuid.New()
	repository := new(createRepository)
	value, err := NewService(repository).Create(context.Background(), userID, CreateRequest{
		AccountType: " wechat ", AccountName: " wechat1 ",
	})

	if err != nil {
		t.Fatalf("Create(): %v", err)
	}
	if repository.created != value || value.UserID != userID || value.AccountType != "wechat" || value.AccountName != "wechat1" {
		t.Fatalf("created account = %+v", value)
	}
}

func TestCreateRequiresAccountName(t *testing.T) {
	repository := new(createRepository)
	_, err := NewService(repository).Create(context.Background(), uuid.New(), CreateRequest{
		AccountType: "wechat",
	})

	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("Create() error = %v; want ErrInvalidInput", err)
	}
	if repository.created != nil {
		t.Fatal("repository Create() called without account_name")
	}
}
