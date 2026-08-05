BEGIN;

CREATE TABLE users (
    id UUID NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT ck_users_active_matches_deleted_at CHECK (
        (is_active AND deleted_at IS NULL)
        OR (NOT is_active AND deleted_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX ix_users_username ON users (username);

CREATE TABLE accounts (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    account_type VARCHAR(100) NOT NULL,
    initial_balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT pk_accounts PRIMARY KEY (id),
    CONSTRAINT fk_accounts_user_id_users FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT uq_accounts_user_id_id UNIQUE (user_id, id),
    CONSTRAINT ck_accounts_active_matches_deleted_at CHECK (
        (is_active AND deleted_at IS NULL)
        OR (NOT is_active AND deleted_at IS NOT NULL)
    )
);

CREATE INDEX ix_accounts_user_id ON accounts (user_id);
CREATE UNIQUE INDEX uq_accounts_user_account_type_active
    ON accounts (user_id, account_type)
    WHERE deleted_at IS NULL;

CREATE TABLE categories (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    category_name VARCHAR(80) NOT NULL,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT pk_categories PRIMARY KEY (id),
    CONSTRAINT fk_categories_user_id_users FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT uq_categories_user_id_id_type UNIQUE (user_id, id, type),
    CONSTRAINT ck_categories_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT ck_categories_active_matches_deleted_at CHECK (
        (is_active AND deleted_at IS NULL)
        OR (NOT is_active AND deleted_at IS NOT NULL)
    )
);

CREATE INDEX ix_categories_user_id ON categories (user_id);
CREATE UNIQUE INDEX uq_categories_user_type_name_active
    ON categories (user_id, type, category_name)
    WHERE deleted_at IS NULL;

CREATE TABLE transactions (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    account_id UUID NOT NULL,
    category_id UUID NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT pk_transactions PRIMARY KEY (id),
    CONSTRAINT fk_transactions_user_id_users FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_transactions_user_account_accounts FOREIGN KEY (user_id, account_id)
        REFERENCES accounts (user_id, id) ON DELETE RESTRICT,
    CONSTRAINT fk_transactions_user_category_type_categories FOREIGN KEY (user_id, category_id, type)
        REFERENCES categories (user_id, id, type) ON DELETE RESTRICT,
    CONSTRAINT ck_transactions_amount_positive CHECK (amount > 0),
    CONSTRAINT ck_transactions_type_valid CHECK (type IN ('income', 'expense')),
    CONSTRAINT ck_transactions_active_matches_deleted_at CHECK (
        (is_active AND deleted_at IS NULL)
        OR (NOT is_active AND deleted_at IS NOT NULL)
    )
);

CREATE INDEX ix_transactions_user_occurred_at
    ON transactions (user_id, occurred_at);
CREATE INDEX ix_transactions_account_occurred_at
    ON transactions (account_id, occurred_at);
CREATE INDEX ix_transactions_category_occurred_at
    ON transactions (category_id, occurred_at);

COMMIT;
