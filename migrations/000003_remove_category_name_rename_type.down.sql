BEGIN;

ALTER TABLE transactions
    DROP CONSTRAINT fk_transactions_user_category_type_categories;

ALTER TABLE categories
    DROP CONSTRAINT uq_categories_user_id_id_category_type,
    DROP CONSTRAINT ck_categories_category_type_valid;

ALTER TABLE categories
    RENAME COLUMN category_type TO type;

ALTER TABLE categories
    ADD COLUMN category_name VARCHAR(80);

-- 原 category_name 已无法恢复，使用类型和 ID 生成不重复的占位名称。
UPDATE categories
SET category_name = type || '-' || id::text
WHERE category_name IS NULL;

ALTER TABLE categories
    ALTER COLUMN category_name SET NOT NULL,
    ADD CONSTRAINT uq_categories_user_id_id_type
        UNIQUE (user_id, id, type),
    ADD CONSTRAINT ck_categories_type_valid
        CHECK (type IN ('income', 'expense'));

CREATE UNIQUE INDEX uq_categories_user_type_name_active
    ON categories (user_id, type, category_name)
    WHERE deleted_at IS NULL;

ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_user_category_type_categories
        FOREIGN KEY (user_id, category_id, type)
        REFERENCES categories (user_id, id, type) ON DELETE RESTRICT;

COMMIT;
