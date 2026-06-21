-- ============================================================================
-- Migration: 013 account name uniqueness
-- Purpose: Enforce unique account names per family (ALREADY_EXISTS rule, R-uniq).
--          Mirrors the categories active-name uniqueness form: case-insensitive
--          (LOWER(name)), partial — only for active accounts, so soft-deleted
--          names can be reused.
-- ============================================================================

BEGIN;

CREATE UNIQUE INDEX idx_accounts_unique_active_name
    ON accounts (
                 family_id,
                 LOWER(name)  -- case-insensitive, matches idx_categories_unique_active
        )
    WHERE is_active = true;

COMMENT ON INDEX idx_accounts_unique_active_name IS
    'Ensures unique account names within a family, only for active accounts. Allows reusing names from soft-deleted accounts.';

COMMIT;
