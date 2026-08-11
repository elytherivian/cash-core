package database

import (
	"fmt"

	"gorm.io/gorm"
)

// InitializeSchema creates the SQLite schema required by the application.
// The statements are idempotent, so startup is safe with an existing database.
func InitializeSchema(db *gorm.DB) error {
	for _, statement := range schemaStatements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("initialize SQLite schema: %w", err)
		}
	}
	return nil
}

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at DATETIME,
		CONSTRAINT ck_users_active_matches_deleted_at CHECK (
			(is_active AND deleted_at IS NULL) OR (NOT is_active AND deleted_at IS NOT NULL)
		)
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS ix_users_username ON users (username)`,
	`CREATE TABLE IF NOT EXISTS accounts (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL,
		account_type TEXT NOT NULL,
		account_name TEXT NOT NULL,
		initial_balance NUMERIC NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at DATETIME,
		CONSTRAINT fk_accounts_user_id_users FOREIGN KEY (user_id)
			REFERENCES users (id) ON DELETE RESTRICT,
		CONSTRAINT uq_accounts_user_id_id UNIQUE (user_id, id),
		CONSTRAINT ck_accounts_active_matches_deleted_at CHECK (
			(is_active AND deleted_at IS NULL) OR (NOT is_active AND deleted_at IS NOT NULL)
		)
	)`,
	`CREATE INDEX IF NOT EXISTS ix_accounts_user_id ON accounts (user_id)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS uq_accounts_user_type_name_active
		ON accounts (user_id, account_type, account_name) WHERE deleted_at IS NULL`,
	`CREATE TABLE IF NOT EXISTS categories (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL,
		category_name TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at DATETIME,
		CONSTRAINT fk_categories_user_id_users FOREIGN KEY (user_id)
			REFERENCES users (id) ON DELETE RESTRICT,
		CONSTRAINT uq_categories_user_id_id UNIQUE (user_id, id),
		CONSTRAINT ck_categories_active_matches_deleted_at CHECK (
			(is_active AND deleted_at IS NULL) OR (NOT is_active AND deleted_at IS NOT NULL)
		)
	)`,
	`CREATE INDEX IF NOT EXISTS ix_categories_user_id ON categories (user_id)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_user_category_name_active
		ON categories (user_id, category_name) WHERE deleted_at IS NULL`,
	`CREATE TABLE IF NOT EXISTS transactions (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		amount NUMERIC NOT NULL,
		account_id TEXT NOT NULL,
		category_id TEXT NOT NULL,
		occurred_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at DATETIME,
		CONSTRAINT fk_transactions_user_id_users FOREIGN KEY (user_id)
			REFERENCES users (id) ON DELETE RESTRICT,
		CONSTRAINT fk_transactions_user_account_accounts FOREIGN KEY (user_id, account_id)
			REFERENCES accounts (user_id, id) ON DELETE RESTRICT,
		CONSTRAINT fk_transactions_user_category_categories FOREIGN KEY (user_id, category_id)
			REFERENCES categories (user_id, id) ON DELETE RESTRICT,
		CONSTRAINT ck_transactions_amount_positive CHECK (amount > 0),
		CONSTRAINT ck_transactions_type_valid CHECK (type IN ('income', 'expense')),
		CONSTRAINT ck_transactions_active_matches_deleted_at CHECK (
			(is_active AND deleted_at IS NULL) OR (NOT is_active AND deleted_at IS NOT NULL)
		)
	)`,
	`CREATE INDEX IF NOT EXISTS ix_transactions_user_occurred_at ON transactions (user_id, occurred_at)`,
	`CREATE INDEX IF NOT EXISTS ix_transactions_account_occurred_at ON transactions (account_id, occurred_at)`,
	`CREATE INDEX IF NOT EXISTS ix_transactions_category_occurred_at ON transactions (category_id, occurred_at)`,
}
