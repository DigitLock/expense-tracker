-- ============================================================================
-- Migration: 011 add user role
-- Purpose: Add role column to users for ownership/membership distinction.
--          Self-registration creates a family owner; later members default to 'member'.
-- ============================================================================

BEGIN;

ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'member';

-- Backfill existing users: each pre-existing user owns their own family.
UPDATE users SET role = 'owner';

ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('owner', 'member'));

COMMENT ON COLUMN users.role IS
    'User role within the family. owner = created the family (full control), member = invited user.';

COMMIT;
