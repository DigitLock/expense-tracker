-- ============================================================================
-- Table: families
-- Purpose: Root entity for family/household financial data isolation
-- ============================================================================
BEGIN;
CREATE TABLE families (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name VARCHAR(255) NOT NULL,
                          base_currency VARCHAR(3) NOT NULL DEFAULT 'RSD',
                          created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          is_active BOOLEAN NOT NULL DEFAULT true,

                          CONSTRAINT families_base_currency_check
                              CHECK (base_currency = 'RSD'),
                          CONSTRAINT families_name_length
                              CHECK (LENGTH(name) >= 1)
);

CREATE INDEX idx_families_active
    ON families(is_active)
    WHERE is_active = true;

COMMENT ON TABLE families IS
    'Family/household entities for financial data isolation. Root table for multi-tenancy.';
COMMENT ON COLUMN families.id IS
    'UUID primary key. Generated automatically. Used as family_id in all related tables.';
COMMENT ON COLUMN families.name IS
    'Family display name. Example: "Kudinov Family", "Test Family"';
COMMENT ON COLUMN families.base_currency IS
    'Base currency for financial reporting. MVP: Only RSD supported. Post-MVP: User-configurable (EUR, USD, etc.)';
COMMENT ON COLUMN families.created_at IS
    'Timestamp when family was created. Set automatically.';
COMMENT ON COLUMN families.updated_at IS
    'Timestamp of last update. Updated automatically by trigger.';
COMMENT ON COLUMN families.is_active IS
    'Soft delete flag. true = active family, false = deleted family (data preserved for history)';

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS
    'Automatically updates updated_at column to current timestamp on UPDATE';

CREATE TRIGGER trigger_families_updated_at
    BEFORE UPDATE ON families
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TRIGGER trigger_families_updated_at ON families IS
    'Updates updated_at timestamp automatically on every UPDATE';

COMMIT;