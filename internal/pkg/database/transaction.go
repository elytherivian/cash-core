package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// WithinTransaction runs operation in a GORM transaction. Use this from an
// orchestration service when one use case must update multiple repositories.
func WithinTransaction(ctx context.Context, db *gorm.DB, operation func(tx *gorm.DB) error) error {
	if err := db.WithContext(ctx).Transaction(operation); err != nil {
		return fmt.Errorf("execute database transaction: %w", err)
	}
	return nil
}
