BEGIN;

DROP INDEX uq_accounts_user_type_name_active;

ALTER TABLE accounts
    DROP COLUMN account_name;

-- 如果同一用户在同一类型下已经创建多个账户，此约束无法重建，回滚会安全失败。
CREATE UNIQUE INDEX uq_accounts_user_account_type_active
    ON accounts (user_id, account_type)
    WHERE deleted_at IS NULL;

COMMIT;
