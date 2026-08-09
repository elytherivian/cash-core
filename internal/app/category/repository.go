package category

import (
	"context"

	"cash-core/internal/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateCategory(ctx context.Context, category *Category) error
	ListCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]Category, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateCategory(ctx context.Context, category *Category) error {
	return database.NormalizeError(r.db.WithContext(ctx).Create(category).Error)
}

func (r *repository) ListCategoriesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]Category, error) {
	query := r.db.WithContext(ctx).Model(&Category{}).
		Where("user_id = ? AND is_active = TRUE", userID)

	categories := make([]Category, 0)
	err := query.Order("created_at DESC").Find(&categories).Error
	return categories, database.NormalizeError(err)
}
