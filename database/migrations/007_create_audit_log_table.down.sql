BEGIN;

DROP VIEW IF EXISTS v_recent_audit_log;

DROP TRIGGER IF EXISTS trigger_audit_users ON users;
DROP TRIGGER IF EXISTS trigger_audit_accounts ON accounts;
DROP TRIGGER IF EXISTS trigger_audit_categories ON categories;
DROP TRIGGER IF EXISTS trigger_audit_transactions ON transactions;

DROP FUNCTION IF EXISTS audit_trigger() CASCADE;

DROP TABLE IF EXISTS audit_log CASCADE;

COMMIT;