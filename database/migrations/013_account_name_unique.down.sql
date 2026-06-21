-- ============================================================================
-- Rollback: 013 account name uniqueness
-- ============================================================================

BEGIN;

DROP INDEX IF EXISTS idx_accounts_unique_active_name;

COMMIT;
