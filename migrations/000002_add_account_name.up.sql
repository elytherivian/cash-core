BEGIN;

ALTER TABLE accounts
    ADD COLUMN account_name VARCHAR(100);

-- 兼容已有账户：升级时使用账户类型回填名称，确保新增的非空约束可以安全建立。
UPDATE accounts
SET account_name = account_type
WHERE account_name IS NULL;

ALTER TABLE accounts
    ALTER COLUMN account_name SET NOT NULL;

DROP INDEX uq_accounts_user_account_type_active;

CREATE UNIQUE INDEX uq_accounts_user_type_name_active
    ON accounts (user_id, account_type, account_name)
    WHERE deleted_at IS NULL;

COMMIT;
