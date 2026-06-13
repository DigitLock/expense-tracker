BEGIN;

DROP TRIGGER IF EXISTS trigger_accounts_updated_at ON accounts;

DROP TABLE IF EXISTS accounts CASCADE;

COMMIT;