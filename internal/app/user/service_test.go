package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

type conflictRepository struct{}

func (conflictRepository) Create(context.Context, *User) error {
	return common.ErrConflict
}

func (conflictRepository) FindByID(context.Context, uuid.UUID) (*User, error) {
	return nil, common.ErrNotFound
}

func (conflictRepository) Delete(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func TestCreateMapsDuplicateUserToBusinessCode(t *testing.T) {
	service := NewService(conflictRepository{})
	_, err := service.Create(context.Background(), CreateRequest{
		Username: "existing-user",
		Password: "password123",
	})

	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("error = %v, want conflict", err)
	}
	code, ok := common.BusinessCode(err)
	if !ok || code != common.CodeRegisterUserAlreadyExists {
		t.Fatalf("business code = %d, %t; want %d, true", code, ok, common.CodeRegisterUserAlreadyExists)
	}
}
