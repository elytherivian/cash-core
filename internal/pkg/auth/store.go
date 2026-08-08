package auth

import (
	"context"
	"errors"
	"time"

	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMUserStateStore struct {
	db *gorm.DB
}

func NewGORMUserStateStore(db *gorm.DB) *GORMUserStateStore {
	return &GORMUserStateStore{db: db}
}

func (s *GORMUserStateStore) AuthenticationState(ctx context.Context, userID uuid.UUID) (int64, bool, error) {
	if s.db == nil {
		return 0, false, errors.New("authentication database is not configured")
	}
	var state struct {
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		IsActive  bool       `gorm:"column:is_active"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
	}
	err := s.db.WithContext(ctx).
		Table("users").
		Select("updated_at", "is_active", "deleted_at").
		Where("id = ?", userID).
		Take(&state).Error
	if err != nil {
		return 0, false, database.NormalizeError(err)
	}
	return state.UpdatedAt.UTC().UnixMicro(), state.IsActive && state.DeletedAt == nil, nil
}
