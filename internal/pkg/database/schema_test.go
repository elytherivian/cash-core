package database

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"cash-core/internal/config"
)

func TestInitializeSchemaCreatesSQLiteSchemaAndEnforcesForeignKeys(t *testing.T) {
	cfg := config.Database{
		Path: filepath.Join(t.TempDir(), "data", "cash.db"),
	}
	connection, err := Open(context.Background(), cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), "error")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = connection.Close() })

	if err := InitializeSchema(connection.GORM); err != nil {
		t.Fatalf("InitializeSchema(): %v", err)
	}
	if err := InitializeSchema(connection.GORM); err != nil {
		t.Fatalf("InitializeSchema() must be idempotent: %v", err)
	}

	var tableCount int
	if err := connection.GORM.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('users', 'accounts', 'categories', 'transactions')").Scan(&tableCount).Error; err != nil {
		t.Fatalf("query schema tables: %v", err)
	}
	if tableCount != 4 {
		t.Fatalf("schema table count=%d, want 4", tableCount)
	}
	if err := connection.GORM.Exec("INSERT INTO accounts (id, user_id, account_type, account_name) VALUES (?, ?, ?, ?)", "account-1", "missing-user", "WeChat", "wallet").Error; err == nil {
		t.Fatal("expected foreign-key constraint violation")
	}
}
