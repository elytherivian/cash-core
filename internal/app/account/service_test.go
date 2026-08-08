package account

import (
	"context"
	"errors"
	"testing"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type accountRepositoryStub struct {
	created      *Account
	listedUserID uuid.UUID
	accounts     []Account
}

func (r *accountRepositoryStub) Create(_ context.Context, account *Account) error {
	r.created = account
	return nil
}

func (r *accountRepositoryStub) ListActiveAccountsByUserID(_ context.Context, userID uuid.UUID) ([]Account, error) {
	r.listedUserID = userID
	return r.accounts, nil
}

func TestCreateSupportsMultipleNamedAccountsOfSameType(t *testing.T) {
	userID := uuid.New()
	repository := new(accountRepositoryStub)
	createdAccount, err := NewService(repository).Create(context.Background(), userID, CreateAccountRequest{
		AccountType: " WeChat ", AccountName: " wechat1 ",
	})

	if err != nil {
		t.Fatalf("Create(): %v", err)
	}
	if repository.created != createdAccount || createdAccount.UserID != userID || createdAccount.AccountType != AccountTypeWeChat || createdAccount.AccountName != "wechat1" {
		t.Fatalf("created account = %+v", createdAccount)
	}
}

func TestCreateRequiresAccountName(t *testing.T) {
	repository := new(accountRepositoryStub)
	_, err := NewService(repository).Create(context.Background(), uuid.New(), CreateAccountRequest{
		AccountType: AccountTypeWeChat,
	})

	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("Create() error = %v; want ErrInvalidInput", err)
	}
	if repository.created != nil {
		t.Fatal("repository Create() called without account_name")
	}
}

func TestCreateRejectsUnsupportedAccountType(t *testing.T) {
	repository := new(accountRepositoryStub)
	_, err := NewService(repository).Create(context.Background(), uuid.New(), CreateAccountRequest{
		AccountType: "PayPal",
		AccountName: "paypal1",
	})

	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("Create() error = %v; want ErrInvalidInput", err)
	}
	if repository.created != nil {
		t.Fatal("repository Create() called with unsupported account_type")
	}
}

func TestServiceListAccountsUsesUserID(t *testing.T) {
	userID := uuid.New()
	expectedAccounts := []Account{{ID: uuid.New(), UserID: userID, AccountType: AccountTypeBOC, AccountName: "boc1"}}
	repository := &accountRepositoryStub{accounts: expectedAccounts}

	accounts, err := NewService(repository).ListAccounts(context.Background(), userID)

	if err != nil {
		t.Fatalf("ListAccounts(): %v", err)
	}
	if repository.listedUserID != userID || len(accounts) != 1 || accounts[0].ID != expectedAccounts[0].ID {
		t.Fatalf("listed user ID = %s, accounts = %+v", repository.listedUserID, accounts)
	}
}
