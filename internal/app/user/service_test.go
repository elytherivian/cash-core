package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/pkg/auth"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type conflictRepository struct{}

type tokenService struct {
	pair          auth.TokenPair
	err           error
	issuedUserID  uuid.UUID
	issuedVersion int64
	refreshToken  string
}

func (s *tokenService) Issue(userID uuid.UUID, version int64) (auth.TokenPair, error) {
	s.issuedUserID = userID
	s.issuedVersion = version
	return s.pair, s.err
}

func (s *tokenService) Refresh(_ context.Context, refreshToken string) (auth.TokenPair, error) {
	s.refreshToken = refreshToken
	return s.pair, s.err
}

func (conflictRepository) Create(context.Context, *User) error {
	return common.ErrConflict
}

func (conflictRepository) FindByID(context.Context, uuid.UUID) (*User, error) {
	return nil, common.ErrNotFound
}

func (conflictRepository) FindByUsername(context.Context, string) (*User, error) {
	return nil, common.ErrNotFound
}

func (conflictRepository) FindByUsernameIncludingDeleted(context.Context, string) (*User, error) {
	return nil, common.ErrNotFound
}

func (conflictRepository) Delete(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func (conflictRepository) Restore(context.Context, uuid.UUID, time.Time) error {
	return nil
}

type deleteRepository struct {
	user         *User
	findErr      error
	deleteErr    error
	deleteCalled bool
	deletedID    uuid.UUID
	deletedAt    time.Time
}

func (r *deleteRepository) Create(context.Context, *User) error { return nil }

func (r *deleteRepository) FindByID(context.Context, uuid.UUID) (*User, error) {
	return nil, common.ErrNotFound
}

func (r *deleteRepository) FindByUsername(context.Context, string) (*User, error) {
	return r.user, r.findErr
}

func (r *deleteRepository) FindByUsernameIncludingDeleted(context.Context, string) (*User, error) {
	return r.user, r.findErr
}

func (r *deleteRepository) Delete(_ context.Context, id uuid.UUID, deletedAt time.Time) error {
	r.deleteCalled = true
	r.deletedID = id
	r.deletedAt = deletedAt
	return r.deleteErr
}

func (r *deleteRepository) Restore(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func TestDeleteVerifiesPasswordAndSoftDeletesUser(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	userID := uuid.New()
	repository := &deleteRepository{user: &User{ID: userID, PasswordHash: string(passwordHash)}}

	err = NewService(repository, nil).Delete(context.Background(), DeleteUserRequest{
		Username: " user ",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("Delete(): %v", err)
	}
	if !repository.deleteCalled || repository.deletedID != userID {
		t.Fatalf("Delete() called = %t, id = %s; want true, %s", repository.deleteCalled, repository.deletedID, userID)
	}
	if repository.deletedAt.IsZero() || repository.deletedAt.Location() != time.UTC {
		t.Fatalf("deletedAt = %v; want a non-zero UTC time", repository.deletedAt)
	}
}

func TestDeleteRejectsIncorrectPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repository := &deleteRepository{user: &User{ID: uuid.New(), PasswordHash: string(passwordHash)}}

	err = NewService(repository, nil).Delete(context.Background(), DeleteUserRequest{
		Username: "user",
		Password: "incorrect",
	})

	if !errors.Is(err, common.ErrUnauthenticated) {
		t.Fatalf("Delete() error = %v; want ErrUnauthenticated", err)
	}
	if repository.deleteCalled {
		t.Fatal("repository Delete() called for an incorrect password")
	}
}

func TestDeleteRejectsInvalidRequestBeforeQuery(t *testing.T) {
	repository := &deleteRepository{}
	err := NewService(repository, nil).Delete(context.Background(), DeleteUserRequest{})

	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("Delete() error = %v; want ErrInvalidInput", err)
	}
	if repository.deleteCalled {
		t.Fatal("repository Delete() called for an invalid request")
	}
}

type restoreRepository struct {
	user          *User
	findErr       error
	restoreErr    error
	restoreCalled bool
	restoredID    uuid.UUID
	restoredAt    time.Time
}

func (r *restoreRepository) Create(context.Context, *User) error { return nil }

func (r *restoreRepository) FindByID(context.Context, uuid.UUID) (*User, error) {
	return nil, common.ErrNotFound
}

func (r *restoreRepository) FindByUsername(context.Context, string) (*User, error) {
	return nil, common.ErrNotFound
}

func (r *restoreRepository) FindByUsernameIncludingDeleted(context.Context, string) (*User, error) {
	return r.user, r.findErr
}

func (r *restoreRepository) Delete(context.Context, uuid.UUID, time.Time) error { return nil }

func (r *restoreRepository) Restore(_ context.Context, id uuid.UUID, restoredAt time.Time) error {
	r.restoreCalled = true
	r.restoredID = id
	r.restoredAt = restoredAt
	return r.restoreErr
}

func TestRestoreVerifiesPasswordAndRestoresDeletedUser(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	userID := uuid.New()
	deletedAt := time.Now().UTC().Add(-time.Hour)
	repository := &restoreRepository{user: &User{
		ID:           userID,
		PasswordHash: string(passwordHash),
		Lifecycle:    common.Lifecycle{IsActive: false, DeletedAt: &deletedAt},
	}}

	err = NewService(repository, nil).Restore(context.Background(), RestoreUserRequest{
		Username: " user ",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("Restore(): %v", err)
	}
	if !repository.restoreCalled || repository.restoredID != userID {
		t.Fatalf("Restore() called = %t, id = %s; want true, %s", repository.restoreCalled, repository.restoredID, userID)
	}
	if repository.restoredAt.IsZero() || repository.restoredAt.Location() != time.UTC {
		t.Fatalf("restoredAt = %v; want a non-zero UTC time", repository.restoredAt)
	}
}

func TestRestoreRejectsMissingUser(t *testing.T) {
	repository := &restoreRepository{findErr: common.ErrNotFound}
	err := NewService(repository, nil).Restore(context.Background(), RestoreUserRequest{
		Username: "missing-user",
		Password: "password123",
	})

	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("Restore() error = %v; want ErrNotFound", err)
	}
	if repository.restoreCalled {
		t.Fatal("repository Restore() called for a missing user")
	}
}

func TestRestoreRejectsActiveUser(t *testing.T) {
	repository := &restoreRepository{user: &User{
		ID:        uuid.New(),
		Lifecycle: common.Lifecycle{IsActive: true},
	}}
	err := NewService(repository, nil).Restore(context.Background(), RestoreUserRequest{
		Username: "active-user",
		Password: "password123",
	})

	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("Restore() error = %v; want ErrConflict", err)
	}
	if repository.restoreCalled {
		t.Fatal("repository Restore() called for an active user")
	}
}

func TestRestoreRejectsIncorrectPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	deletedAt := time.Now().UTC().Add(-time.Hour)
	repository := &restoreRepository{user: &User{
		ID:           uuid.New(),
		PasswordHash: string(passwordHash),
		Lifecycle:    common.Lifecycle{IsActive: false, DeletedAt: &deletedAt},
	}}

	err = NewService(repository, nil).Restore(context.Background(), RestoreUserRequest{
		Username: "deleted-user",
		Password: "incorrect",
	})

	if !errors.Is(err, common.ErrUnauthenticated) {
		t.Fatalf("Restore() error = %v; want ErrUnauthenticated", err)
	}
	if repository.restoreCalled {
		t.Fatal("repository Restore() called for an incorrect password")
	}
}

func TestLoginVerifiesPasswordAndIssuesTokens(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	userID := uuid.New()
	updatedAt := time.Date(2026, time.August, 8, 10, 0, 0, 123456000, time.UTC)
	repository := &deleteRepository{user: &User{
		ID: userID, PasswordHash: string(passwordHash),
		Lifecycle: common.Lifecycle{IsActive: true, UpdatedAt: updatedAt},
	}}
	want := auth.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	tokens := &tokenService{pair: want}

	got, err := NewService(repository, tokens).Login(context.Background(), LoginRequest{
		Username: " user ", Password: "password123",
	})

	if err != nil {
		t.Fatalf("Login(): %v", err)
	}
	if got != want {
		t.Fatalf("Login() = %+v, want %+v", got, want)
	}
	if tokens.issuedUserID != userID || tokens.issuedVersion != updatedAt.UnixMicro() {
		t.Fatalf("issued user/version = %s/%d", tokens.issuedUserID, tokens.issuedVersion)
	}
}

func TestLoginHidesMissingUsername(t *testing.T) {
	repository := &deleteRepository{findErr: common.ErrNotFound}
	_, err := NewService(repository, &tokenService{}).Login(context.Background(), LoginRequest{
		Username: "missing", Password: "password123",
	})
	if !errors.Is(err, common.ErrUnauthenticated) {
		t.Fatalf("Login() error = %v; want ErrUnauthenticated", err)
	}
}

func TestRefreshDelegatesToTokenService(t *testing.T) {
	want := auth.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}
	tokens := &tokenService{pair: want}
	got, err := NewService(&deleteRepository{}, tokens).Refresh(context.Background(), RefreshTokenRequest{
		RefreshToken: " refresh-token ",
	})
	if err != nil {
		t.Fatalf("Refresh(): %v", err)
	}
	if got != want || tokens.refreshToken != "refresh-token" {
		t.Fatalf("Refresh() = %+v, token = %q", got, tokens.refreshToken)
	}
}
