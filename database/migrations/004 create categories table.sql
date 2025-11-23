-- ============================================================================
-- Table: categories
-- Purpose: Transaction categories (income and expense) with hierarchical support
-- ============================================================================

BEGIN;

CREATE TABLE categories (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            family_id UUID NOT NULL,
                            name VARCHAR(255) NOT NULL,
                            type VARCHAR(50) NOT NULL,
                            parent_id UUID,
                            description TEXT,
                            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            is_active BOOLEAN NOT NULL DEFAULT true,

                            CONSTRAINT fk_categories_family
                                FOREIGN KEY (family_id)
                                    REFERENCES families(id)
                                    ON DELETE CASCADE,
                            CONSTRAINT fk_categories_parent
                                FOREIGN KEY (parent_id)
                                    REFERENCES categories(id)
                                    ON DELETE SET NULL,
                            CONSTRAINT categories_type_check
                                CHECK (type IN ('income', 'expense')),
                            CONSTRAINT categories_name_length
                                CHECK (LENGTH(name) >= 1),
                            CONSTRAINT categories_unique_name
                                UNIQUE (family_id, name, type)
);

CREATE INDEX idx_categories_family
    ON categories(family_id);
CREATE INDEX idx_categories_type
    ON categories(type);
CREATE INDEX idx_categories_parent
    ON categories(parent_id);
CREATE INDEX idx_categories_family_active
    ON categories(family_id, is_active)
    WHERE is_active = true;
CREATE INDEX idx_categories_family_type
    ON categories(family_id, type)
    WHERE is_active = true;

COMMENT ON TABLE categories IS
    'Income and expense categories for transaction classification. Supports hierarchical structure (parent-child relationships).';
COMMENT ON COLUMN categories.id IS
    'UUID primary key. Generated automatically.';
COMMENT ON COLUMN categories.family_id IS
    'Foreign key to families. ON DELETE CASCADE.';
COMMENT ON COLUMN categories.name IS
    'Category name. Examples: "Food", "Groceries", "Salary". Must be unique per family and type.';
COMMENT ON COLUMN categories.type IS
    'Category type: income (for income transactions) or expense (for expense transactions)';
COMMENT ON COLUMN categories.parent_id IS
    'Parent category ID for hierarchical structure. NULL = top-level category. Example: "Food" (parent) -> "Groceries" (child with parent_id = Food.id)';
COMMENT ON COLUMN categories.description IS
    'Optional category description';
COMMENT ON COLUMN categories.created_at IS
    'Timestamp when category was created. Set automatically.';
COMMENT ON COLUMN categories.updated_at IS
    'Timestamp of last update. Updated automatically by trigger.';
COMMENT ON COLUMN categories.is_active IS
    'Soft delete flag. true = active category, false = deleted category (transactions preserved)';

CREATE TRIGGER trigger_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
COMMENT ON TRIGGER trigger_categories_updated_at ON categories IS
    'Updates updated_at timestamp automatically on every UPDATE';

COMMIT;