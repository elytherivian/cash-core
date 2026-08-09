BEGIN;

ALTER TABLE transactions
    DROP CONSTRAINT fk_transactions_user_category_categories;

DROP INDEX uq_categories_user_category_name_active;

ALTER TABLE categories
    DROP CONSTRAINT uq_categories_user_id_id,
    ADD COLUMN category_type VARCHAR(20);

-- 新模型允许同一分类用于任意流水方向；回退时按已有流水推断一个方向，无法推断时使用 expense。
UPDATE categories AS category
SET category_type = COALESCE((
    SELECT entry.type
    FROM transactions AS entry
    WHERE entry.user_id = category.user_id
      AND entry.category_id = category.id
    ORDER BY entry.occurred_at DESC
    LIMIT 1
), 'expense');

ALTER TABLE categories
    ALTER COLUMN category_type SET NOT NULL,
    ADD CONSTRAINT uq_categories_user_id_id_category_type UNIQUE (user_id, id, category_type),
    ADD CONSTRAINT ck_categories_category_type_valid CHECK (category_type IN ('income', 'expense')),
    DROP COLUMN category_name;

-- 若一个新分类同时用于收入和支出，旧模型无法完整表达；保留现有流水并让后续写入受约束。
ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_user_category_type_categories
        FOREIGN KEY (user_id, category_id, type)
        REFERENCES categories (user_id, id, category_type) ON DELETE RESTRICT NOT VALID;

COMMIT;
