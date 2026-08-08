package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/config"

	"github.com/google/uuid"
)

type stateStore struct {
	version int64
	active  bool
	err     error
}

func (s *stateStore) AuthenticationState(context.Context, uuid.UUID) (int64, bool, error) {
	return s.version, s.active, s.err
}

func testManager(store *stateStore) *Manager {
	manager := NewManager(config.Auth{
		Issuer: "test", JWTSecret: "test-jwt-secret-that-is-at-least-32-bytes",
		AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 30 * 24 * time.Hour,
	}, store)
	manager.now = func() time.Time {
		return time.Date(2026, time.August, 8, 10, 0, 0, 0, time.UTC)
	}
	return manager
}

func TestManagerIssuesAndVerifiesAccessToken(t *testing.T) {
	userID := uuid.New()
	store := &stateStore{version: 123, active: true}
	manager := testManager(store)
	pair, err := manager.Issue(userID, store.version)
	if err != nil {
		t.Fatalf("Issue(): %v", err)
	}

	got, err := manager.Verify(context.Background(), pair.AccessToken)
	if err != nil || got != userID {
		t.Fatalf("Verify() = %s, %v; want %s, nil", got, err, userID)
	}
	if pair.TokenType != "Bearer" || pair.ExpiresIn != 900 || pair.RefreshExpiresIn != 2592000 {
		t.Fatalf("unexpected token metadata: %+v", pair)
	}
	if pair.UserID != userID {
		t.Fatalf("token user ID = %s, want %s", pair.UserID, userID)
	}
}

func TestManagerRefreshesOnlyWithRefreshToken(t *testing.T) {
	store := &stateStore{version: 123, active: true}
	manager := testManager(store)
	pair, err := manager.Issue(uuid.New(), store.version)
	if err != nil {
		t.Fatalf("Issue(): %v", err)
	}
	if _, err := manager.Refresh(context.Background(), pair.AccessToken); !errors.Is(err, common.ErrUnauthenticated) {
		t.Fatalf("Refresh(access token) error = %v; want ErrUnauthenticated", err)
	}
	if _, err := manager.Refresh(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Refresh(refresh token): %v", err)
	}
}

func TestManagerRejectsInactiveOrChangedUser(t *testing.T) {
	store := &stateStore{version: 123, active: true}
	manager := testManager(store)
	pair, err := manager.Issue(uuid.New(), store.version)
	if err != nil {
		t.Fatalf("Issue(): %v", err)
	}

	store.active = false
	if _, err := manager.Verify(context.Background(), pair.AccessToken); !errors.Is(err, common.ErrUnauthenticated) {
		t.Fatalf("Verify(inactive) error = %v; want ErrUnauthenticated", err)
	}
	store.active = true
	store.version++
	if _, err := manager.Verify(context.Background(), pair.AccessToken); !errors.Is(err, common.ErrUnauthenticated) {
		t.Fatalf("Verify(changed version) error = %v; want ErrUnauthenticated", err)
	}
}
