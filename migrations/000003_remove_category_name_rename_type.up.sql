BEGIN;

ALTER TABLE transactions
    DROP CONSTRAINT fk_transactions_user_category_type_categories;

DROP INDEX uq_categories_user_type_name_active;

ALTER TABLE categories
    DROP CONSTRAINT uq_categories_user_id_id_type,
    DROP CONSTRAINT ck_categories_type_valid,
    DROP COLUMN category_name;

ALTER TABLE categories
    RENAME COLUMN type TO category_type;

ALTER TABLE categories
    ADD CONSTRAINT uq_categories_user_id_id_category_type
        UNIQUE (user_id, id, category_type),
    ADD CONSTRAINT ck_categories_category_type_valid
        CHECK (category_type IN ('income', 'expense'));

ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_user_category_type_categories
        FOREIGN KEY (user_id, category_id, type)
        REFERENCES categories (user_id, id, category_type) ON DELETE RESTRICT;

COMMIT;
