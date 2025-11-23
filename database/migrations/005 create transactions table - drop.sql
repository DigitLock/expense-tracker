BEGIN;

DROP TRIGGER IF EXISTS trigger_transactions_update_balance ON transactions;
DROP TRIGGER IF EXISTS trigger_transactions_updated_at ON transactions;

DROP FUNCTION IF EXISTS update_account_balance() CASCADE;

DROP TABLE IF EXISTS transactions CASCADE;

COMMIT;