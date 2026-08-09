BEGIN;

ALTER TABLE transactions
    DROP CONSTRAINT fk_transactions_user_category_type_categories;

ALTER TABLE categories
    DROP CONSTRAINT uq_categories_user_id_id_category_type,
    DROP CONSTRAINT ck_categories_category_type_valid,
    ADD COLUMN category_name VARCHAR(80);

-- 已有记录只保留了收入/支出方向，无法还原原始分类名称；使用可辨识且不重复的迁移名称保留记录。
UPDATE categories
SET category_name = 'legacy-' || category_type || '-' || id::text
WHERE category_name IS NULL;

ALTER TABLE categories
    ALTER COLUMN category_name SET NOT NULL,
    DROP COLUMN category_type,
    ADD CONSTRAINT uq_categories_user_id_id UNIQUE (user_id, id);

CREATE UNIQUE INDEX uq_categories_user_category_name_active
    ON categories (user_id, category_name)
    WHERE deleted_at IS NULL;

ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_user_category_categories
        FOREIGN KEY (user_id, category_id)
        REFERENCES categories (user_id, id) ON DELETE RESTRICT;

COMMIT;
