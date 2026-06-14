-- ============================================================================
-- Rollback: 011 add user role
-- ============================================================================

BEGIN;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users DROP COLUMN IF EXISTS role;

COMMIT;
