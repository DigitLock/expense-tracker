-- ============================================================================
-- Table: audit_log
-- Purpose: Complete audit trail of all data changes
-- ============================================================================

BEGIN;

CREATE TABLE audit_log (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           family_id UUID NOT NULL,
                           user_id UUID,
                           table_name VARCHAR(100) NOT NULL,
                           record_id UUID NOT NULL,
                           action VARCHAR(20) NOT NULL,
                           before_data JSONB,
                           after_data JSONB,
                           created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

                           CONSTRAINT fk_audit_log_family
                               FOREIGN KEY (family_id)
                                   REFERENCES families(id)
                                   ON DELETE CASCADE,
                           CONSTRAINT fk_audit_log_user
                               FOREIGN KEY (user_id)
                                   REFERENCES users(id)
                                   ON DELETE SET NULL,
                           CONSTRAINT audit_log_action_check
                               CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
                           CONSTRAINT audit_log_data_consistency
                               CHECK (
                                   (action = 'INSERT' AND before_data IS NULL AND after_data IS NOT NULL) OR
                                   (action = 'UPDATE' AND before_data IS NOT NULL AND after_data IS NOT NULL) OR
                                   (action = 'DELETE' AND before_data IS NOT NULL AND after_data IS NULL)
                                   )
);

CREATE INDEX idx_audit_log_family
    ON audit_log(family_id);
CREATE INDEX idx_audit_log_user
    ON audit_log(user_id);
CREATE INDEX idx_audit_log_table
    ON audit_log(table_name);
CREATE INDEX idx_audit_log_created
    ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_record
    ON audit_log(table_name, record_id, created_at DESC);
CREATE INDEX idx_audit_log_before_data
    ON audit_log USING GIN (before_data);
CREATE INDEX idx_audit_log_after_data
    ON audit_log USING GIN (after_data);

COMMENT ON TABLE audit_log IS
    'Complete audit trail of all data changes. Automatically populated by triggers.';
COMMENT ON COLUMN audit_log.user_id IS
    'User who made the change. Retrieved from session variable app.current_user_id. NULL for system changes or if user was deleted.';
COMMENT ON COLUMN audit_log.table_name IS
    'Name of table that was changed: users, accounts, categories, transactions';
COMMENT ON COLUMN audit_log.record_id IS
    'ID of the record that was changed';
COMMENT ON COLUMN audit_log.action IS
    'Type of operation: INSERT, UPDATE, DELETE';
COMMENT ON COLUMN audit_log.before_data IS
    'JSONB snapshot of record BEFORE change. NULL for INSERT.';
COMMENT ON COLUMN audit_log.after_data IS
    'JSONB snapshot of record AFTER change. NULL for DELETE.';

CREATE OR REPLACE FUNCTION audit_trigger()
    RETURNS TRIGGER AS $$
DECLARE
    v_user_id UUID;
    v_family_id UUID;
BEGIN
    BEGIN
        v_user_id := current_setting('app.current_user_id', true)::UUID;
    EXCEPTION
        WHEN OTHERS THEN
            v_user_id := NULL;
    END;

    v_family_id := COALESCE(NEW.family_id, OLD.family_id);

    INSERT INTO audit_log (
        family_id,
        user_id,
        table_name,
        record_id,
        action,
        before_data,
        after_data,
        created_at
    ) VALUES (
                 v_family_id,
                 v_user_id,
                 TG_TABLE_NAME,
                 COALESCE(NEW.id, OLD.id),
                 TG_OP,
                 CASE
                     WHEN TG_OP IN ('UPDATE', 'DELETE') THEN row_to_json(OLD)::JSONB
                     ELSE NULL
                     END,
                 CASE
                     WHEN TG_OP IN ('INSERT', 'UPDATE') THEN row_to_json(NEW)::JSONB
                     ELSE NULL
                     END,
                 CURRENT_TIMESTAMP
             );

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;
COMMENT ON FUNCTION audit_trigger() IS
    'Automatically logs INSERT/UPDATE/DELETE operations to audit_log table. Triggered on users, accounts, categories, transactions.';

CREATE TRIGGER trigger_audit_users
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW
EXECUTE FUNCTION audit_trigger();
CREATE TRIGGER trigger_audit_accounts
    AFTER INSERT OR UPDATE OR DELETE ON accounts
    FOR EACH ROW
EXECUTE FUNCTION audit_trigger();
CREATE TRIGGER trigger_audit_categories
    AFTER INSERT OR UPDATE OR DELETE ON categories
    FOR EACH ROW
EXECUTE FUNCTION audit_trigger();
CREATE TRIGGER trigger_audit_transactions
    AFTER INSERT OR UPDATE OR DELETE ON transactions
    FOR EACH ROW
EXECUTE FUNCTION audit_trigger();

COMMENT ON TRIGGER trigger_audit_users ON users IS
    'Logs all changes to users table';
COMMENT ON TRIGGER trigger_audit_accounts ON accounts IS
    'Logs all changes to accounts table';
COMMENT ON TRIGGER trigger_audit_categories ON categories IS
    'Logs all changes to categories table';
COMMENT ON TRIGGER trigger_audit_transactions ON transactions IS
    'Logs all changes to transactions table';

CREATE OR REPLACE VIEW v_recent_audit_log AS
SELECT
    a.id,
    f.name AS family_name,
    u.name AS user_name,
    u.email AS user_email,
    a.table_name,
    a.action,
    a.record_id,
    a.created_at,
    a.before_data,
    a.after_data
FROM audit_log a
         LEFT JOIN families f ON a.family_id = f.id
         LEFT JOIN users u ON a.user_id = u.id
ORDER BY a.created_at DESC;
COMMENT ON VIEW v_recent_audit_log IS
    'Convenient view of audit log with joined family and user names';

COMMIT;