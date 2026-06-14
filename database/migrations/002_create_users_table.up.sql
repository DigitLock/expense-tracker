-- ============================================================================
-- Table: users
-- Purpose: User accounts with authentication and family membership
-- ============================================================================

BEGIN;

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       family_id UUID NOT NULL,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       name VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       last_login_at TIMESTAMP,
                       is_active BOOLEAN NOT NULL DEFAULT true,

                       CONSTRAINT fk_users_family
                           FOREIGN KEY (family_id)
                               REFERENCES families(id)
                               ON DELETE CASCADE,
                       CONSTRAINT users_email_format
                           CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
                       CONSTRAINT users_name_length
                           CHECK (LENGTH(name) >= 1)
);

CREATE INDEX idx_users_family
    ON users(family_id);
CREATE INDEX idx_users_email
    ON users(email)
    WHERE is_active = true;
CREATE INDEX idx_users_active
    ON users(is_active)
    WHERE is_active = true;
CREATE INDEX idx_users_family_active
    ON users(family_id, is_active)
    WHERE is_active = true;

COMMENT ON TABLE users IS
    'User accounts for family members with authentication credentials';
COMMENT ON COLUMN users.id IS
    'UUID primary key. Generated automatically.';
COMMENT ON COLUMN users.family_id IS
    'Foreign key to families table. Every user belongs to one family. ON DELETE CASCADE.';
COMMENT ON COLUMN users.email IS
    'User email address. Used for login. Must be globally unique. Format validated by constraint.';
COMMENT ON COLUMN users.password_hash IS
    'Bcrypt hashed password (12 rounds minimum recommended). NEVER store plain text passwords!';
COMMENT ON COLUMN users.name IS
    'Display name for user. Example: "Igor Kudinov", "Жена"';
COMMENT ON COLUMN users.created_at IS
    'Timestamp when user account was created. Set automatically.';
COMMENT ON COLUMN users.updated_at IS
    'Timestamp of last profile update. Updated automatically by trigger.';
COMMENT ON COLUMN users.last_login_at IS
    'Timestamp of last successful login. Updated by application on login. NULL if never logged in.';
COMMENT ON COLUMN users.is_active IS
    'Soft delete flag. true = active user, false = deleted/blocked user (data preserved for audit)';

CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
COMMENT ON TRIGGER trigger_users_updated_at ON users IS
    'Updates updated_at timestamp automatically on every UPDATE';

COMMIT;